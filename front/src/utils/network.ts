// Service / Ingress 共用的展示工具
import type { ServicePortVO, ServiceVO } from '@/api/service'

export const SERVICE_TYPE_OPTIONS = [
  { value: 'ClusterIP', label: '虚拟 IP（ClusterIP）' },
  { value: 'Headless', label: '内部域名（Headless）' },
  { value: 'NodePort', label: '节点端口（NodePort）' },
  { value: 'LoadBalancer', label: '负载均衡器（LoadBalancer）' },
  { value: 'ExternalName', label: '外部服务（ExternalName）' }
]

export const serviceTypeLabel = (svc: Pick<ServiceVO, 'serviceType' | 'headless'>) => {
  if (svc.headless) return 'Headless'
  return svc.serviceType || 'ClusterIP'
}

export const serviceTypeTag = (svc: Pick<ServiceVO, 'serviceType' | 'headless'>) => {
  if (svc.headless) return 'success'
  return (({ NodePort: 'warning', LoadBalancer: 'danger', ExternalName: 'info' }) as Record<string, 'warning' | 'danger' | 'info'>)[svc.serviceType] || 'primary'
}

// 80/TCP→8080
export const servicePortText = (p: ServicePortVO) => {
  let text = `${p.port}/${p.protocol}`
  if (p.targetPort && p.targetPort !== String(p.port)) text += `→${p.targetPort}`
  if (p.nodePort) text += `:${p.nodePort}`
  return text
}

export const PATH_TYPE_OPTIONS = [
  { value: 'Prefix', label: '前缀匹配（Prefix）' },
  { value: 'Exact', label: '精确匹配（Exact）' },
  { value: 'ImplementationSpecific', label: '由控制器决定（ImplementationSpecific）' }
]

// 常用的 ingress-nginx 注解, 创建应用路由时一键添加
export const NGINX_ANNOTATION_PRESETS = [
  { key: 'nginx.ingress.kubernetes.io/rewrite-target', value: '/$2', desc: '路径重写' },
  { key: 'nginx.ingress.kubernetes.io/ssl-redirect', value: 'true', desc: 'HTTP 跳转 HTTPS' },
  { key: 'nginx.ingress.kubernetes.io/proxy-body-size', value: '100m', desc: '请求体大小上限' },
  { key: 'nginx.ingress.kubernetes.io/backend-protocol', value: 'HTTPS', desc: '后端协议' },
  { key: 'nginx.ingress.kubernetes.io/proxy-read-timeout', value: '600', desc: '读取超时（秒）' },
  { key: 'nginx.ingress.kubernetes.io/whitelist-source-range', value: '10.0.0.0/8', desc: '访问白名单' },
  { key: 'nginx.ingress.kubernetes.io/enable-cors', value: 'true', desc: '跨域' }
]

export const serviceYamlExample = (namespace = 'default') =>
  [
    'apiVersion: v1',
    'kind: Service',
    'metadata:',
    '  name: example-service',
    `  namespace: ${namespace}`,
    'spec:',
    '  type: ClusterIP',
    '  selector:',
    '    app: example',
    '  ports:',
    '    - name: http',
    '      protocol: TCP',
    '      port: 80',
    '      targetPort: 8080'
  ].join('\n')

export const ingressYamlExample = (namespace = 'default') =>
  [
    'apiVersion: networking.k8s.io/v1',
    'kind: Ingress',
    'metadata:',
    '  name: example-ingress',
    `  namespace: ${namespace}`,
    'spec:',
    '  ingressClassName: nginx',
    '  rules:',
    '    - host: demo.example.com',
    '      http:',
    '        paths:',
    '          - path: /',
    '            pathType: Prefix',
    '            backend:',
    '              service:',
    '                name: example-service',
    '                port:',
    '                  number: 80'
  ].join('\n')
