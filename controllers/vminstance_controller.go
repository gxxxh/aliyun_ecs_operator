/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	aliyunecsv1 "cloudOperator/api/v1"
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VMInstanceReconciler reconciles a VMInstance object
type VMInstanceReconciler struct {
	client.Client
	Kind     string
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=aliyun.ecs.doslab.io,resources=vminstances,verbs=get;list;watch;create;update;patch;delete.json
//+kubebuilder:rbac:groups=aliyun.ecs.doslab.io,resources=vminstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aliyun.ecs.doslab.io,resources=vminstances/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VMInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile

func (r *VMInstanceReconciler) GetCrdJsonFromK8s(ctx context.Context, req ctrl.Request) ([]byte, error) {
	crd := reflect.New(reflect.TypeOf(EmptyCrdObjects[r.Kind]).Elem()).Interface().(client.Object)
	if err := r.Get(ctx, req.NamespacedName, crd); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("didn't find the object", "Req", req)
			return nil, nil
		}
		r.Log.Error(err, "Get crd info error")
		return nil, err
	}
	jsonBytes, err := json.Marshal(crd)
	if err != nil {
		return nil, errors.Wrap(err, "GetCrdJsonFromK8s.JsonMarshal:")
	}
	return jsonBytes, nil
}

func (r *VMInstanceReconciler) GetCrdOjbect(jsonBytes []byte) (client.Object, error) {
	crd := reflect.New(reflect.TypeOf(EmptyCrdObjects[r.Kind]).Elem()).Interface().(client.Object)
	err := json.Unmarshal(jsonBytes, crd)
	if err != nil {
		r.Log.Error(err, "Can't get runtime.Object from json")
		return nil, err
	}
	return crd, nil
}

func (r *VMInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log
	// TODO(user): your logic here
	var (
		err          error
		crdJsonBytes []byte
	)
	//get crd json
	crdJsonBytes, err = r.GetCrdJsonFromK8s(ctx, req)
	if err != nil || crdJsonBytes == nil {
		return ctrl.Result{}, err
	}

	//get crd object, this object is using to update event
	crdObject, err := r.GetCrdOjbect(crdJsonBytes)
	if err != nil {
		return ctrl.Result{}, err
	}

	//get secrets and init Service
	//todo ,using get interface
	secret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: gjson.GetBytes(crdJsonBytes, "spec.secretRef.namespace").String(),
		Name:      gjson.GetBytes(crdJsonBytes, "spec.secretRef.name").String(),
	}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "Not found secret to the cloud")
			return ctrl.Result{}, err
		}
		logger.Error(err, "Get Secret Error")
		return ctrl.Result{}, err
	}

	//init a executor to execute request
	executor, err := NewExecutor(r.Kind, r.Log, secret.Data)
	if err != nil {
		logger.Error(err, "Init executor error")
		return ctrl.Result{}, err
	}
	oldLifeCycle := gjson.GetBytes(crdJsonBytes, "spec.lifeCycle").String()
	oldDomain := gjson.GetBytes(crdJsonBytes, "spec.domain").String()
	// 无需处理
	if oldLifeCycle == "" || oldLifeCycle == "{}" {
		if oldDomain == "" || oldDomain == "{}" {
			//加入k8s管理，初始为空

			crdJsonBytes, err = executor.UpdateCrdDomain(crdJsonBytes)
			if err != nil {
				return ctrl.Result{}, err
			}
			if err := r.UpdateCrd(ctx, crdJsonBytes); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("Add Instance to kubernetes cluster, ", "Instance:", types.NamespacedName{
				Namespace: req.Namespace,
				Name:      req.Name,
			}.String())
			r.Recorder.Event(crdObject, corev1.EventTypeNormal, "Add Instance to kubernetes cluster", r.Kind)
		}
		logger.Info("Instance lifecycle is empty", "Instance:", types.NamespacedName{
			Namespace: req.Namespace,
			Name:      req.Name,
		}.String())
		return ctrl.Result{}, nil
	}

	//todo 执行结果写入到event中
	//todo 测试cloudAPI的err
	resp, err := executor.ServiceCall([]byte(oldLifeCycle))
	if err != nil {
		logger.Error(err, "call Cloud API Error")
		r.Recorder.Event(crdObject, "Error", "Call Cloud API Error", err.Error())
		return ctrl.Result{}, err
	}
	// succeed event
	r.Recorder.Eventf(crdObject, corev1.EventTypeNormal, "Call Cloud API Succeed", "RequestInfo: %s, ResponseInfo: %s", string(oldLifeCycle), string(resp))
	// update spec to nil
	crdJsonBytes, err = sjson.SetBytes(crdJsonBytes, "spec.lifeCycle", nil)
	if err != nil {
		logger.Error(err, "Set Spec to nil error")
	}
	// update domain to new info
	crdJsonBytes, err = executor.UpdateCrdDomain(crdJsonBytes)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := r.UpdateCrd(ctx, crdJsonBytes); err != nil {
		return ctrl.Result{}, err
	}
	//todo 删除资源时，调用删除
	return ctrl.Result{}, nil
}

func (r *VMInstanceReconciler) UpdateCrd(ctx context.Context, crdJsonBytes []byte) error {
	crdObject, err := r.GetCrdOjbect(crdJsonBytes)
	if err != nil {
		r.Log.Error(err, "Can't get runtime.Object from json")
		return err
	}
	if err := r.Update(ctx, crdObject); err != nil {
		r.Log.Error(err, "Update Crd Info error")
		r.Recorder.Event(crdObject, corev1.EventTypeWarning, "Update Resource's Domain error", err.Error())
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VMInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aliyunecsv1.VMInstance{}).
		Complete(r)
}

func (r *VMInstanceReconciler) SetUpWithManagerCrd(mgr ctrl.Manager, emptyObject client.Object) error {
	return ctrl.NewControllerManagedBy(mgr).For(emptyObject).Complete(r)
}
