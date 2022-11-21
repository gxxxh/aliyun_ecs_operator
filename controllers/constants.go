package controllers

import (
	aliyunecsv1 "cloudOperator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ALIYUN_ECS_INSTANCE = "VMInstance"
	ALIYUN_ECS_SNAPSHOT = "ECSSnapshot"
)

var CRDS = []string{
	ALIYUN_ECS_INSTANCE,
	ALIYUN_ECS_SNAPSHOT,
}

var EmptyCrdObjects = map[string]client.Object{
	ALIYUN_ECS_INSTANCE: &aliyunecsv1.VMInstance{},
	ALIYUN_ECS_SNAPSHOT: &aliyunecsv1.ECSSnapshot{},
}

var DomainJsonPath = map[string]string{
	ALIYUN_ECS_SNAPSHOT: "Snapshots.Snapshot.0",
	ALIYUN_ECS_INSTANCE: "Instances.Instance.0",
}

var InitJsonFile = map[string]string{
	ALIYUN_ECS_INSTANCE: "/root/go/src/aliyun_ecs_operator/config/init/ecs/describe_instance.json",
	ALIYUN_ECS_SNAPSHOT: "/root/go/src/aliyun_ecs_operator/config/init/ecs/describe_instance.json",
}
