<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="1180px"
    top="4vh"
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
          <YamlEditor :model-value="previewYaml" :readonly="true" height="100%" :title="`${meta.label} YAML 预览`" :filename="`${basic.name || kind}.yaml`" />
        </div>

        <!-- 1. 基本信息 -->
        <div v-else-if="current === 0" class="basic-form">
          <el-form ref="basicRef" :model="basic" :rules="rules" label-width="100px">
            <el-form-item label="命名空间" prop="namespace">
              <el-select v-model="basic.namespace" filterable placeholder="请选择命名空间" :loading="nsLoading" style="width: 100%" @change="onNamespaceChange">
                <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
            <el-form-item label="名称" prop="name">
              <el-input v-model="basic.name" placeholder="例如 web-frontend" :maxlength="nameMax" show-word-limit />
              <div class="field-hint">小写字母、数字和 "-"，以字母或数字开头和结尾，创建后不可修改</div>
            </el-form-item>
            <el-form-item label="别名">
              <el-input v-model="basic.alias" placeholder="可选，用于列表展示" maxlength="50" show-word-limit />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="basic.describe" type="textarea" :rows="3" placeholder="可选，说明用途" maxlength="200" show-word-limit />
            </el-form-item>
          </el-form>
        </div>

        <!-- 2. 容器组设置 -->
        <div v-else-if="current === 1" class="basic-form">
          <el-form :model="pod" label-width="130px">
            <el-form-item v-if="kind !== 'daemonset'" label="容器组副本数">
              <el-input-number v-model="pod.replicas" :min="0" :max="1000" controls-position="right" />
              <span class="fd-inline-hint">同时运行的容器组数量</span>
            </el-form-item>
            <el-form-item v-else label="运行方式">
              <span class="wc-note">在每个符合条件的节点上运行一个容器组，可在「高级设置」中用节点选择器限定节点</span>
            </el-form-item>
            <el-form-item label="更新策略">
              <el-radio-group v-model="pod.strategy">
                <el-radio-button value="RollingUpdate">滚动更新</el-radio-button>
                <el-radio-button v-if="kind === 'deployment'" value="Recreate">重新创建</el-radio-button>
                <el-radio-button v-else value="OnDelete">删除容器组时更新</el-radio-button>
              </el-radio-group>
              <div class="field-hint">{{ strategyHint }}</div>
            </el-form-item>
            <template v-if="pod.strategy === 'RollingUpdate'">
              <el-form-item v-if="kind === 'deployment'" label="最大超出副本数">
                <el-input v-model="pod.maxSurge" placeholder="25%" style="width: 160px" />
                <span class="fd-inline-hint">更新时最多比期望副本数多出的容器组，数字或百分比</span>
              </el-form-item>
              <el-form-item v-if="kind !== 'statefulset'" label="最大不可用数">
                <el-input v-model="pod.maxUnavailable" :placeholder="kind === 'deployment' ? '25%' : '1'" style="width: 160px" />
                <span class="fd-inline-hint">更新时最多允许不可用的容器组，数字或百分比</span>
              </el-form-item>
              <el-form-item v-if="kind === 'statefulset'" label="分批更新序号">
                <el-input-number v-model="pod.partition" :min="0" :max="1000" controls-position="right" />
                <span class="fd-inline-hint">只更新序号不小于该值的容器组，0 表示全部更新</span>
              </el-form-item>
            </template>
            <el-form-item v-if="kind === 'statefulset'" label="容器组管理策略">
              <el-radio-group v-model="pod.podManagementPolicy">
                <el-radio-button value="OrderedReady">按顺序（OrderedReady）</el-radio-button>
                <el-radio-button value="Parallel">并行（Parallel）</el-radio-button>
              </el-radio-group>
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
          <template v-if="kind === 'statefulset'">
            <p class="fd-subtitle">存储卷模板<small>为每个容器组自动创建独立的持久卷声明，容器组重建后数据保留</small></p>
            <div v-for="(t, i) in claims" :key="'claim' + i" class="wl-volume">
              <div class="wl-volume__head">
                <span class="wl-volume__title">{{ t.name || `模板 ${i + 1}` }}</span>
                <el-button link type="danger" @click="claims.splice(i, 1)"><el-icon><Delete /></el-icon>删除</el-button>
              </div>
              <el-form :model="t" label-width="90px" class="wl-volume__form">
                <el-form-item label="名称"><el-input v-model="t.name" placeholder="例如 data" /></el-form-item>
                <el-form-item label="存储类">
                  <el-select v-model="t.storageClassName" clearable placeholder="使用默认存储类" style="width: 100%">
                    <el-option v-for="sc in storageClasses" :key="sc.name" :label="sc.isDefault ? `${sc.name}（默认）` : sc.name" :value="sc.name" />
                  </el-select>
                </el-form-item>
                <el-form-item label="访问模式">
                  <el-select v-model="t.accessMode" style="width: 100%">
                    <el-option v-for="m in ACCESS_MODE_OPTIONS" :key="m.value" :label="`${m.label}（${m.desc}）`" :value="m.value" />
                  </el-select>
                </el-form-item>
                <el-form-item label="容量">
                  <div class="fd-capacity">
                    <el-input-number v-model="t.capacityValue" :min="1" :max="1048576" controls-position="right" />
                    <el-select v-model="t.capacityUnit">
                      <el-option v-for="u in CAPACITY_UNITS" :key="u" :label="u" :value="u" />
                    </el-select>
                  </div>
                </el-form-item>
              </el-form>
            </div>
            <el-button class="wl-add" @click="claims.push(createClaimTemplateForm(claims.length))">
              <el-icon><Plus /></el-icon>
              添加存储卷模板
            </el-button>
          </template>

          <p class="fd-subtitle">存储卷<small>添加后在容器设置的「存储挂载」中选择挂载路径</small></p>
          <VolumeListEditor
            :volumes="volumes"
            :pvc-options="pvcOptions"
            :config-map-options="configMapOptions"
            :secret-options="secretOptions"
            :loading="optionsLoading"
          />
        </div>

        <!-- 4. 高级设置 -->
        <div v-else class="basic-form">
          <el-form :model="advanced" label-width="130px">
            <el-form-item label="同步主机时区">
              <el-switch v-model="advanced.syncHostTimezone" />
              <span class="fd-inline-hint">挂载节点的 /etc/localtime，使容器时间与节点一致</span>
            </el-form-item>
            <el-form-item label="服务账户">
              <el-select v-model="advanced.serviceAccountName" clearable filterable placeholder="default" style="width: 320px" :loading="optionsLoading">
                <el-option v-for="o in serviceAccountOptions" :key="o" :label="o" :value="o" />
              </el-select>
            </el-form-item>
            <el-form-item label="镜像拉取凭证">
              <el-select v-model="advanced.imagePullSecrets" multiple clearable filterable placeholder="私有镜像仓库的 Secret" style="width: 100%" :loading="optionsLoading">
                <el-option v-for="o in pullSecretOptions" :key="o" :label="o" :value="o" />
              </el-select>
            </el-form-item>
            <el-form-item label="优雅终止时间">
              <el-input-number v-model="advanced.terminationGracePeriodSeconds" :min="0" :max="3600" controls-position="right" />
              <span class="fd-inline-hint">秒，删除容器组时等待应用退出的时间</span>
            </el-form-item>
            <el-form-item label="最短就绪时间">
              <el-input-number v-model="pod.minReadySeconds" :min="0" :max="3600" controls-position="right" />
              <span class="fd-inline-hint">秒，容器组就绪后保持这段时间才算可用</span>
            </el-form-item>
            <el-form-item label="保留历史版本">
              <el-input-number v-model="pod.revisionHistoryLimit" :min="0" :max="100" controls-position="right" />
              <span class="fd-inline-hint">个，用于版本回退</span>
            </el-form-item>
            <el-form-item v-if="kind === 'statefulset'" label="Headless 服务">
              <el-input v-model="advanced.stsServiceName" :placeholder="`留空时自动创建 ${basic.name || '同名'} 服务`" style="width: 320px" />
              <div class="field-hint">为每个容器组提供固定域名 &lt;容器组名&gt;.&lt;服务名&gt;.&lt;命名空间&gt;.svc</div>
            </el-form-item>
            <el-form-item v-else label="同时创建服务">
              <el-radio-group v-model="advanced.serviceType">
                <el-radio-button value="">不创建</el-radio-button>
                <el-radio-button value="ClusterIP">集群内访问（ClusterIP）</el-radio-button>
                <el-radio-button value="NodePort">节点端口（NodePort）</el-radio-button>
              </el-radio-group>
              <div class="field-hint">
                {{ advanced.serviceType ? `创建与工作负载同名的服务，端口为容器端口（${containerPortText || '请先设置容器端口'}）` : '也可以稍后在「Service管理」中创建' }}
              </div>
            </el-form-item>
          </el-form>
          <KVRowsTable v-model="labels" title="标签" add-text="添加标签" empty-text="不填写时使用 app: 名称" />
          <div class="field-hint wl-hint">标签同时作为容器组标签和选择器，服务通过它找到容器组，创建后不能修改选择器</div>
          <KVRowsTable v-model="nodeSelector" title="节点选择器" add-text="添加节点标签" empty-text="不限制节点" key-placeholder="节点标签键" value-placeholder="值" />
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
import { computed, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Check, Delete, Plus } from '@element-plus/icons-vue'
import YamlEditor from '@/components/YamlEditor.vue'
import KVRowsTable from '@/components/KVRowsTable.vue'
import WorkloadContainerEditor from '@/components/workload/WorkloadContainerEditor.vue'
import VolumeListEditor from '@/components/workload/VolumeListEditor.vue'
import { workloadApi, type WorkloadCreateDTO, type WorkloadKind } from '@/api/workload'
import { namespaceApi } from '@/api/namespace'
import { configMapApi } from '@/api/configmap'
import { secretApi } from '@/api/secret'
import { pvcApi } from '@/api/pvc'
import { serviceAccountApi } from '@/api/serviceaccount'
import { storageClassApi, type StorageClassVO } from '@/api/storageclass'
import { jsonToYaml } from '@/utils/yaml'
import { ACCESS_MODE_OPTIONS, CAPACITY_UNITS } from '@/utils/storageclass'
import { DNS_LABEL, type KVRow } from '@/utils/job'
import {
  WORKLOAD_KINDS,
  claimToDTO,
  containerPreview,
  createClaimTemplateForm,
  createWorkloadContainerForm,
  rowsToRecord,
  validateWorkloadContainers,
  validateWorkloadStorage,
  volumePreview,
  volumeToDTO,
  workloadContainerToDTO,
  type ClaimTemplateForm,
  type VolumeForm,
  type WorkloadContainerForm
} from '@/utils/workload'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    clusterId: string
    kind: WorkloadKind
    defaultNamespace?: string
  }>(),
  { defaultNamespace: '' }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: [namespace: string, name: string]
}>()

const meta = computed(() => WORKLOAD_KINDS[props.kind])
const title = computed(() => `创建${meta.value.cn}（${meta.value.label}）`)
// StatefulSet 名称最长 52, 控制器会在名称后追加版本哈希
const nameMax = computed(() => (props.kind === 'statefulset' ? 52 : 63))

const steps = [
  { title: '基本信息', desc: '命名空间、名称与描述' },
  { title: '容器组设置', desc: '副本、更新策略与容器' },
  { title: '存储设置', desc: '存储卷与挂载' },
  { title: '高级设置', desc: '标签、调度与服务' }
]

const current = ref(0)
const yamlMode = ref(false)
const submitting = ref(false)
const basicRef = ref<FormInstance>()

const namespaces = ref<string[]>([])
const nsLoading = ref(false)
const optionsLoading = ref(false)
const pvcOptions = ref<string[]>([])
const configMapOptions = ref<string[]>([])
const secretOptions = ref<string[]>([])
const pullSecretOptions = ref<string[]>([])
const serviceAccountOptions = ref<string[]>([])
const storageClasses = ref<StorageClassVO[]>([])

const basic = reactive({ namespace: '', name: '', alias: '', describe: '' })
const createPodForm = () => ({
  replicas: 1,
  strategy: 'RollingUpdate',
  maxSurge: '',
  maxUnavailable: '',
  partition: 0,
  podManagementPolicy: 'OrderedReady',
  minReadySeconds: 0,
  revisionHistoryLimit: 10
})
const pod = reactive(createPodForm())
const createAdvancedForm = () => ({
  syncHostTimezone: true,
  serviceAccountName: '',
  imagePullSecrets: [] as string[],
  terminationGracePeriodSeconds: 30,
  stsServiceName: '',
  serviceType: '' as '' | 'ClusterIP' | 'NodePort'
})
const advanced = reactive(createAdvancedForm())

let uid = 0
interface ContainerItem {
  uid: number
  form: WorkloadContainerForm
}
const containers = ref<ContainerItem[]>([])
const openContainer = ref<number | string>('')
const volumes = ref<VolumeForm[]>([])
const claims = ref<ClaimTemplateForm[]>([])
const labels = ref<KVRow[]>([])
const nodeSelector = ref<KVRow[]>([])
const annotations = ref<KVRow[]>([])

const volumeNames = computed(() => [...claims.value.map(c => c.name.trim()), ...volumes.value.map(v => v.name.trim())].filter(Boolean))

const strategyHint = computed(() => {
  if (pod.strategy === 'Recreate') return '先删除所有旧容器组再创建新的，更新期间服务会中断'
  if (pod.strategy === 'OnDelete') return '更新模板后不会自动替换容器组，手动删除容器组时才使用新模板创建'
  if (props.kind === 'statefulset') return '按序号从大到小逐个替换容器组'
  return '逐步用新容器组替换旧容器组，更新期间服务不中断'
})

const containerPortText = computed(() =>
  containers.value
    .filter(c => !c.form.init)
    .flatMap(c => c.form.ports.filter(p => p.containerPort).map(p => `${p.containerPort}/${p.protocol}`))
    .join('、')
)

const rules = computed<FormRules>(() => ({
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    {
      validator: (_rule, value: string, callback) => {
        const v = (value || '').trim()
        if (!DNS_LABEL.test(v)) callback(new Error('只能包含小写字母、数字和 "-"，且以字母或数字开头和结尾'))
        else if (v.length > nameMax.value) callback(new Error(`名称不能超过 ${nameMax.value} 个字符`))
        else callback()
      },
      trigger: 'blur'
    }
  ]
}))

const addContainer = () => {
  const item = { uid: ++uid, form: createWorkloadContainerForm(containers.value.length) }
  containers.value.push(item)
  openContainer.value = item.uid
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

// 命名空间内的可选资源: 持久卷声明、ConfigMap、Secret、服务账户
const loadNamespaceOptions = async () => {
  const namespace = basic.namespace
  if (!namespace) return
  optionsLoading.value = true
  const params = { clusterId: props.clusterId, namespace, page: 1, pageSize: 10000 }
  const names = (items?: { name: string }[]) => (items || []).map(i => i.name).sort()
  const [pvcs, cms, secrets, sas] = await Promise.allSettled([
    pvcApi.getPVCList(params),
    configMapApi.getList(params),
    secretApi.getList(params),
    serviceAccountApi.getList(params)
  ])
  if (namespace !== basic.namespace) return
  pvcOptions.value = pvcs.status === 'fulfilled' ? names(pvcs.value.data?.items) : []
  configMapOptions.value = cms.status === 'fulfilled' ? names(cms.value.data?.items) : []
  if (secrets.status === 'fulfilled') {
    const items = secrets.value.data?.items || []
    secretOptions.value = names(items)
    pullSecretOptions.value = names(items.filter(s => !s.type || s.type === 'kubernetes.io/dockerconfigjson' || s.type === 'kubernetes.io/dockercfg'))
  } else {
    secretOptions.value = []
    pullSecretOptions.value = []
  }
  serviceAccountOptions.value = sas.status === 'fulfilled' ? names(sas.value.data?.items) : []
  optionsLoading.value = false
}

const onNamespaceChange = () => {
  volumes.value.forEach(v => {
    if (v.type !== 'emptyDir' && v.type !== 'hostPath') v.source = ''
  })
  advanced.serviceAccountName = ''
  advanced.imagePullSecrets = []
  loadNamespaceOptions()
}

const loadStorageClasses = async () => {
  try {
    const resp = await storageClassApi.getList({ clusterId: props.clusterId, page: 1, pageSize: 1000 })
    storageClasses.value = resp.data?.items || []
  } catch {
    storageClasses.value = []
  }
}

const reset = () => {
  Object.assign(basic, { namespace: props.defaultNamespace || '', name: '', alias: '', describe: '' })
  Object.assign(pod, createPodForm())
  Object.assign(advanced, createAdvancedForm())
  containers.value = []
  addContainer()
  volumes.value = []
  claims.value = []
  labels.value = []
  nodeSelector.value = []
  annotations.value = []
  current.value = 0
  yamlMode.value = false
  loadNamespaces()
  if (props.kind === 'statefulset') loadStorageClasses()
  if (basic.namespace) loadNamespaceOptions()
}

const PERCENT_OR_INT = /^(\d+%?)?$/

const validateContainers = () => {
  const message = validateWorkloadContainers(containers.value.map(c => c.form))
  if (message) return message
  if (props.kind === 'deployment' && pod.strategy === 'RollingUpdate') {
    if (!PERCENT_OR_INT.test(pod.maxSurge.trim()) || !PERCENT_OR_INT.test(pod.maxUnavailable.trim())) return '最大超出副本数和最大不可用数需要是数字或百分比，例如 25% 或 1'
  }
  if (props.kind === 'daemonset' && pod.strategy === 'RollingUpdate' && !PERCENT_OR_INT.test(pod.maxUnavailable.trim())) return '最大不可用数需要是数字或百分比'
  return ''
}

const validateStorage = () =>
  validateWorkloadStorage(
    volumes.value,
    claims.value,
    containers.value.map(c => c.form)
  )

const validateAdvanced = () => {
  if (advanced.serviceType && !containerPortText.value) return '同时创建服务需要容器设置端口，请在容器设置中添加端口或选择「不创建」'
  if (props.kind === 'statefulset' && advanced.stsServiceName.trim() && !/^[a-z]([-a-z0-9]*[a-z0-9])?$/.test(advanced.stsServiceName.trim())) return 'Headless 服务名称需要以小写字母开头'
  return ''
}

const validateStep = async (step: number) => {
  if (step === 0) {
    if (!basicRef.value || current.value !== 0) {
      const ok = !!basic.namespace && DNS_LABEL.test(basic.name.trim()) && basic.name.trim().length <= nameMax.value
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
  // 挂载引用存储卷, 第二步只检查容器本身, 挂载在第三步检查
  const message = step === 1 ? validateContainers() : step === 2 ? validateStorage() : validateAdvanced()
  if (message) {
    ElMessage.warning(message)
    return false
  }
  return true
}

const next = async () => {
  if (await validateStep(current.value)) {
    current.value++
    if (current.value === 2 && basic.namespace && !configMapOptions.value.length && !optionsLoading.value) loadNamespaceOptions()
  }
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

const buildPayload = (): WorkloadCreateDTO => {
  const labelMap = rowsToRecord(labels.value)
  const payload: WorkloadCreateDTO = {
    clusterId: props.clusterId,
    kind: props.kind,
    namespace: basic.namespace,
    name: basic.name.trim(),
    alias: basic.alias.trim(),
    describe: basic.describe.trim(),
    labels: labelMap,
    annotations: rowsToRecord(annotations.value),
    containers: containers.value.map(c => workloadContainerToDTO(c.form)),
    volumes: volumes.value.map(volumeToDTO),
    strategy: pod.strategy,
    minReadySeconds: pod.minReadySeconds,
    revisionHistoryLimit: pod.revisionHistoryLimit,
    nodeSelector: rowsToRecord(nodeSelector.value),
    serviceAccountName: advanced.serviceAccountName,
    imagePullSecrets: advanced.imagePullSecrets,
    syncHostTimezone: advanced.syncHostTimezone,
    terminationGracePeriodSeconds: advanced.terminationGracePeriodSeconds
  }
  if (props.kind !== 'daemonset') payload.replicas = pod.replicas
  if (pod.strategy === 'RollingUpdate') {
    if (props.kind === 'deployment') payload.maxSurge = pod.maxSurge.trim()
    if (props.kind !== 'statefulset') payload.maxUnavailable = pod.maxUnavailable.trim()
    if (props.kind === 'statefulset') payload.partition = pod.partition
  }
  if (props.kind === 'statefulset') {
    payload.volumeClaimTemplates = claims.value.map(claimToDTO)
    payload.podManagementPolicy = pod.podManagementPolicy
    payload.serviceName = advanced.stsServiceName.trim()
  } else if (advanced.serviceType) {
    payload.service = { type: advanced.serviceType }
  }
  return payload
}

// 与后端生成的对象保持一致, 方便确认
const previewYaml = computed(() => {
  const p = buildPayload()
  const name = p.name || `example-${props.kind}`
  const podLabels = Object.keys(p.labels || {}).length ? (p.labels as Record<string, string>) : { app: name }
  const workContainers = p.containers.filter(c => !c.init)
  const initContainers = p.containers.filter(c => c.init)
  const vols = (p.volumes || []).map(volumePreview)
  const mount = (list: Record<string, unknown>[]) =>
    p.syncHostTimezone ? list.map(c => ({ ...c, volumeMounts: [...((c.volumeMounts as unknown[]) || []), { name: 'host-time', mountPath: '/etc/localtime', readOnly: true }] })) : list
  if (p.syncHostTimezone) vols.push({ name: 'host-time', hostPath: { path: '/etc/localtime' } })
  const podSpec: Record<string, unknown> = {}
  if (initContainers.length) podSpec.initContainers = mount(initContainers.map(containerPreview))
  podSpec.containers = mount(workContainers.map(containerPreview))
  if (vols.length) podSpec.volumes = vols
  if (Object.keys(p.nodeSelector || {}).length) podSpec.nodeSelector = p.nodeSelector
  if (p.serviceAccountName) podSpec.serviceAccountName = p.serviceAccountName
  if (p.imagePullSecrets?.length) podSpec.imagePullSecrets = p.imagePullSecrets.map(n => ({ name: n }))
  podSpec.terminationGracePeriodSeconds = p.terminationGracePeriodSeconds
  const annotationsMap: Record<string, string> = { ...(p.annotations || {}) }
  if (p.alias) annotationsMap.alias = p.alias
  if (p.describe) annotationsMap.describe = p.describe
  const metadata: Record<string, unknown> = { name, namespace: p.namespace || 'default', labels: podLabels }
  if (Object.keys(annotationsMap).length) metadata.annotations = annotationsMap
  const spec: Record<string, unknown> = {}
  if (props.kind !== 'daemonset') spec.replicas = p.replicas
  if (props.kind === 'statefulset') spec.serviceName = p.serviceName || name
  spec.selector = { matchLabels: podLabels }
  if (props.kind === 'deployment') {
    spec.strategy =
      p.strategy === 'Recreate'
        ? { type: 'Recreate' }
        : { type: 'RollingUpdate', rollingUpdate: { maxSurge: p.maxSurge || '25%', maxUnavailable: p.maxUnavailable || '25%' } }
  } else if (props.kind === 'statefulset') {
    spec.podManagementPolicy = p.podManagementPolicy
    spec.updateStrategy = p.strategy === 'OnDelete' ? { type: 'OnDelete' } : p.partition ? { type: 'RollingUpdate', rollingUpdate: { partition: p.partition } } : { type: 'RollingUpdate' }
  } else {
    spec.updateStrategy = p.strategy === 'OnDelete' ? { type: 'OnDelete' } : { type: 'RollingUpdate', rollingUpdate: { maxUnavailable: p.maxUnavailable || 1 } }
  }
  if (p.minReadySeconds) spec.minReadySeconds = p.minReadySeconds
  spec.revisionHistoryLimit = p.revisionHistoryLimit
  spec.template = { metadata: { labels: podLabels }, spec: podSpec }
  if (p.volumeClaimTemplates?.length) {
    spec.volumeClaimTemplates = p.volumeClaimTemplates.map(t => ({
      metadata: { name: t.name },
      spec: { ...(t.storageClassName ? { storageClassName: t.storageClassName } : {}), accessModes: t.accessModes, resources: { requests: { storage: t.capacity } } }
    }))
  }
  const docs = [jsonToYaml(JSON.parse(JSON.stringify({ apiVersion: 'apps/v1', kind: meta.value.label, metadata, spec })))]
  const serviceType = props.kind === 'statefulset' ? (p.serviceName ? '' : 'Headless') : p.service?.type || ''
  if (serviceType) {
    const ports = workContainers.flatMap(c => (c.ports || []).map(port => ({ name: port.name || `${(port.protocol || 'TCP').toLowerCase()}-${port.containerPort}`, protocol: port.protocol || 'TCP', port: port.containerPort, targetPort: port.containerPort })))
    const svcSpec: Record<string, unknown> = serviceType === 'Headless' ? { clusterIP: 'None' } : { type: serviceType }
    svcSpec.selector = podLabels
    if (ports.length) svcSpec.ports = ports
    docs.push(jsonToYaml({ apiVersion: 'v1', kind: 'Service', metadata: { name: p.serviceName || name, namespace: p.namespace || 'default', labels: podLabels }, spec: svcSpec }))
  }
  return docs.join('\n---\n')
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
    await workloadApi.create(payload)
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

.wl-volume {
  margin-bottom: 14px;
  padding: 14px 16px 2px;
  border: 1px solid #e3eaf4;
  border-radius: 12px;
  background: #fbfcfe;
}

.wl-volume__head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.wl-volume__head .el-button {
  margin-left: auto;
}

.wl-volume__title {
  color: #1e293b;
  font-weight: 600;
}

.wl-volume__form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.wl-volume-types {
  grid-template-columns: repeat(5, minmax(0, 1fr));
  margin-bottom: 14px;
}

.wl-volume-types .fd-choice {
  padding: 10px 12px;
}

.wc-note {
  color: #64748b;
}

.wl-hint {
  margin: -10px 0 18px;
}

.mono-input :deep(input) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
