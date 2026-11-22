package vo

import "time"

// ResourceTableColumnVO 列定义(与 kubectl get 的列相同); Priority 大于 0 的列对应 kubectl get -o wide
type ResourceTableColumnVO struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Format      string `json:"format"`
	Description string `json:"description"`
	Priority    int32  `json:"priority"`
}

// ResourceTableRowVO 一行数据; Cells 与 Columns 一一对应
type ResourceTableRowVO struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Uid        string            `json:"uid"`
	CreateTime time.Time         `json:"createTime"`
	Labels     map[string]string `json:"labels"`
	OwnerKind  string            `json:"ownerKind"`
	OwnerName  string            `json:"ownerName"`
	Cells      []interface{}     `json:"cells"`
}

// ResourceTableVO 通用资源列表
type ResourceTableVO struct {
	Columns  []ResourceTableColumnVO `json:"columns"`
	Items    []ResourceTableRowVO    `json:"items"`
	Total    int                     `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"pageSize"`
}
