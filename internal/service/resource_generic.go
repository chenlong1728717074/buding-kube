package service

import (
	"buding-kube/internal/web/vo"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// 通用资源接口: 列表(Table 格式, 与 kubectl get 的列一致)、对象、删除.
// 适合 ReplicaSet / HPA / NetworkPolicy / ResourceQuota / CRD 实例等不需要专门页面逻辑的资源

const maxResourcePageSize = 500

// ResourceTableQuery 通用列表的查询条件
type ResourceTableQuery struct {
	ClusterId     string
	Ref           ResourceRef
	Namespace     string
	Keyword       string
	LabelSelector string
	Page          int
	PageSize      int
}

// apiPath 资源的 REST 路径, 例如 /apis/apps/v1/namespaces/default/replicasets
func apiPath(rk resourceKind, namespace string) string {
	var b strings.Builder
	if rk.gvr.Group == "" {
		b.WriteString("/api/" + rk.gvr.Version)
	} else {
		b.WriteString("/apis/" + rk.gvr.Group + "/" + rk.gvr.Version)
	}
	if rk.namespaced && namespace != "" {
		b.WriteString("/namespaces/" + url.PathEscape(namespace))
	}
	b.WriteString("/" + rk.gvr.Resource)
	return b.String()
}

func resourceError(rk resourceKind, action string, err error) error {
	switch {
	case apierrors.IsNotFound(err):
		return fmt.Errorf("集群中没有 %s（%s/%s）资源，可能没有安装对应的组件", rk.kind, rk.gvr.GroupVersion().String(), rk.gvr.Resource)
	case apierrors.IsForbidden(err):
		return fmt.Errorf("没有权限%s %s，请检查集群凭证的 RBAC 权限", action, rk.kind)
	default:
		return fmt.Errorf("%s %s 失败: %v", action, rk.kind, err)
	}
}

// ListResourceTable 以 Table 格式列出资源, 支持按名称关键字、标签选择器过滤和分页
func ListResourceTable(ctx context.Context, q ResourceTableQuery) (*vo.ResourceTableVO, error) {
	rk, err := q.Ref.resolve()
	if err != nil {
		return nil, err
	}
	cs, err := ClusterMap.GetClientSet(q.ClusterId)
	if err != nil {
		return nil, err
	}
	req := cs.RESTClient().Get().AbsPath(apiPath(rk, q.Namespace)).
		Param("resourceVersion", "0").
		Param("includeObject", string(metav1.IncludeMetadata)).
		SetHeader("Accept", tableAccept)
	if s := strings.TrimSpace(q.LabelSelector); s != "" {
		req = req.Param("labelSelector", s)
	}
	raw, err := req.DoRaw(ctx)
	if err != nil {
		return nil, resourceError(rk, "查询", err)
	}
	columns, rows, err := decodeTable(raw, rk.gvr.Resource)
	if err != nil {
		return nil, err
	}

	result := &vo.ResourceTableVO{Columns: make([]vo.ResourceTableColumnVO, 0, len(columns)), Items: []vo.ResourceTableRowVO{}}
	for _, c := range columns {
		result.Columns = append(result.Columns, vo.ResourceTableColumnVO{Name: c.Name, Type: c.Type, Format: c.Format, Description: c.Description, Priority: c.Priority})
	}
	keyword := strings.ToLower(strings.TrimSpace(q.Keyword))
	filtered := make([]tableRow, 0, len(rows))
	for _, r := range rows {
		if keyword == "" || strings.Contains(strings.ToLower(r.Meta.Name), keyword) {
			filtered = append(filtered, r)
		}
	}
	result.Total = len(filtered)
	result.Page, result.PageSize = q.Page, q.PageSize
	if result.Page < 1 {
		result.Page = 1
	}
	if result.PageSize < 1 {
		result.PageSize = 10
	}
	if result.PageSize > maxResourcePageSize {
		result.PageSize = maxResourcePageSize
	}
	start := (result.Page - 1) * result.PageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := min(start+result.PageSize, len(filtered))
	for _, r := range filtered[start:end] {
		row := vo.ResourceTableRowVO{
			Name:       r.Meta.Name,
			Namespace:  r.Meta.Namespace,
			Uid:        string(r.Meta.UID),
			CreateTime: r.Meta.CreationTimestamp.Time,
			Labels:     r.Meta.Labels,
			Cells:      r.RawCells,
		}
		if owner := metav1.GetControllerOfNoCopy(&r.Meta); owner != nil {
			row.OwnerKind, row.OwnerName = owner.Kind, owner.Name
		}
		if row.Cells == nil {
			row.Cells = []interface{}{}
		}
		result.Items = append(result.Items, row)
	}
	return result, nil
}

// cleanObject 去掉 managedFields, 前端展示用
func cleanObject(obj *unstructured.Unstructured) map[string]interface{} {
	content := obj.Object
	if meta, ok := content["metadata"].(map[string]interface{}); ok {
		delete(meta, "managedFields")
	}
	return content
}

// GetResourceObject 查询单个资源的完整对象(包含 status)
func GetResourceObject(ctx context.Context, clusterId string, ref ResourceRef, namespace, name string) (map[string]interface{}, error) {
	rk, err := ref.resolve()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("资源名称不能为空")
	}
	if rk.namespaced && namespace == "" {
		return nil, errors.New("命名空间不能为空")
	}
	dyn, err := getDynamicClient(clusterId)
	if err != nil {
		return nil, err
	}
	obj, err := resourceClient(dyn, rk, namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("%s %s 不存在", rk.kind, name)
		}
		return nil, resourceError(rk, "获取", err)
	}
	return cleanObject(obj), nil
}

// ListResourceObjects 列出完整对象; namespace 为空时查询所有命名空间
func ListResourceObjects(ctx context.Context, clusterId string, ref ResourceRef, namespace, labelSelector string) ([]map[string]interface{}, error) {
	rk, err := ref.resolve()
	if err != nil {
		return nil, err
	}
	dyn, err := getDynamicClient(clusterId)
	if err != nil {
		return nil, err
	}
	list, err := resourceClient(dyn, rk, namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0", LabelSelector: strings.TrimSpace(labelSelector)})
	if err != nil {
		return nil, resourceError(rk, "查询", err)
	}
	result := make([]map[string]interface{}, 0, len(list.Items))
	for i := range list.Items {
		result = append(result, cleanObject(&list.Items[i]))
	}
	return result, nil
}

// DeleteResource 删除资源, 由它创建的子资源在后台级联删除
func DeleteResource(ctx context.Context, clusterId string, ref ResourceRef, namespace, name string) error {
	rk, err := ref.resolve()
	if err != nil {
		return err
	}
	if strings.TrimSpace(name) == "" {
		return errors.New("资源名称不能为空")
	}
	if rk.namespaced && namespace == "" {
		return errors.New("命名空间不能为空")
	}
	dyn, err := getDynamicClient(clusterId)
	if err != nil {
		return err
	}
	propagation := metav1.DeletePropagationBackground
	err = resourceClient(dyn, rk, namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	if err != nil && !apierrors.IsNotFound(err) {
		return resourceError(rk, "删除", err)
	}
	return nil
}
