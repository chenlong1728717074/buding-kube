package vo

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"sort"
	"time"
)

type EventVO struct {
	Type      string    `json:"type"`      // Normal, Warning
	Reason    string    `json:"reason"`    // 事件原因
	Object    string    `json:"object"`    // 涉及的对象
	Message   string    `json:"message"`   // 事件消息
	Source    string    `json:"source"`    // 事件源组件
	Count     int32     `json:"count"`     // 事件发生次数
	FirstTime time.Time `json:"firstTime"` // 首次发生时间
	LastTime  time.Time `json:"lastTime"`  // 最后发生时间
	Age       string    `json:"age"`       // 距离现在的时间
}

func buildNodeEvents(events *corev1.EventList) []EventVO {
	result := make([]EventVO, 0)

	if events == nil {
		return result
	}

	for _, event := range events.Items {
		eventVO := EventVO{
			Type:      event.Type,
			Reason:    event.Reason,
			Object:    fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
			Message:   event.Message,
			Source:    event.Source.Component,
			Count:     event.Count,
			FirstTime: event.FirstTimestamp.Time,
			LastTime:  event.LastTimestamp.Time,
			Age:       time.Since(event.LastTimestamp.Time).String(),
		}

		// 如果FirstTimestamp为空，使用EventTime
		if event.FirstTimestamp.IsZero() && !event.EventTime.IsZero() {
			eventVO.FirstTime = event.EventTime.Time
		}

		// 如果LastTimestamp为空，使用EventTime
		if event.LastTimestamp.IsZero() && !event.EventTime.IsZero() {
			eventVO.LastTime = event.EventTime.Time
		}

		result = append(result, eventVO)
	}

	// 按时间倒序排列（最新的在前面）
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastTime.After(result[j].LastTime)
	})

	return result
}
