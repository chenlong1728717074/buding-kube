// 弹性伸缩(HorizontalPodAutoscaler autoscaling/v2)的展示
import type { KubeObject } from '@/api/resource'

export interface HpaMetricRow {
  label: string
  current: string
  target: string
}

const RESOURCE_TEXT: Record<string, string> = { cpu: 'CPU', memory: '内存' }

const targetText = (t?: KubeObject) => {
  if (!t) return '-'
  if (t.type === 'Utilization') return `${t.averageUtilization}%`
  if (t.type === 'AverageValue') return `平均 ${t.averageValue}`
  return String(t.value ?? '-')
}

const currentText = (c?: KubeObject) => {
  if (!c) return '-'
  if (c.averageUtilization !== undefined) return `${c.averageUtilization}%${c.averageValue ? `（${c.averageValue}）` : ''}`
  return String(c.averageValue ?? c.value ?? '-')
}

// 指标: 当前值来自 status.currentMetrics, 按类型和名称匹配
export function hpaMetricRows(hpa: KubeObject): HpaMetricRow[] {
  const current: KubeObject[] = hpa.status?.currentMetrics || []
  return (hpa.spec?.metrics || []).map((m: KubeObject) => {
    if (m.type === 'Resource' || m.type === 'ContainerResource') {
      const src = m.resource || m.containerResource
      const cur = current.find(c => c.type === m.type && (c.resource || c.containerResource)?.name === src?.name)
      const label = `${RESOURCE_TEXT[src?.name] || src?.name}${src?.target?.type === 'Utilization' ? ' 使用率' : ' 用量'}${m.type === 'ContainerResource' ? `（容器 ${src?.container}）` : ''}`
      return { label, current: currentText((cur?.resource || cur?.containerResource)?.current), target: targetText(src?.target) }
    }
    const key = m.type === 'Pods' ? 'pods' : m.type === 'Object' ? 'object' : 'external'
    const src = m[key]
    const cur = current.find(c => c.type === m.type && c[key]?.metric?.name === src?.metric?.name)
    return { label: `${m.type} ${src?.metric?.name || ''}`, current: currentText(cur?.[key]?.current), target: targetText(src?.target) }
  })
}

export const HPA_CONDITION_TEXT: Record<string, string> = {
  AbleToScale: '可以伸缩（AbleToScale）',
  ScalingActive: '指标可用（ScalingActive）',
  ScalingLimited: '达到副本上下限（ScalingLimited）'
}

export const hpaTargetKind = (kind: string) => (kind === 'statefulset' ? 'StatefulSet' : 'Deployment')
