package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	cloudservice "github.com/kube-stack/multicloud_service/src/service"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"os"
)

type Executor struct {
	Service *cloudservice.MultiCloudService
	Log     logr.Logger
	Kind    string
}

func NewExecutor(kind string, logger logr.Logger, secretInfo map[string][]byte) (*Executor, error) {
	e := &Executor{
		Service: nil,
		Log:     logger,
		Kind:    kind,
	}
	params := make(map[string]string)
	for key, value := range secretInfo {
		params[key] = string(value)
	}
	mcs, err := cloudservice.NewMultiCloudService(params)
	if err != nil {
		logger.Error(err, "InitServiceBySecret err")
		return nil, err
	}
	e.Service = mcs
	return e, nil
}

// 传入inst.Spec，从中获取对应类型的元数据
func (e *Executor) GetMetaFromSpec(specInfo []byte) (map[string]string, error) {
	params := make(map[string]string)
	switch e.Kind {
	case ALIYUN_ECS_INSTANCE:
		params["RegionId"] = gjson.GetBytes(specInfo, "regionId").String()
		params["InstanceId"] = gjson.GetBytes(specInfo, "instanceId").String()
	default:
		params["RegionId"] = gjson.GetBytes(specInfo, "regionId").String()
		params["SnapshotId"] = gjson.GetBytes(specInfo, "snapshotId").String()
	}
	return params, nil
}

// 从describe中获取元数据
func (e *Executor) GetMetaFromDomain(domainInfo []byte) map[string]string {
	params := make(map[string]string)
	switch e.Kind {
	case ALIYUN_ECS_INSTANCE:
		params["regionId"] = gjson.GetBytes(domainInfo, "RegionId").String()
		params["instanceId"] = gjson.GetBytes(domainInfo, "InstanceId").String()
	case ALIYUN_ECS_SNAPSHOT:
		params["regionId"] = gjson.GetBytes(domainInfo, "RegionId").String()
		params["snapshotId"] = gjson.GetBytes(domainInfo, "SnapshotId").String()
	default:

	}
	return params
}

func (e *Executor) getCrdInfoFromCloud(specInfo []byte) ([]byte, error) {
	params, err := e.GetMetaFromSpec(specInfo)
	if err != nil {
		return nil, err
	}
	switch e.Kind {
	case ALIYUN_ECS_INSTANCE:
		initBytes, err := os.ReadFile(InitJsonFile[e.Kind])
		if err != nil {
			e.Log.Error(err, "Open Crd Template file error: ")
			return nil, err
		}
		initBytes, err = sjson.SetBytes(initBytes, "DescribeInstances.RegionId", params["RegionId"])
		if err != nil {
			e.Log.Error(err, "SetJson error")
			return nil, err
		}
		initBytes, err = sjson.SetBytes(initBytes, "DescribeInstances.InstanceIds", "[\""+params["InstanceId"]+"\"]")
		if err != nil {
			e.Log.Error(err, "SetJson error")
			return nil, err
		}
		resp, err := e.ServiceCall(initBytes)
		if err != nil {
			e.Log.Error(err, "Call Cloud API Error")
			return nil, errors.Wrap(err, "CallCloudAPI:")
		}
		return resp, nil
	case ALIYUN_ECS_SNAPSHOT:
		initBytes, err := os.ReadFile(InitJsonFile[e.Kind])
		if err != nil {
			e.Log.Error(err, "Open Crd Template file error: ")
			return nil, err
		}
		initBytes, err = sjson.SetBytes(initBytes, "DescribeSnapshots.RegionId", params["RegionId"])
		if err != nil {
			e.Log.Error(err, "SetJson error")
			return nil, err
		}
		initBytes, err = sjson.SetBytes(initBytes, "DescribeSnapshotIds.SnapshotIds", "[\""+params["SnapshotId"]+"\"]")
		if err != nil {
			e.Log.Error(err, "SetJson error")
			return nil, err
		}
		resp, err := e.ServiceCall(initBytes)
		if err != nil {
			e.Log.Error(err, "Call Cloud API Error")
			return nil, errors.Wrap(err, "CallCloudAPI:")
		}
		return resp, nil
	default:
		err := fmt.Errorf("getCrdInfoFromCloud: unsupport Kind %s\n", e.Kind)
		e.Log.Error(err, "getCrdInfoFromCloud: ")
		return nil, err
	}
	//return nil, nil
}

func (e *Executor) ServiceCall(requestInfo []byte) ([]byte, error) {
	requestMap, err := jsonByte2Map(requestInfo)
	if err != nil {
		e.Log.Error(err, "Marshal requestinfo to map error:")
		return nil, err
	}
	//only on request
	for APIName, APIParameters := range requestMap {
		jsonBytes, err := json.Marshal(APIParameters)
		if err != nil {
			e.Log.Error(err, "Marshal parameters to json error")
			return nil, err
		}
		resp, err := e.Service.CallCloudAPI(APIName, jsonBytes)
		if err != nil {
			e.Log.Error(err, "Call Cloud API Error")
			return nil, err
		}
		return resp, err
	}
	return nil, nil
}

func (e *Executor) UpdateCrdDomain(crdJsonBytes []byte) ([]byte, error) {
	specInfo := []byte(gjson.GetBytes(crdJsonBytes, "spec").String())
	resp, err := e.getCrdInfoFromCloud(specInfo)
	if err != nil {
		return nil, err
	}
	domain := gjson.GetBytes(resp, DomainJsonPath[e.Kind]).Raw
	newCrd, err := sjson.SetRawBytes(crdJsonBytes, "spec.domain", []byte(domain))
	if err != nil {
		e.Log.Error(err, "setting domain err")
		return nil, err
	}
	return newCrd, nil
}
