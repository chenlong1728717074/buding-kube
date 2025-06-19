package dto

// ReplicaSetCreateDTO ReplicaSet创建DTO
type ReplicaSetCreateDTO struct {
	WorkloadBaseDTO
	Replicas int32 `json:"replicas" binding:"required"`
}

// ReplicaSetBaseDTO ReplicaSet基础DTO
type ReplicaSetBaseDTO struct {
	WorkloadBaseDTO
}

// ReplicaSetPageQueryBaseDTO ReplicaSet分页查询DTO
type ReplicaSetPageQueryBaseDTO struct {
	PageQueryDTO
	ReplicaSetBaseDTO
}

// ReplicaSetApplyDTO ReplicaSet应用DTO
type ReplicaSetApplyDTO struct {
	ReplicaSetBaseDTO
	Yaml string `json:"yaml" binding:"required"`
}
