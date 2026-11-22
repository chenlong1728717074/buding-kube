package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

const (
	// EventDefaultLimit 默认只保留最近的事件条数(查询全部时不限制)
	EventDefaultLimit = 1000
	// eventCacheTTL 翻页时复用最近一次查询的结果, 不重新查询集群
	eventCacheTTL = 10 * time.Second
)

// EventListResult 过滤后的事件(按最后发生时间倒序)
type EventListResult struct {
	Items []vo.EventVO
	// Matched 符合条件的事件总数(截断前)
	Matched int
	// Limited 是否只保留了最近 EventDefaultLimit 条
	Limited bool
}

type eventCacheEntry struct {
	at     time.Time
	result *EventListResult
}

var eventCache sync.Map

func eventCacheKey(q dto.EventQueryDTO) string {
	return strings.Join([]string{q.ClusterId, q.Namespace, q.Type, q.Kind, q.Name, q.Uid, strings.ToLower(strings.TrimSpace(q.Keyword)), fmt.Sprint(q.All)}, "\x00")
}

// pruneEventCache 清理过期的缓存, 避免查询条件很多时缓存一直增长
func pruneEventCache() {
	eventCache.Range(func(k, v any) bool {
		if time.Since(v.(eventCacheEntry).at) > eventCacheTTL {
			eventCache.Delete(k)
		}
		return true
	})
}

// ListEvents 事件中心: 按命名空间、类型、涉及对象过滤, 关键字匹配原因 / 消息 / 对象名称, 按最后发生时间倒序.
// apiserver 不支持按时间排序和截取, 需要取回后排序; 默认只保留最近 EventDefaultLimit 条, q.All 为 true 时返回全部
func ListEvents(ctx context.Context, q dto.EventQueryDTO) (*EventListResult, error) {
	key := eventCacheKey(q)
	if !q.Refresh {
		if v, ok := eventCache.Load(key); ok {
			if e := v.(eventCacheEntry); time.Since(e.at) < eventCacheTTL {
				return e.result, nil
			}
		}
	}
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
	if keyword := strings.ToLower(strings.TrimSpace(q.Keyword)); keyword != "" {
		filtered := make([]vo.EventVO, 0, len(list))
		for _, e := range list {
			if strings.Contains(strings.ToLower(e.Reason), keyword) ||
				strings.Contains(strings.ToLower(e.Message), keyword) ||
				strings.Contains(strings.ToLower(e.ObjectName), keyword) {
				filtered = append(filtered, e)
			}
		}
		list = filtered
	}
	result := &EventListResult{Items: list, Matched: len(list)}
	if !q.All && len(list) > EventDefaultLimit {
		result.Items = list[:EventDefaultLimit]
		result.Limited = true
	}
	pruneEventCache()
	eventCache.Store(key, eventCacheEntry{at: time.Now(), result: result})
	return result, nil
}
