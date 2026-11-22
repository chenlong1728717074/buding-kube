// YAML 添加时的示例
export const REPLICASET_YAML = (ns: string) => `apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: example-rs
  namespace: ${ns}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: example-rs
  template:
    metadata:
      labels:
        app: example-rs
    spec:
      containers:
        - name: nginx
          image: nginx:1.27
`

export const HPA_YAML = (ns: string) => `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: example-hpa
  namespace: ${ns}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: example
  minReplicas: 1
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
`

export const QUOTA_YAML = (ns: string) => `apiVersion: v1
kind: ResourceQuota
metadata:
  name: default-quota
  namespace: ${ns}
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    pods: "50"
`

export const LIMIT_RANGE_YAML = (ns: string) => `apiVersion: v1
kind: LimitRange
metadata:
  name: default-limits
  namespace: ${ns}
spec:
  limits:
    - type: Container
      defaultRequest:
        cpu: 100m
        memory: 128Mi
      default:
        cpu: 500m
        memory: 512Mi
`

export const NETWORK_POLICY_YAML = (ns: string) => `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-ingress
  namespace: ${ns}
spec:
  podSelector: {}
  policyTypes:
    - Ingress
`

export const PDB_YAML = (ns: string) => `apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: example-pdb
  namespace: ${ns}
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: example
`

export const PRIORITY_CLASS_YAML = () => `apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 1000000
globalDefault: false
preemptionPolicy: PreemptLowerPriority
description: 重要业务使用
`

export const INGRESS_CLASS_YAML = () => `apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: nginx
spec:
  controller: k8s.io/ingress-nginx
`

// 自定义资源的示例: 只填 apiVersion / kind / metadata
export const customResourceYaml = (apiVersion: string, kind: string, namespace?: string) => `apiVersion: ${apiVersion}
kind: ${kind}
metadata:
  name: example
${namespace ? `  namespace: ${namespace}\n` : ''}spec: {}
`
