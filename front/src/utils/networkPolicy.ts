// 网络策略表单与 NetworkPolicy 对象的相互转换; 表单只支持常用写法, 其他写法请使用 YAML
import type { KubeObject } from '@/api/resource'
import type { KVRow } from '@/utils/job'

export type PeerType = 'namespacePods' | 'pods' | 'namespaces' | 'ipBlock'

export interface PeerForm {
  type: PeerType
  labels: KVRow[]
  namespaces: string[]
  cidr: string
  except: string
}

export interface PortForm {
  protocol: 'TCP' | 'UDP' | 'SCTP'
  port: string
}

export interface RuleForm {
  // 任意来源 / 目标
  any: boolean
  peers: PeerForm[]
  // 为空表示所有端口
  ports: PortForm[]
}

export interface NetworkPolicyForm {
  podScope: 'all' | 'labels'
  podLabels: KVRow[]
  ingressEnabled: boolean
  ingress: RuleForm[]
  egressEnabled: boolean
  egress: RuleForm[]
}

export const PEER_TYPES: { value: PeerType; label: string }[] = [
  { value: 'namespacePods', label: '本命名空间的所有容器组' },
  { value: 'pods', label: '本命名空间指定标签的容器组' },
  { value: 'namespaces', label: '指定命名空间的所有容器组' },
  { value: 'ipBlock', label: 'IP 段（CIDR）' }
]

const NS_LABEL = 'kubernetes.io/metadata.name'

export const createPeer = (type: PeerType = 'namespacePods'): PeerForm => ({ type, labels: [], namespaces: [], cidr: '', except: '' })
export const createRule = (any = false): RuleForm => ({ any, peers: any ? [] : [createPeer()], ports: [] })
export const createNetworkPolicyForm = (): NetworkPolicyForm => ({
  podScope: 'all',
  podLabels: [],
  ingressEnabled: true,
  ingress: [],
  egressEnabled: false,
  egress: []
})

const rowsToMap = (rows: KVRow[]) => Object.fromEntries(rows.filter(r => r.key.trim()).map(r => [r.key.trim(), r.value.trim()]))
const mapToRows = (m?: Record<string, string>) => Object.entries(m || {}).map(([key, value]) => ({ key, value }))

const buildPeer = (p: PeerForm): KubeObject => {
  switch (p.type) {
    case 'namespacePods':
      return { podSelector: {} }
    case 'pods':
      return { podSelector: { matchLabels: rowsToMap(p.labels) } }
    case 'namespaces':
      return { namespaceSelector: { matchExpressions: [{ key: NS_LABEL, operator: 'In', values: p.namespaces }] } }
    default: {
      const except = p.except.split(/[,，\s]+/).map(s => s.trim()).filter(Boolean)
      return { ipBlock: except.length ? { cidr: p.cidr.trim(), except } : { cidr: p.cidr.trim() } }
    }
  }
}

const buildRule = (r: RuleForm, peerKey: 'from' | 'to'): KubeObject => {
  const rule: KubeObject = {}
  if (!r.any && r.peers.length) rule[peerKey] = r.peers.map(buildPeer)
  const ports = r.ports.filter(p => p.port.trim() || p.protocol)
  if (ports.length) {
    rule.ports = ports.map(p => {
      const port = p.port.trim()
      return port ? { protocol: p.protocol, port: /^\d+$/.test(port) ? Number(port) : port } : { protocol: p.protocol }
    })
  }
  return rule
}

// 表单 -> spec
export function buildNetworkPolicySpec(f: NetworkPolicyForm): KubeObject {
  const spec: KubeObject = { podSelector: f.podScope === 'all' ? {} : { matchLabels: rowsToMap(f.podLabels) } }
  const types: string[] = []
  if (f.ingressEnabled) {
    types.push('Ingress')
    spec.ingress = f.ingress.map(r => buildRule(r, 'from'))
  }
  if (f.egressEnabled) {
    types.push('Egress')
    spec.egress = f.egress.map(r => buildRule(r, 'to'))
  }
  spec.policyTypes = types
  return spec
}

class Unsupported extends Error {}

const parseLabels = (selector: KubeObject | undefined): KVRow[] => {
  if (selector?.matchExpressions?.length) throw new Unsupported('选择器使用了 matchExpressions')
  return mapToRows(selector?.matchLabels)
}

const parsePeer = (p: KubeObject): PeerForm => {
  if (p.ipBlock) return { ...createPeer('ipBlock'), cidr: p.ipBlock.cidr || '', except: (p.ipBlock.except || []).join(', ') }
  if (p.podSelector && p.namespaceSelector) throw new Unsupported('同时使用了命名空间选择器和容器组选择器')
  if (p.podSelector) {
    const labels = parseLabels(p.podSelector)
    return labels.length ? { ...createPeer('pods'), labels } : createPeer('namespacePods')
  }
  if (p.namespaceSelector) {
    const ns = p.namespaceSelector
    const exprs: KubeObject[] = ns.matchExpressions || []
    const labels = ns.matchLabels || {}
    if (!exprs.length && Object.keys(labels).length === 1 && labels[NS_LABEL]) return { ...createPeer('namespaces'), namespaces: [labels[NS_LABEL]] }
    if (exprs.length === 1 && !Object.keys(labels).length && exprs[0].key === NS_LABEL && exprs[0].operator === 'In') {
      return { ...createPeer('namespaces'), namespaces: exprs[0].values || [] }
    }
    throw new Unsupported('命名空间选择器使用了自定义标签')
  }
  throw new Unsupported('来源 / 目标格式不支持')
}

const parseRule = (r: KubeObject, peerKey: 'from' | 'to'): RuleForm => {
  const peers: KubeObject[] = r[peerKey] || []
  const ports: KubeObject[] = r.ports || []
  if (ports.some(p => p.endPort)) throw new Unsupported('端口使用了 endPort 范围')
  return {
    any: !peers.length,
    peers: peers.map(parsePeer),
    ports: ports.map(p => ({ protocol: p.protocol || 'TCP', port: p.port === undefined ? '' : String(p.port) }))
  }
}

// spec -> 表单; 遇到表单不支持的写法时返回原因
export function parseNetworkPolicySpec(spec: KubeObject = {}): { form: NetworkPolicyForm } | { reason: string } {
  try {
    const labels = parseLabels(spec.podSelector)
    const types: string[] = spec.policyTypes?.length ? spec.policyTypes : ['Ingress', ...(spec.egress ? ['Egress'] : [])]
    return {
      form: {
        podScope: labels.length ? 'labels' : 'all',
        podLabels: labels,
        ingressEnabled: types.includes('Ingress'),
        ingress: (spec.ingress || []).map((r: KubeObject) => parseRule(r, 'from')),
        egressEnabled: types.includes('Egress'),
        egress: (spec.egress || []).map((r: KubeObject) => parseRule(r, 'to'))
      }
    }
  } catch (e) {
    if (e instanceof Unsupported) return { reason: e.message }
    throw e
  }
}

// 预设
export const NETWORK_POLICY_PRESETS: { key: string; label: string; desc: string; apply: () => NetworkPolicyForm }[] = [
  {
    key: 'deny-ingress',
    label: '拒绝所有入站',
    desc: '命名空间内的容器组不接收任何入站流量',
    apply: () => ({ ...createNetworkPolicyForm(), ingressEnabled: true, ingress: [] })
  },
  {
    key: 'same-namespace',
    label: '只允许本命名空间访问',
    desc: '只接收来自同一命名空间容器组的流量',
    apply: () => ({ ...createNetworkPolicyForm(), ingressEnabled: true, ingress: [createRule()] })
  },
  {
    key: 'deny-egress',
    label: '拒绝出站（保留 DNS）',
    desc: '容器组只能访问集群 DNS，不能访问其他地址',
    apply: () => ({
      ...createNetworkPolicyForm(),
      ingressEnabled: false,
      egressEnabled: true,
      egress: [{ any: true, peers: [], ports: [{ protocol: 'UDP', port: '53' }, { protocol: 'TCP', port: '53' }] }]
    })
  }
]

const selectorText = (sel?: KubeObject) => {
  const labels = Object.entries(sel?.matchLabels || {}).map(([k, v]) => `${k}=${v}`)
  const exprs = (sel?.matchExpressions || []).map((e: KubeObject) => `${e.key} ${e.operator} ${(e.values || []).join(',')}`)
  return [...labels, ...exprs].join(', ')
}

// 列表展示: 应用于哪些容器组
export const podSelectorText = (spec?: KubeObject) => selectorText(spec?.podSelector) || '所有容器组'

// 列表展示: 入站 / 出站规则概况
export function directionText(spec: KubeObject | undefined, dir: 'Ingress' | 'Egress') {
  const types: string[] = spec?.policyTypes?.length ? spec.policyTypes : ['Ingress', ...(spec?.egress ? ['Egress'] : [])]
  if (!types.includes(dir)) return '不限制'
  const rules: KubeObject[] = (dir === 'Ingress' ? spec?.ingress : spec?.egress) || []
  if (!rules.length) return '全部拒绝'
  if (rules.some(r => !(r.from || r.to)?.length && !r.ports?.length)) return '全部允许'
  return `${rules.length} 条放行规则`
}
