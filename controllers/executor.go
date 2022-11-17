package controllers

import (
	"encoding/json"
	"fmt"
	cloudservice "github.com/kube-stack/multicloud_service/src/service"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"os"
)

type Executor struct {
	Service *cloudservice.MultiCloudService
	Kind    string
}

func NewExecutor(kind string, data map[string][]byte) (*Executor, error) {
	e := &Executor{
		Service: nil,
		Kind:    kind,
	}
	err := e.InitServiceBySecret(data)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Executor) InitServiceBySecret(data map[string][]byte) (err error) {
	// init Service with secret
	params := make(map[string]string)
	for key, value := range data {
		params[key] = string(value)
	}
	e.Service, err = cloudservice.NewMultiCloudService(params)
	if err != nil {
		return errors.Wrap(err, "InitServiceBySecret: ")
	}
	return err
}

// 传入inst.Spec，从中获取对应类型的元数据
func (e *Executor) initParams(object interface{}) (map[string]string, error) {
	jsonBytes, err := json.Marshal(object)
	if err != nil {
		return nil, errors.Wrap(err, "initParams:")
	}
	params := make(map[string]string)
	switch e.Kind {
	case ALIYUN_INSTANCE:
		params["RegionId"] = gjson.GetBytes(jsonBytes, "regionId").Str
		params["InstanceId"] = gjson.GetBytes(jsonBytes, "instanceId").Str
	}
	return params, nil
}

func (e *Executor) GetCRDInfo(object interface{}) ([]byte, error) {
	params, err := e.initParams(object)
	if err != nil {
		return nil, err
	}
	switch e.Kind {
	case ALIYUN_INSTANCE:
		initBytes, err := os.ReadFile(InitJsonFile[e.Kind])
		if err != nil {
			return nil, errors.Wrap(err, "GetCRDInfo:")
		}
		sjson.SetBytes(initBytes, "DescribeInstances.RegionId", params["RegionId"])
		sjson.SetBytes(initBytes, "DescribeInstances.InstanceIds", "["+params["InstanceId"]+"]")
		return e.ServiceCall(initBytes)
	default:
		return nil, fmt.Errorf("GetCRDInfo: unsupport Kind %s\n", e.Kind)
	}
	//return nil, nil
}

func (e *Executor) ServiceCall(requestInfo []byte) ([]byte, error) {
	requestMap, err := jsonByte2Map(requestInfo)
	if err != nil {
		return nil, errors.Wrap(err, "ServiceCall: ")
	}
	//only on request
	for APIName, APIParameters := range requestMap {
		jsonBytes, err := json.Marshal(APIParameters)
		if err != nil {
			return nil, err
		}
		resp, err := e.Service.CallCloudAPI(APIName, jsonBytes)
		if err != nil {
			return nil, errors.Wrap(err, "CallCloudAPI:")
		}
		return resp, err
	}
	return nil, nil
}

func (e *Executor) GetDomain(object interface{}) ([]byte, error) {
	resp, err := e.GetCRDInfo(object)
	if err != nil {
		return nil, errors.Wrap(err, "GetDomain")
	}

	return []byte(gjson.GetBytes(resp, DomainJsonPath[e.Kind]).Raw), nil
}
