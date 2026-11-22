package service

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// 约定: 列表接口统一使用 ResourceVersion "0" 由 apiserver 的 watch 缓存返回, 不逐条生成 YAML;
// YAML 由 resource_yaml.go 中的 ResourceYaml 按需生成.

// tableAccept 让 apiserver 以 Table 格式返回列表(与 kubectl get 相同):
// 每行只带展示列(如 Secret 的类型、数据条数)和对象元数据, 不返回 data 等大字段
const tableAccept = "application/json;as=Table;v=v1;g=meta.k8s.io,application/json"

// tableRow 一行 Table 结果: 对象元数据 + 展示列(按列名索引, 同时保留原始顺序)
type tableRow struct {
	Meta     metav1.ObjectMeta
	Cells    map[string]interface{}
	RawCells []interface{}
}

// decodeTable 解析 Table 格式的列表结果
func decodeTable(raw []byte, resource string) ([]metav1.TableColumnDefinition, []tableRow, error) {
	var table metav1.Table
	if err := json.Unmarshal(raw, &table); err != nil {
		return nil, nil, fmt.Errorf("解析%s列表失败: %w", resource, err)
	}
	if table.Kind != "Table" {
		return nil, nil, fmt.Errorf("apiserver 未返回 Table 格式(kind=%s)", table.Kind)
	}
	rows := make([]tableRow, 0, len(table.Rows))
	for _, r := range table.Rows {
		var meta metav1.PartialObjectMetadata
		if len(r.Object.Raw) > 0 {
			if err := json.Unmarshal(r.Object.Raw, &meta); err != nil {
				return nil, nil, fmt.Errorf("解析%s元数据失败: %w", resource, err)
			}
		}
		cells := make(map[string]interface{}, len(table.ColumnDefinitions))
		for i, col := range table.ColumnDefinitions {
			if i < len(r.Cells) {
				cells[col.Name] = r.Cells[i]
			}
		}
		rows = append(rows, tableRow{Meta: meta.ObjectMeta, Cells: cells, RawCells: r.Cells})
	}
	return table.ColumnDefinitions, rows, nil
}

// listAsTable 以 Table 格式列出某种核心资源(core/v1), 只取元数据和展示列.
//   - namespace 为空时查询所有命名空间, 只发一次请求
//   - resourceVersion=0: 由 apiserver 的 watch 缓存直接返回, 不必去 etcd 做一致性读
//
// 适合 Secret / ConfigMap 这类单个对象可能很大, 但列表页只需要名称、时间等少量字段的资源
func listAsTable(ctx context.Context, client rest.Interface, resource, namespace string) ([]tableRow, error) {
	raw, err := client.Get().
		Namespace(namespace).
		Resource(resource).
		Param("resourceVersion", "0").
		Param("includeObject", string(metav1.IncludeMetadata)).
		SetHeader("Accept", tableAccept).
		DoRaw(ctx)
	if err != nil {
		return nil, err
	}
	_, rows, err := decodeTable(raw, resource)
	return rows, err
}

// cellString 读取字符串类型的展示列
func (r tableRow) cellString(name string) string {
	v, _ := r.Cells[name].(string)
	return v
}

// cellInt 读取数字类型的展示列(JSON 数字解析后是 float64)
func (r tableRow) cellInt(name string) int {
	v, _ := r.Cells[name].(float64)
	return int(v)
}
