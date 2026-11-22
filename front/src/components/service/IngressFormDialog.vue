<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="1120px"
    top="5vh"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--wide"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="reset"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">{{ title }}</h3>
        <div class="dialog-actions">
          <el-switch v-model="yamlMode" active-text="YAML 预览" />
        </div>
      </div>
    </template>

    <div v-loading="loading" class="config-editor">
      <div v-if="!yamlMode" class="step-tabs">
        <div
          v-for="(step, index) in steps"
          :key="step.title"
          :class="['step-tab', { active: current === index, completed: index < current }]"
          @click="goStep(index)"
        >
          <div class="step-icon">
            <el-icon v-if="index < current"><Check /></el-icon>
            <span v-else>{{ index + 1 }}</span>
          </div>
          <div class="step-info">
            <div class="step-title">{{ step.title }}</div>
            <div class="step-desc">{{ step.desc }}</div>
          </div>
        </div>
      </div>

      <div class="config-content">
        <div v-if="yamlMode" class="yaml-view">
          <YamlEditor :model-value="previewYaml" :readonly="true" height="100%" title="Ingress YAML 预览" :filename="`${form.name || 'ingress'}.yaml`" />
        </div>

        <!-- 1. 基本信息 -->
        <div v-else-if="current === 0" class="basic-form">
          <el-form ref="basicRef" :model="form" :rules="rules" label-width="100px">
            <el-form-item label="命名空间" prop="namespace">
              <el-select v-model="form.namespace" filterable :disabled="isEdit" placeholder="请选择命名空间" :loading="nsLoading" style="width: 100%" @change="onNamespaceChange">
                <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" :disabled="isEdit" placeholder="例如 web-ingress" maxlength="253" show-word-limit />
            </el-form-item>
            <el-form-item label="别名">
              <el-input v-model="form.alias" placeholder="可选，用于列表展示" maxlength="50" show-word-limit />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="form.describe" type="textarea" :rows="3" placeholder="可选，说明用途" maxlength="200" show-word-limit />
            </el-form-item>
            <el-form-item label="Ingress 类">
              <el-select v-model="form.ingressClassName" clearable placeholder="使用集群默认的 Ingress 类" style="width: 100%">
                <el-option v-for="c in classes" :key="c.name" :value="c.name" :label="c.isDefault ? `${c.name}（默认）` : c.name">
                  <div class="ig-option">
                    <span>{{ c.name }}<em v-if="c.isDefault">默认</em></span>
                    <span class="ig-option__desc">{{ c.controller }}</span>
                  </div>
                </el-option>
              </el-select>
              <div class="field-hint">{{ classes.length ? '决定由哪个 Ingress 控制器处理这条路由' : '集群中没有 IngressClass，需要先安装 ingress-nginx 等控制器' }}</div>
            </el-form-item>
          </el-form>
        </div>

        <!-- 2. 路由规则 -->
        <div v-else-if="current === 1" class="basic-form">
          <div v-for="(rule, ri) in rules_" :key="ri" class="ig-rule">
            <div class="ig-rule__head">
              <span class="ig-rule__title">规则 {{ ri + 1 }}</span>
              <el-radio-group v-model="rule.https" size="small">
                <el-radio-button :value="false">HTTP</el-radio-button>
                <el-radio-button :value="true">HTTPS</el-radio-button>
              </el-radio-group>
              <el-button v-if="rules_.length > 1" link type="danger" class="ig-rule__del" @click="rules_.splice(ri, 1)">
                <el-icon><Delete /></el-icon>
                删除规则
              </el-button>
            </div>
            <el-form :model="rule" label-width="90px">
              <el-form-item label="域名">
                <el-input v-model="rule.host" placeholder="例如 demo.example.com，支持 *.example.com；留空匹配所有域名" class="mono-input" />
              </el-form-item>
              <el-form-item v-if="rule.https" label="证书">
                <el-select v-model="rule.secretName" filterable placeholder="选择 kubernetes.io/tls 类型的 Secret" style="width: 100%" :loading="secretLoading">
                  <el-option v-for="s in tlsSecrets" :key="s" :label="s" :value="s" />
                </el-select>
                <div v-if="!tlsSecrets.length" class="field-hint">命名空间中还没有证书，请先在「Secret管理」创建 TLS 类型的 Secret</div>
              </el-form-item>
            </el-form>
            <div class="ig-paths">
              <div class="ig-paths__head">
                <span style="width: 220px">路径</span>
                <span style="width: 200px">匹配方式</span>
                <span style="width: 220px">服务</span>
                <span style="width: 170px">端口</span>
              </div>
              <div v-for="(p, pi) in rule.paths" :key="pi" class="ig-paths__row">
                <el-input v-model="p.path" placeholder="/" class="mono-input" style="width: 220px" />
                <el-select v-model="p.pathType" style="width: 200px">
                  <el-option v-for="o in PATH_TYPE_OPTIONS" :key="o.value" :label="o.label" :value="o.value" />
                </el-select>
                <el-select v-model="p.serviceName" filterable placeholder="选择服务" style="width: 220px" :loading="svcLoading" @change="onServiceChange(p)">
                  <el-option v-for="s in services" :key="s.name" :label="s.name" :value="s.name" />
                </el-select>
                <el-select v-model="p.servicePort" filterable allow-create default-first-option placeholder="端口" style="width: 170px">
                  <el-option v-for="o in portOptions(p.serviceName)" :key="o.value" :label="o.label" :value="o.value" />
                </el-select>
                <el-button v-if="rule.paths.length > 1" link type="danger" @click="rule.paths.splice(pi, 1)"><el-icon><Delete /></el-icon></el-button>
              </div>
              <el-button size="small" @click="rule.paths.push(createPath())">
                <el-icon><Plus /></el-icon>
                添加路径
              </el-button>
            </div>
          </div>
          <el-button class="ig-add" @click="rules_.push(createRule())">
            <el-icon><Plus /></el-icon>
            添加规则
          </el-button>
        </div>

        <!-- 3. 高级设置 -->
        <div v-else class="basic-form">
          <p class="fd-subtitle">常用注解<small>点击添加 ingress-nginx 的常用设置，其他控制器请参考对应文档</small></p>
          <div class="ig-presets">
            <button v-for="a in NGINX_ANNOTATION_PRESETS" :key="a.key" type="button" class="ig-preset" :title="a.key" @click="addAnnotation(a.key, a.value)">
              <el-icon><Plus /></el-icon>
              {{ a.desc }}
            </button>
          </div>
          <KVRowsTable v-model="annotations" title="注解" add-text="添加注解" empty-text="暂无注解" />
          <KVRowsTable v-model="labels" title="标签" add-text="添加标签" empty-text="暂无标签" />
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <template v-if="!yamlMode">
          <el-button v-if="current > 0" @click="current--">上一步</el-button>
          <el-button v-if="current < steps.length - 1" type="primary" @click="next">下一步</el-button>
          <el-button v-else type="primary" :loading="submitting" @click="submit">{{ isEdit ? '保存' : '创建' }}</el-button>
        </template>
        <el-button v-else type="primary" :loading="submitting" @click="submit">{{ isEdit ? '保存' : '创建' }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Check, Delete, Plus } from '@element-plus/icons-vue'
import YamlEditor from '@/components/YamlEditor.vue'
import KVRowsTable from '@/components/KVRowsTable.vue'
import { ingressApi, type IngressClassVO, type IngressFormDTO, type IngressPathDTO } from '@/api/ingress'
import { serviceApi, type ServiceVO } from '@/api/service'
import { namespaceApi } from '@/api/namespace'
import { ANNOTATIONS } from '@/utils/annotations'
import { jsonToYaml } from '@/utils/yaml'
import type { KVRow } from '@/utils/job'
import { recordToRows, rowsToRecord } from '@/utils/workload'
import { NGINX_ANNOTATION_PRESETS, PATH_TYPE_OPTIONS } from '@/utils/network'

type PathRow = IngressPathDTO
interface RuleRow {
  host: string
  https: boolean
  secretName: string
  paths: PathRow[]
}

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    clusterId: string
    defaultNamespace?: string
    editTarget?: { namespace: string; name: string } | null
  }>(),
  { defaultNamespace: '', editTarget: null }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: [namespace: string, name: string]
}>()

const steps = [
  { title: '基本信息', desc: '名称、描述与 Ingress 类' },
  { title: '路由规则', desc: '域名、路径与后端服务' },
  { title: '高级设置', desc: '注解与标签' }
]

const isEdit = computed(() => !!props.editTarget)
const title = computed(() => (isEdit.value ? '编辑应用路由' : '创建应用路由'))
const current = ref(0)
const yamlMode = ref(false)
const loading = ref(false)
const submitting = ref(false)
const basicRef = ref<FormInstance>()

const namespaces = ref<string[]>([])
const nsLoading = ref(false)
const classes = ref<IngressClassVO[]>([])
const services = ref<ServiceVO[]>([])
const svcLoading = ref(false)
const tlsSecrets = ref<string[]>([])
const secretLoading = ref(false)

const form = reactive({ namespace: '', name: '', alias: '', describe: '', ingressClassName: '' })
const createPath = (): PathRow => ({ path: '/', pathType: 'Prefix', serviceName: '', servicePort: '' })
const createRule = (): RuleRow => ({ host: '', https: false, secretName: '', paths: [createPath()] })
// rules 已被表单校验规则占用, 路由规则用 rules_
const rules_ = ref<RuleRow[]>([])
const labels = ref<KVRow[]>([])
const annotations = ref<KVRow[]>([])

const rules: FormRules = {
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { pattern: /^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$/, message: '只能包含小写字母、数字、"-" 和 "."，以字母或数字开头和结尾', trigger: 'blur' }
  ]
}

const portOptions = (serviceName: string) => {
  const svc = services.value.find(s => s.name === serviceName)
  return (svc?.ports || []).map(p => ({ value: String(p.port), label: p.name ? `${p.port}（${p.name}）` : String(p.port) }))
}

const onServiceChange = (p: PathRow) => {
  const options = portOptions(p.serviceName)
  p.servicePort = options[0]?.value || ''
}

const addAnnotation = (key: string, value: string) => {
  const exists = annotations.value.find(a => a.key === key)
  if (exists) {
    ElMessage.info('该注解已添加')
    return
  }
  annotations.value.push({ key, value })
}

const loadNamespaces = async () => {
  nsLoading.value = true
  try {
    const resp = await namespaceApi.getList({ clusterId: props.clusterId, page: 1, pageSize: 10000 })
    namespaces.value = (resp.data?.items || []).map(ns => ns.name)
  } catch {
    namespaces.value = []
  } finally {
    nsLoading.value = false
  }
}

const loadClasses = async () => {
  try {
    const resp = await ingressApi.getClasses({ clusterId: props.clusterId })
    classes.value = resp.data || []
    if (!isEdit.value && !form.ingressClassName) form.ingressClassName = classes.value.find(c => c.isDefault)?.name || (classes.value.length === 1 ? classes.value[0].name : '')
  } catch {
    classes.value = []
  }
}

const loadNamespaceOptions = async () => {
  const namespace = form.namespace
  if (!namespace) {
    services.value = []
    tlsSecrets.value = []
    return
  }
  svcLoading.value = true
  secretLoading.value = true
  const [svcs, secrets] = await Promise.allSettled([
    serviceApi.getServices({ clusterId: props.clusterId, namespace, page: 1, pageSize: 10000 }),
    ingressApi.getTlsSecrets({ clusterId: props.clusterId, namespace })
  ])
  if (namespace !== form.namespace) return
  services.value = svcs.status === 'fulfilled' ? (svcs.value.data?.items || []).filter(s => s.serviceType !== 'ExternalName' || s.ports.length) : []
  tlsSecrets.value = secrets.status === 'fulfilled' ? secrets.value.data || [] : []
  svcLoading.value = false
  secretLoading.value = false
}

const onNamespaceChange = () => {
  rules_.value.forEach(r => {
    r.secretName = ''
    r.paths.forEach(p => {
      p.serviceName = ''
      p.servicePort = ''
    })
  })
  loadNamespaceOptions()
}

const HIDDEN_ANNOTATIONS = ['alias', 'describe', ANNOTATIONS.creator, ANNOTATIONS.kubesphereCreator, ANNOTATIONS.lastApplied]

const loadForEdit = async () => {
  if (!props.editTarget) return
  loading.value = true
  try {
    const { data } = await ingressApi.getDetail({ clusterId: props.clusterId, ...props.editTarget })
    Object.assign(form, {
      namespace: data.namespace,
      name: data.name,
      alias: data.alias || '',
      describe: data.describe || '',
      ingressClassName: data.ingressClassName || ''
    })
    // 规则按域名分组, HTTPS 根据 TLS 配置中的域名判断
    const grouped = new Map<string, RuleRow>()
    data.rules.forEach(r => {
      let row = grouped.get(r.host)
      if (!row) {
        const tls = data.tlsList.find(t => (r.host ? t.hosts.includes(r.host) : !t.hosts.length))
        row = { host: r.host, https: !!tls, secretName: tls?.secretName || '', paths: [] }
        grouped.set(r.host, row)
      }
      row.paths.push({ path: r.path || '/', pathType: (r.pathType as PathRow['pathType']) || 'Prefix', serviceName: r.serviceName, servicePort: r.servicePort })
    })
    rules_.value = grouped.size ? [...grouped.values()] : [createRule()]
    labels.value = recordToRows(data.labels)
    annotations.value = recordToRows(data.annotations, HIDDEN_ANNOTATIONS)
  } catch {
    emit('update:modelValue', false)
  } finally {
    loading.value = false
  }
}

const reset = async () => {
  Object.assign(form, { namespace: props.defaultNamespace || '', name: '', alias: '', describe: '', ingressClassName: '' })
  rules_.value = [createRule()]
  labels.value = []
  annotations.value = []
  services.value = []
  tlsSecrets.value = []
  current.value = 0
  yamlMode.value = false
  loadNamespaces()
  if (props.editTarget) await loadForEdit()
  loadClasses()
  loadNamespaceOptions()
}

const HOST = /^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/

const validateRules = () => {
  const hosts = new Set<string>()
  for (let i = 0; i < rules_.value.length; i++) {
    const r = rules_.value[i]
    const label = `规则 ${i + 1}`
    const host = r.host.trim().toLowerCase()
    if (host && !HOST.test(host)) return `${label}的域名 ${r.host} 不合法`
    if (hosts.has(host)) return `域名 ${host || '（所有域名）'} 重复，请把路径放在同一条规则中`
    hosts.add(host)
    if (r.https && !r.secretName) return `请选择${label}使用的证书`
    if (r.https && !host) return `${label}开启了 HTTPS，需要填写域名`
    const paths = new Set<string>()
    for (const p of r.paths) {
      const path = p.path.trim() || '/'
      if (!path.startsWith('/')) return `${label}的路径 ${p.path} 必须以 / 开头`
      if (paths.has(path + p.pathType)) return `${label}的路径 ${path} 重复`
      paths.add(path + p.pathType)
      if (!p.serviceName) return `请选择${label}中路径 ${path} 转发到的服务`
      if (!String(p.servicePort).trim()) return `请选择${label}中路径 ${path} 的服务端口`
    }
  }
  return ''
}

const validateStep = async (step: number) => {
  if (step === 0) {
    if (isEdit.value) return true
    if (!basicRef.value || current.value !== 0) {
      const ok = !!form.namespace && /^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$/.test(form.name.trim())
      if (!ok) ElMessage.warning('请选择命名空间并填写正确的名称')
      return ok
    }
    try {
      await basicRef.value.validate()
      return true
    } catch {
      return false
    }
  }
  if (step === 1) {
    const message = validateRules()
    if (message) {
      ElMessage.warning(message)
      return false
    }
  }
  return true
}

const next = async () => {
  if (await validateStep(current.value)) current.value++
}

const goStep = async (index: number) => {
  if (index <= current.value) {
    current.value = index
    return
  }
  while (current.value < index) {
    if (!(await validateStep(current.value))) return
    current.value++
  }
}

const buildPayload = (): IngressFormDTO => {
  // 同一个证书的域名合并为一条 TLS 配置
  const tlsMap = new Map<string, string[]>()
  rules_.value.forEach(r => {
    if (!r.https || !r.secretName) return
    const hosts = tlsMap.get(r.secretName) || []
    if (r.host.trim()) hosts.push(r.host.trim().toLowerCase())
    tlsMap.set(r.secretName, hosts)
  })
  return {
    clusterId: props.clusterId,
    namespace: form.namespace,
    name: form.name.trim(),
    alias: form.alias.trim(),
    describe: form.describe.trim(),
    ingressClassName: form.ingressClassName,
    rules: rules_.value.map(r => ({
      host: r.host.trim().toLowerCase(),
      paths: r.paths.map(p => ({ path: p.path.trim() || '/', pathType: p.pathType, serviceName: p.serviceName, servicePort: String(p.servicePort).trim() }))
    })),
    tls: [...tlsMap.entries()].map(([secretName, hosts]) => ({ secretName, hosts })),
    labels: rowsToRecord(labels.value),
    annotations: rowsToRecord(annotations.value)
  }
}

const previewYaml = computed(() => {
  const p = buildPayload()
  const annotationsMap: Record<string, string> = { ...(p.annotations || {}) }
  if (p.alias) annotationsMap.alias = p.alias
  if (p.describe) annotationsMap.describe = p.describe
  const metadata: Record<string, unknown> = { name: p.name || 'example-ingress', namespace: p.namespace || 'default' }
  if (Object.keys(p.labels || {}).length) metadata.labels = p.labels
  if (Object.keys(annotationsMap).length) metadata.annotations = annotationsMap
  const spec: Record<string, unknown> = {}
  if (p.ingressClassName) spec.ingressClassName = p.ingressClassName
  if (p.tls?.length) spec.tls = p.tls.map(t => (t.hosts.length ? { hosts: t.hosts, secretName: t.secretName } : { secretName: t.secretName }))
  spec.rules = p.rules.map(r => {
    const http = {
      paths: r.paths.map(x => ({
        path: x.path,
        pathType: x.pathType,
        backend: { service: { name: x.serviceName, port: /^\d+$/.test(x.servicePort) ? { number: Number(x.servicePort) } : { name: x.servicePort } } }
      }))
    }
    return r.host ? { host: r.host, http } : { http }
  })
  return jsonToYaml(JSON.parse(JSON.stringify({ apiVersion: 'networking.k8s.io/v1', kind: 'Ingress', metadata, spec })))
})

const submit = async () => {
  for (let step = 0; step < steps.length; step++) {
    if (!(await validateStep(step))) {
      yamlMode.value = false
      current.value = step
      return
    }
  }
  submitting.value = true
  try {
    const payload = buildPayload()
    if (isEdit.value) await ingressApi.update(payload)
    else await ingressApi.create(payload)
    ElMessage.success(isEdit.value ? '保存成功' : '创建成功')
    emit('update:modelValue', false)
    emit('success', payload.namespace, payload.name)
  } catch {
    // 请求拦截器已提示错误
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.ig-option {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.ig-option em {
  margin-left: 6px;
  color: #3b74f2;
  font-style: normal;
  font-size: 12px;
}

.ig-option__desc {
  color: #94a3b8;
  font-size: 12px;
}

.ig-rule {
  margin-bottom: 16px;
  padding: 14px 18px 12px;
  border: 1px solid #e3eaf4;
  border-radius: 12px;
  background: #fbfcfe;
}

.ig-rule__head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.ig-rule__title {
  color: #1e293b;
  font-weight: 600;
}

.ig-rule__del {
  margin-left: auto;
}

.ig-paths {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  margin-left: 90px;
}

.ig-paths__head,
.ig-paths__row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ig-paths__head span {
  color: #64748b;
  font-size: 12px;
}

.ig-add {
  width: 100%;
  border-style: dashed;
}

.ig-presets {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}

.ig-preset {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 30px;
  padding: 0 12px;
  border: 1px dashed #c9d6ea;
  border-radius: 15px;
  background: #fff;
  color: #475569;
  font-size: 13px;
  cursor: pointer;
}

.ig-preset:hover {
  border-color: #3b74f2;
  color: #3b74f2;
}

.mono-input :deep(input) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
