<template>
  <el-dialog
    :model-value="modelValue"
    title="创建容器组（Pod）"
    width="1180px"
    top="4vh"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--wide"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="reset"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">创建容器组（Pod）</h3>
        <div class="dialog-actions">
          <el-switch v-model="yamlMode" active-text="YAML 预览" />
        </div>
      </div>
    </template>

    <div class="config-editor">
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
          <YamlEditor :model-value="previewYaml" :readonly="true" height="100%" title="Pod YAML 预览" :filename="`${basic.name || 'pod'}.yaml`" />
        </div>

        <!-- 1. 基本信息 -->
        <div v-else-if="current === 0" class="basic-form">
          <el-alert type="info" :closable="false" show-icon class="pc-alert">
            单独创建的容器组被删除或所在节点故障后不会自动重建，适合调试和一次性任务；长期运行的应用建议创建 Deployment / StatefulSet。
          </el-alert>
          <el-form ref="basicRef" :model="basic" :rules="rules" label-width="100px">
            <el-form-item label="命名空间" prop="namespace">
              <InfiniteSelect v-model="basic.namespace" :fetch-data="namespaceFetcher" v-bind="namespaceSelectConfig" :clearable="false" placeholder="请选择命名空间" style="width: 100%" @change="onNamespaceChange" />
            </el-form-item>
            <el-form-item label="名称" prop="name">
              <el-input v-model="basic.name" placeholder="例如 debug-tools" maxlength="253" show-word-limit />
              <div class="field-hint">小写字母、数字、"-" 和 "."，以字母或数字开头和结尾，创建后不可修改</div>
            </el-form-item>
            <el-form-item label="别名">
              <el-input v-model="basic.alias" placeholder="可选，用于列表展示" maxlength="50" show-word-limit />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="basic.describe" type="textarea" :rows="3" placeholder="可选，说明用途" maxlength="200" show-word-limit />
            </el-form-item>
          </el-form>
        </div>

        <!-- 2. 容器设置 -->
        <div v-else-if="current === 1" class="basic-form">
          <el-form label-width="100px">
            <el-form-item label="重启策略">
              <el-radio-group v-model="advanced.restartPolicy">
                <el-radio-button value="Always">总是重启（Always）</el-radio-button>
                <el-radio-button value="OnFailure">失败时重启（OnFailure）</el-radio-button>
                <el-radio-button value="Never">不重启（Never）</el-radio-button>
              </el-radio-group>
              <div class="field-hint">{{ restartHint }}</div>
            </el-form-item>
          </el-form>
          <p class="fd-subtitle">容器<small>至少需要一个工作容器</small></p>
          <el-collapse v-model="openContainer" accordion class="wl-containers">
            <el-collapse-item v-for="(c, index) in containers" :key="c.uid" :name="c.uid">
              <template #title>
                <div class="wl-container-title">
                  <span class="wl-container-title__name">{{ c.form.name || `容器 ${index + 1}` }}</span>
                  <el-tag size="small" :type="c.form.init ? 'warning' : 'primary'">{{ c.form.init ? '初始化容器' : '工作容器' }}</el-tag>
                  <span class="wl-container-title__image">{{ c.form.image || '未填写镜像' }}</span>
                  <el-button v-if="containers.length > 1" link type="danger" class="wl-container-title__del" @click.stop="containers.splice(index, 1)">
                    <el-icon><Delete /></el-icon>
                    删除
                  </el-button>
                </div>
              </template>
              <WorkloadContainerEditor :container="c.form" :volume-names="volumeNames" />
            </el-collapse-item>
          </el-collapse>
          <el-button class="wl-add" @click="addContainer">
            <el-icon><Plus /></el-icon>
            添加容器
          </el-button>
        </div>

        <!-- 3. 存储设置 -->
        <div v-else-if="current === 2" class="basic-form">
          <p class="fd-subtitle">存储卷<small>添加后在容器设置的「存储挂载」中选择挂载路径</small></p>
          <VolumeListEditor
            :volumes="volumes"
            :pvc-options="options.pvcOptions.value"
            :config-map-options="options.configMapOptions.value"
            :secret-options="options.secretOptions.value"
            :loading="options.loading.value"
          />
        </div>

        <!-- 4. 高级设置 -->
        <div v-else class="basic-form">
          <el-form :model="advanced" label-width="130px">
            <el-form-item label="调度方式">
              <el-radio-group v-model="advanced.schedule">
                <el-radio-button value="auto">自动调度</el-radio-button>
                <el-radio-button value="node">指定节点</el-radio-button>
                <el-radio-button value="selector">按节点标签</el-radio-button>
              </el-radio-group>
              <div class="field-hint">{{ scheduleHint }}</div>
            </el-form-item>
            <el-form-item v-if="advanced.schedule === 'node'" label="节点" required>
              <el-select v-model="advanced.nodeName" filterable placeholder="选择节点" style="width: 360px" :loading="nodeLoading">
                <el-option v-for="n in nodes" :key="n.name" :label="n.name" :value="n.name">
                  <div class="pc-node">
                    <span>{{ n.name }}</span>
                    <span class="pc-node__desc">{{ n.desc }}</span>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="使用节点网络">
              <el-switch v-model="advanced.hostNetwork" />
              <span class="fd-inline-hint">容器直接使用节点的网络和端口（hostNetwork），注意端口冲突</span>
            </el-form-item>
            <el-form-item label="同步主机时区">
              <el-switch v-model="advanced.syncHostTimezone" />
              <span class="fd-inline-hint">挂载节点的 /etc/localtime，使容器时间与节点一致</span>
            </el-form-item>
            <el-form-item label="服务账户">
              <el-select v-model="advanced.serviceAccountName" clearable filterable placeholder="default" style="width: 360px" :loading="options.loading.value">
                <el-option v-for="o in options.serviceAccountOptions.value" :key="o" :label="o" :value="o" />
              </el-select>
            </el-form-item>
            <el-form-item label="镜像拉取凭证">
              <el-select v-model="advanced.imagePullSecrets" multiple clearable filterable placeholder="私有镜像仓库的 Secret" style="width: 100%" :loading="options.loading.value">
                <el-option v-for="o in options.pullSecretOptions.value" :key="o" :label="o" :value="o" />
              </el-select>
            </el-form-item>
            <el-form-item label="优雅终止时间">
              <el-input-number v-model="advanced.terminationGracePeriodSeconds" :min="0" :max="3600" controls-position="right" />
              <span class="fd-inline-hint">秒，删除容器组时等待应用退出的时间</span>
            </el-form-item>
          </el-form>
          <KVRowsTable v-if="advanced.schedule === 'selector'" v-model="nodeSelector" title="节点选择器" add-text="添加节点标签" empty-text="请添加节点标签，例如 kubernetes.io/os = linux" key-placeholder="节点标签键" value-placeholder="值" />
          <KVRowsTable v-model="labels" title="标签" add-text="添加标签" empty-text="暂无标签，可添加 app 等标签方便服务选择" />
          <KVRowsTable v-model="annotations" title="注解" add-text="添加注解" empty-text="暂无注解" />
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <template v-if="!yamlMode">
          <el-button v-if="current > 0" @click="current--">上一步</el-button>
          <el-button v-if="current < steps.length - 1" type="primary" @click="next">下一步</el-button>
          <el-button v-else type="primary" :loading="submitting" @click="submit">创建</el-button>
        </template>
        <el-button v-else type="primary" :loading="submitting" @click="submit">创建</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { computed, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Check, Delete, Plus } from '@element-plus/icons-vue'
import YamlEditor from '@/components/YamlEditor.vue'
import KVRowsTable from '@/components/KVRowsTable.vue'
import WorkloadContainerEditor from '@/components/workload/WorkloadContainerEditor.vue'
import VolumeListEditor from '@/components/workload/VolumeListEditor.vue'
import { podApi, type PodCreateDTO } from '@/api/pod'
import { clusterApi } from '@/api/cluster'
import { useNamespaceOptions } from '@/composables/useNamespaceOptions'
import { jsonToYaml } from '@/utils/yaml'
import type { KVRow } from '@/utils/job'
import {
  containerPreview,
  createWorkloadContainerForm,
  rowsToRecord,
  validateWorkloadContainers,
  validateWorkloadStorage,
  volumePreview,
  volumeToDTO,
  withHostTime,
  workloadContainerToDTO,
  type VolumeForm,
  type WorkloadContainerForm
} from '@/utils/workload'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    clusterId: string
    defaultNamespace?: string
  }>(),
  { defaultNamespace: '' }
)
// 命名空间下拉框分页加载
const namespaceFetcher = computed(() => useNamespaceFetcher(props.clusterId))

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: [namespace: string, name: string]
}>()

const steps = [
  { title: '基本信息', desc: '命名空间、名称与描述' },
  { title: '容器设置', desc: '重启策略与容器' },
  { title: '存储设置', desc: '存储卷与挂载' },
  { title: '高级设置', desc: '调度、网络与标签' }
]

const NAME = /^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$/

const current = ref(0)
const yamlMode = ref(false)
const submitting = ref(false)
const basicRef = ref<FormInstance>()

const nodes = ref<{ name: string; desc: string }[]>([])
const nodeLoading = ref(false)
const options = useNamespaceOptions(() => props.clusterId)

const basic = reactive({ namespace: '', name: '', alias: '', describe: '' })
const createAdvanced = () => ({
  restartPolicy: 'Always' as 'Always' | 'OnFailure' | 'Never',
  schedule: 'auto' as 'auto' | 'node' | 'selector',
  nodeName: '',
  hostNetwork: false,
  syncHostTimezone: true,
  serviceAccountName: '',
  imagePullSecrets: [] as string[],
  terminationGracePeriodSeconds: 30
})
const advanced = reactive(createAdvanced())

let uid = 0
const containers = ref<{ uid: number; form: WorkloadContainerForm }[]>([])
const openContainer = ref<number | string>('')
const volumes = ref<VolumeForm[]>([])
const labels = ref<KVRow[]>([])
const annotations = ref<KVRow[]>([])
const nodeSelector = ref<KVRow[]>([])

const volumeNames = computed(() => volumes.value.map(v => v.name.trim()).filter(Boolean))

const restartHint = computed(
  () =>
    ({
      Always: '容器退出后总是重启，适合长期运行的服务',
      OnFailure: '容器异常退出（退出码不为 0）时重启，正常结束后不再运行',
      Never: '容器退出后不重启，适合一次性的调试或脚本'
    })[advanced.restartPolicy]
)

const scheduleHint = computed(
  () =>
    ({
      auto: '由调度器根据资源情况选择节点',
      node: '直接运行在指定节点上，不经过调度器（不检查资源是否充足）',
      selector: '只调度到带有指定标签的节点'
    })[advanced.schedule]
)

const rules: FormRules = {
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { pattern: NAME, message: '只能包含小写字母、数字、"-" 和 "."，以字母或数字开头和结尾', trigger: 'blur' }
  ]
}

const addContainer = () => {
  const item = { uid: ++uid, form: createWorkloadContainerForm(containers.value.length) }
  containers.value.push(item)
  openContainer.value = item.uid
}

const loadNodes = async () => {
  nodeLoading.value = true
  try {
    const resp = await clusterApi.getNodes(props.clusterId, { page: 1, pageSize: 10000 })
    nodes.value = (resp.data?.items || []).map((n: any) => ({
      name: n.hostname || n.name,
      desc: [n.status, n.unSchedule || n.unschedulable ? '不可调度' : '', n.internalIP || n.ip || ''].filter(Boolean).join(' · ')
    }))
  } catch {
    nodes.value = []
  } finally {
    nodeLoading.value = false
  }
}

const onNamespaceChange = () => {
  volumes.value.forEach(v => {
    if (v.type !== 'emptyDir' && v.type !== 'hostPath') v.source = ''
  })
  advanced.serviceAccountName = ''
  advanced.imagePullSecrets = []
  options.load(basic.namespace)
}

const reset = () => {
  Object.assign(basic, { namespace: props.defaultNamespace || '', name: '', alias: '', describe: '' })
  Object.assign(advanced, createAdvanced())
  containers.value = []
  addContainer()
  volumes.value = []
  labels.value = []
  annotations.value = []
  nodeSelector.value = []
  current.value = 0
  yamlMode.value = false
  loadNodes()
  options.load(basic.namespace)
}

const validateAdvanced = () => {
  if (advanced.schedule === 'node' && !advanced.nodeName) return '请选择节点'
  if (advanced.schedule === 'selector' && !Object.keys(rowsToRecord(nodeSelector.value)).length) return '请至少添加一个节点标签'
  return ''
}

const validateStep = async (step: number) => {
  if (step === 0) {
    if (!basicRef.value || current.value !== 0) {
      const ok = !!basic.namespace && NAME.test(basic.name.trim())
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
  const forms = containers.value.map(c => c.form)
  const message =
    step === 1 ? validateWorkloadContainers(forms) : step === 2 ? validateWorkloadStorage(volumes.value, [], forms) : validateAdvanced()
  if (message) {
    ElMessage.warning(message)
    return false
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

const buildPayload = (): PodCreateDTO => ({
  clusterId: props.clusterId,
  namespace: basic.namespace,
  name: basic.name.trim(),
  alias: basic.alias.trim(),
  describe: basic.describe.trim(),
  labels: rowsToRecord(labels.value),
  annotations: rowsToRecord(annotations.value),
  containers: containers.value.map(c => workloadContainerToDTO(c.form)),
  volumes: volumes.value.map(volumeToDTO),
  restartPolicy: advanced.restartPolicy,
  nodeName: advanced.schedule === 'node' ? advanced.nodeName : '',
  nodeSelector: advanced.schedule === 'selector' ? rowsToRecord(nodeSelector.value) : {},
  serviceAccountName: advanced.serviceAccountName,
  imagePullSecrets: advanced.imagePullSecrets,
  syncHostTimezone: advanced.syncHostTimezone,
  terminationGracePeriodSeconds: advanced.terminationGracePeriodSeconds,
  hostNetwork: advanced.hostNetwork
})

// 与后端生成的对象保持一致
const previewYaml = computed(() => {
  const p = buildPayload()
  const podSpec: Record<string, unknown> = {}
  const initContainers = p.containers.filter(c => c.init).map(containerPreview)
  if (initContainers.length) podSpec.initContainers = initContainers
  podSpec.containers = p.containers.filter(c => !c.init).map(containerPreview)
  const vols = (p.volumes || []).map(volumePreview)
  if (vols.length) podSpec.volumes = vols
  if (p.syncHostTimezone) withHostTime(podSpec)
  podSpec.restartPolicy = p.restartPolicy
  if (p.nodeName) podSpec.nodeName = p.nodeName
  if (Object.keys(p.nodeSelector || {}).length) podSpec.nodeSelector = p.nodeSelector
  if (p.hostNetwork) {
    podSpec.hostNetwork = true
    podSpec.dnsPolicy = 'ClusterFirstWithHostNet'
  }
  if (p.serviceAccountName) podSpec.serviceAccountName = p.serviceAccountName
  if (p.imagePullSecrets?.length) podSpec.imagePullSecrets = p.imagePullSecrets.map(n => ({ name: n }))
  podSpec.terminationGracePeriodSeconds = p.terminationGracePeriodSeconds
  const annotationsMap: Record<string, string> = { ...(p.annotations || {}) }
  if (p.alias) annotationsMap.alias = p.alias
  if (p.describe) annotationsMap.describe = p.describe
  const metadata: Record<string, unknown> = { name: p.name || 'example-pod', namespace: p.namespace || 'default' }
  if (Object.keys(p.labels || {}).length) metadata.labels = p.labels
  if (Object.keys(annotationsMap).length) metadata.annotations = annotationsMap
  return jsonToYaml(JSON.parse(JSON.stringify({ apiVersion: 'v1', kind: 'Pod', metadata, spec: podSpec })))
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
    await podApi.create(payload)
    ElMessage.success('创建成功')
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
.pc-alert {
  margin-bottom: 18px;
}

.pc-node {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.pc-node__desc {
  color: #94a3b8;
  font-size: 12px;
}

.wl-containers {
  margin-bottom: 12px;
  border: 1px solid #e3eaf4;
  border-radius: 12px;
  overflow: hidden;
}

.wl-containers :deep(.el-collapse-item__header) {
  height: 52px;
  padding: 0 16px;
  background: #fbfcfe;
}

.wl-containers :deep(.el-collapse-item__content) {
  padding: 18px 18px 4px;
}

.wl-container-title {
  display: flex;
  flex: 1;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.wl-container-title__name {
  color: #1e293b;
  font-weight: 600;
}

.wl-container-title__image {
  overflow: hidden;
  color: #94a3b8;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12px;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.wl-container-title__del {
  margin: 0 12px 0 auto;
}

.wl-add {
  width: 100%;
  margin-bottom: 18px;
  border-style: dashed;
}
</style>
