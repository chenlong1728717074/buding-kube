package dto

type PodDTO struct {
	NamespaceBaseDTO
	Name string `json:"name" form:"name" binding:"required"`
}

type PodQueryDTO struct {
	NamespaceBaseDTO
	PageQueryDTO
	Status string `json:"status" form:"status"`
}

type PodLogDTO struct {
	PodDTO
	SinceTime     *CustomTime `json:"sinceTime"`
	ContainerName string      `json:"containerName"`
	Follow        bool        `json:"follow"`
	TailLines     *int64      `json:"tailLines"`
	Container     string      `json:"container"`
}

type PodDownloadDTO struct {
	PodDTO
	FilePath      string `json:"filePath" form:"filePath"`
	ContainerName string `json:"containerName" form:"containerName"`
}
