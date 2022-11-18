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
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VMInstanceReconciler reconciles a VMInstance object
type VMInstanceReconciler struct {
	client.Client
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
func (r *VMInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log
	// TODO(user): your logic here
	//获取对象
	var (
		err error
	)
	inst := &aliyunecsv1.VMInstance{}
	if err = r.Get(ctx, req.NamespacedName, inst); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Get Virtual machine error")
		return ctrl.Result{}, err
	}
	//get secrets and init Service
	secret := &corev1.Secret{}
	if err = r.Get(ctx, client.ObjectKey{
		Namespace: inst.Spec.SecretRef.Namespace,
		Name:      inst.Spec.SecretRef.Name,
	}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "Not found secret to the cloud")
			return ctrl.Result{}, err
		}
		logger.Error(err, "Get Secret Error")
		return ctrl.Result{}, err
	}

	//init a executor to execute request
	executor, err := NewExecutor(inst.Kind, secret.Data)
	if err != nil {
		logger.Error(err, "Init executor error")
		return ctrl.Result{}, err
	}

	// 无需处理
	if (inst.Spec.LifeCycle.Raw == nil) || (inst.Spec.LifeCycle.Raw != nil && len(inst.Spec.LifeCycle.Raw) == 0) {
		if (inst.Spec.Domain.Raw == nil) || (inst.Spec.Domain.Raw != nil && len(inst.Spec.Domain.Raw) == 0) {
			//init
			domain, err := executor.GetDomain(inst.Spec)
			if err != nil {
				logger.Error(err, fmt.Sprintf("Get Domain info for %s error\n", inst.Kind))
			}
			inst.Spec.Domain.Raw = domain
			logger.Info("Add Instance to kubernetes cluster, ", "Instance:", types.NamespacedName{
				Namespace: inst.Namespace,
				Name:      inst.Name,
			}.String())
			r.Recorder.Event(inst, corev1.EventTypeNormal, "Add Instance to kubernetes cluster", inst.Kind)
		}
		logger.Info("Instance lifecycle is empty", "Instance:", types.NamespacedName{
			Namespace: inst.Namespace,
			Name:      inst.Name,
		}.String())
		return ctrl.Result{}, nil
	}

	//todo 执行结果写入到event中
	//todo 测试cloudAPI的err
	resp, err := executor.ServiceCall(inst.Spec.LifeCycle.Raw)
	if err != nil {
		logger.Error(err, "call Cloud API Error")
		r.Recorder.Event(inst, "Error", "Call Cloud API Error", err.Error())
		return ctrl.Result{}, err
	}
	// succeed event
	r.Recorder.Eventf(inst, corev1.EventTypeNormal, "Call Cloud API Succeed", "RequestInfo: %s, ResponseInfo: %s", string(inst.Spec.LifeCycle.Raw), string(resp))
	// update
	inst.Spec.LifeCycle.Raw = nil
	domain, err := executor.GetDomain(inst.Spec)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Get Domain info for %s error\n", inst.Kind))
		r.Recorder.Event(inst, "Error", "Get Cloud Resource's metadata error", err.Error())
		return ctrl.Result{}, errors.Wrap(err, "Get Domain")
	}
	inst.Spec.Domain.Raw = domain
	// write back
	if err := r.Update(ctx, inst); err != nil {
		r.Recorder.Event(inst, "Error", "Update Resource's Domain error", err.Error())
		return ctrl.Result{}, errors.Wrap(err, "Update")
	}
	//todo 删除资源时，调用删除
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VMInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aliyunecsv1.VMInstance{}).
		Complete(r)
}
