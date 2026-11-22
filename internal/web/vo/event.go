package vo

import (
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
)

type EventVO struct {
	Type      string    `json:"type"`      // Normal, Warning
	Reason    string    `json:"reason"`    // 事件原因
	Object    string    `json:"object"`    // 涉及的对象, Kind/Name
	Message   string    `json:"message"`   // 事件消息
	Source    string    `json:"source"`    // 事件源组件
	Count     int32     `json:"count"`     // 事件发生次数
	FirstTime time.Time `json:"firstTime"` // 首次发生时间
	LastTime  time.Time `json:"lastTime"`  // 最后发生时间
	// 以下字段用于事件中心跳转到对象详情
	Namespace       string `json:"namespace"`
	ObjectKind      string `json:"objectKind"`
	ObjectName      string `json:"objectName"`
	ObjectNamespace string `json:"objectNamespace"`
}

// Event2VO 兼容新旧两种事件格式: 新版组件只设置 eventTime / series, 不设置 firstTimestamp / lastTimestamp / count
func Event2VO(e *corev1.Event) EventVO {
	item := EventVO{
		Type:            e.Type,
		Reason:          e.Reason,
		Object:          fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name),
		Message:         e.Message,
		Source:          firstNonEmpty(e.Source.Component, e.ReportingController),
		Count:           e.Count,
		FirstTime:       e.FirstTimestamp.Time,
		LastTime:        e.LastTimestamp.Time,
		Namespace:       e.Namespace,
		ObjectKind:      e.InvolvedObject.Kind,
		ObjectName:      e.InvolvedObject.Name,
		ObjectNamespace: e.InvolvedObject.Namespace,
	}
	if item.FirstTime.IsZero() && !e.EventTime.IsZero() {
		item.FirstTime = e.EventTime.Time
	}
	if item.LastTime.IsZero() {
		switch {
		case e.Series != nil && !e.Series.LastObservedTime.IsZero():
			item.LastTime = e.Series.LastObservedTime.Time
		case !e.EventTime.IsZero():
			item.LastTime = e.EventTime.Time
		default:
			item.LastTime = e.CreationTimestamp.Time
		}
	}
	if item.Count == 0 {
		item.Count = 1
		if e.Series != nil && e.Series.Count > 0 {
			item.Count = e.Series.Count
		}
	}
	return item
}

// Events2VO 事件列表, 按最后发生时间倒序
func Events2VO(events *corev1.EventList) []EventVO {
	result := make([]EventVO, 0)
	if events == nil {
		return result
	}
	for i := range events.Items {
		result = append(result, Event2VO(&events.Items[i]))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastTime.After(result[j].LastTime)
	})
	return result
}
