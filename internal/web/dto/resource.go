package dto

// ResourceRefDTO 资源类型: Kind 为小写内置类型(如 deployment / networkpolicy);
// Kind 为 custom 时用 Group / Version / Resource / KindName / Namespaced 指定自定义资源(CRD 的实例)
type ResourceRefDTO struct {
	Kind       string `json:"kind" form:"kind" binding:"required"`
	Group      string `json:"group" form:"group"`
	Version    string `json:"version" form:"version"`
	Resource   string `json:"resource" form:"resource"`
	KindName   string `json:"kindName" form:"kindName"`
	Namespaced bool   `json:"namespaced" form:"namespaced"`
}

// ResourceTargetDTO 指定单个资源(查看 YAML / 对象 / 删除)
type ResourceTargetDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	ResourceRefDTO
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name" binding:"required"`
}

// ResourceYamlApplyDTO 保存 YAML 的参数.
// 编辑(PUT /resource/yaml)时 Name 必填, 且不允许在 YAML 里修改名称和命名空间;
// YAML 添加(POST /resource/apply)时 Name 可为空, 资源不存在则创建
type ResourceYamlApplyDTO struct {
	ClusterId string `json:"clusterId" binding:"required"`
	ResourceRefDTO
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Yaml      string `json:"yaml" binding:"required"`
	// CreateOnly 只创建, 同名资源已存在时报错
	CreateOnly bool `json:"createOnly"`
}

// ResourceTableQueryDTO 通用列表
type ResourceTableQueryDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	ResourceRefDTO
	PageQueryDTO
	Namespace     string `json:"namespace" form:"namespace"`
	LabelSelector string `json:"labelSelector" form:"labelSelector"`
}

// ResourceObjectsQueryDTO 列出完整对象
type ResourceObjectsQueryDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	ResourceRefDTO
	Namespace     string `json:"namespace" form:"namespace"`
	LabelSelector string `json:"labelSelector" form:"labelSelector"`
}
