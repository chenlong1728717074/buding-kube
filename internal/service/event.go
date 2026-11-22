package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

// ListEvents 事件中心: 按命名空间、类型、涉及对象过滤, 关键字匹配原因 / 消息 / 对象名称, 按最后发生时间倒序
func ListEvents(ctx context.Context, q dto.EventQueryDTO) ([]vo.EventVO, error) {
	cs, err := ClusterMap.GetClientSet(q.ClusterId)
	if err != nil {
		return nil, err
	}
	selectors := make([]fields.Selector, 0, 4)
	add := func(field, value string) {
		if value = strings.TrimSpace(value); value != "" {
			selectors = append(selectors, fields.OneTermEqualSelector(field, value))
		}
	}
	add("type", q.Type)
	add("involvedObject.kind", q.Kind)
	add("involvedObject.name", q.Name)
	add("involvedObject.uid", q.Uid)
	opts := metav1.ListOptions{ResourceVersion: "0"}
	if len(selectors) > 0 {
		opts.FieldSelector = fields.AndSelectors(selectors...).String()
	}
	events, err := cs.CoreV1().Events(strings.TrimSpace(q.Namespace)).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("查询事件失败: %v", err)
	}
	list := vo.Events2VO(events)
	keyword := strings.ToLower(strings.TrimSpace(q.Keyword))
	if keyword == "" {
		return list, nil
	}
	result := make([]vo.EventVO, 0, len(list))
	for _, e := range list {
		if strings.Contains(strings.ToLower(e.Reason), keyword) ||
			strings.Contains(strings.ToLower(e.Message), keyword) ||
			strings.Contains(strings.ToLower(e.ObjectName), keyword) {
			result = append(result, e)
		}
	}
	return result, nil
}
