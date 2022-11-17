package controllers

var (
	ALIYUN_INSTANCE = "VMInstance"
)

var DomainJsonPath = map[string]string{
	ALIYUN_INSTANCE: "Instances.Instance.0",
}

var InitJsonFile = map[string]string{
	ALIYUN_INSTANCE: "/root/go/src/aliyun_ecs_operator/config/init/ecs/describe_instance.json",
}
