<template>
  <el-dialog
    :model-value="modelValue"
    title="创建持久卷"
    width="1080px"
    top="6vh"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--wide"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="reset"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">创建持久卷</h3>
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
          <YamlEditor :model-value="previewYaml" :readonly="true" height="100%" title="持久卷 YAML 预览" :filename="`${form.name || 'pv'}.yaml`" />
        </div>

        <!-- 第 1 步: 基本信息 -->
        <div v-else-if="current === 0" class="basic-form">
          <el-form ref="basicRef" :model="form" :rules="rules" label-width="100px" label-position="right">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="例如 nfs-pv-01" maxlength="253" show-word-limit />
              <div class="field-hint">小写字母、数字、"-" 和 "."，创建后不可修改</div>
            </el-form-item>
            <el-form-item label="别名">
              <el-input v-model="form.alias" placeholder="可选，用于列表展示" maxlength="50" show-word-limit />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="form.describe" type="textarea" :rows="3" placeholder="可选，说明用途" maxlength="200" show-word-limit />
            </el-form-item>
          </el-form>
          <KVRowsTable v-model="form.labels" title="标签" add-text="添加标签" empty-text="暂无标签，可用于 PVC 的 selector 精确匹配持久卷" />
        </div>

        <!-- 第 2 步: 存储来源 -->
        <div v-else-if="current === 1" class="basic-form">
          <p class="fd-subtitle">存储来源<small>持久卷指向的实际存储</small></p>
          <div class="fd-choice-grid">
            <button
              v-for="src in SOURCES"
              :key="src.value"
              type="button"
              :class="['fd-choice', { 'is-active': form.sourceType === src.value }]"
              @click="selectSource(src.value)"
            >
              <span class="fd-choice__icon" :style="{ background: src.color }">{{ src.short }}</span>
              <span class="fd-choice__body">
                <span class="fd-choice__title"><span>{{ src.label }}</span></span>
                <span class="fd-choice__desc">{{ src.desc }}</span>
              </span>
              <el-icon class="fd-choice__check"><CircleCheckFilled /></el-icon>
            </button>
          </div>

          <el-form :model="form" label-width="100px" label-position="right">
            <template v-if="form.sourceType === 'nfs'">
              <el-form-item label="服务器地址" required>
                <el-input v-model="form.nfs.server" placeholder="例如 192.168.1.100 或 nfs.example.com" />
              </el-form-item>
              <el-form-item label="共享路径" required>
                <el-input v-model="form.nfs.path" placeholder="例如 /data/nfs/pv-01" />
              </el-form-item>
              <el-form-item label="只读挂载">
                <el-switch v-model="form.nfs.readOnly" />
              </el-form-item>
            </template>

            <template v-else-if="form.sourceType === 'hostPath'">
              <el-form-item label="节点目录" required>
                <el-input v-model="form.hostPath.path" placeholder="例如 /data/pv-01" />
                <div class="field-hint">数据保存在 Pod 所在节点上，Pod 换节点后数据不会跟随，只适合测试</div>
              </el-form-item>
              <el-form-item label="目录类型">
                <el-select v-model="form.hostPath.type" style="width: 100%">
                  <el-option v-for="opt in HOST_PATH_TYPES" :key="opt.value" :label="opt.label" :value="opt.value" />
                </el-select>
              </el-form-item>
            </template>

            <template v-else-if="form.sourceType === 'local'">
              <el-form-item label="磁盘路径" required>
                <el-input v-model="form.local.path" placeholder="例如 /mnt/disks/ssd1" />
              </el-form-item>
              <el-form-item label="所在节点" required>
                <el-select v-model="form.local.node" filterable placeholder="选择磁盘所在的节点" :loading="nodesLoading" style="width: 100%">
                  <el-option v-for="node in nodes" :key="node" :label="node" :value="node" />
                </el-select>
                <div class="field-hint">使用该持久卷的 Pod 会被调度到这个节点</div>
              </el-form-item>
              <el-form-item label="文件系统">
                <el-input v-model="form.local.fsType" placeholder="可选，例如 ext4 / xfs" />
              </el-form-item>
            </template>

            <template v-else>
              <el-form-item label="驱动名称" required>
                <el-input v-model="form.csi.driver" placeholder="例如 rbd.csi.ceph.com" />
              </el-form-item>
              <el-form-item label="卷 ID" required>
                <el-input v-model="form.csi.volumeHandle" placeholder="存储系统中已有卷的唯一 ID" />
              </el-form-item>
              <el-form-item label="文件系统">
                <el-input v-model="form.csi.fsType" placeholder="可选，例如 ext4" />
              </el-form-item>
              <el-form-item label="只读挂载">
                <el-switch v-model="form.csi.readOnly" />
              </el-form-item>
            </template>
          </el-form>
          <KVRowsTable v-if="form.sourceType === 'csi'" v-model="form.csi.attributes" title="卷属性" add-text="添加属性" empty-text="驱动需要的额外属性，没有可以留空" />
        </div>

        <!-- 第 3 步: 容量与访问 -->
        <div v-else class="basic-form">
          <el-form :model="form" label-width="100px" label-position="right">
            <el-form-item label="容量" required>
              <div class="fd-capacity">
                <el-input-number v-model="form.capacityValue" :min="1" :max="1048576" :step="1" controls-position="right" />
                <el-select v-model="form.capacityUnit">
                  <el-option v-for="unit in CAPACITY_UNITS" :key="unit" :label="unit" :value="unit" />
                </el-select>
              </div>
            </el-form-item>
            <el-form-item label="存储类">
              <el-select v-model="form.storageClassName" clearable placeholder="不指定" :loading="scLoading" style="width: 100%">
                <el-option v-for="sc in storageClasses" :key="sc.name" :label="sc.alias ? `${sc.name}（${sc.alias}）` : sc.name" :value="sc.name" />
              </el-select>
              <div class="field-hint">PVC 指定相同的存储类时才会绑定这个持久卷</div>
            </el-form-item>
            <el-form-item label="访问模式" required class="is-full">
              <el-checkbox-group v-model="form.accessModes" class="fd-access-modes">
                <el-checkbox v-for="mode in ACCESS_MODE_OPTIONS" :key="mode.value" :value="mode.value" border>
                  <span class="fd-access-mode">{{ mode.label }}</span>
                  <span class="fd-access-mode__desc">{{ mode.desc }}</span>
                </el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="回收策略">
              <el-radio-group v-model="form.reclaimPolicy">
                <el-radio-button value="Retain">保留</el-radio-button>
                <el-radio-button value="Delete">删除</el-radio-button>
                <el-radio-button value="Recycle" :disabled="!supportsRecycle">回收</el-radio-button>
              </el-radio-group>
              <div class="field-hint">{{ reclaimHint }}</div>
            </el-form-item>
            <el-form-item label="卷模式">
              <el-radio-group v-model="form.volumeMode">
                <el-radio-button value="Filesystem">文件系统</el-radio-button>
                <el-radio-button value="Block" :disabled="!supportsBlock">块设备</el-radio-button>
              </el-radio-group>
              <div class="field-hint">块设备模式只支持本地存储和 CSI</div>
            </el-form-item>
            <el-form-item label="挂载选项" class="is-full">
              <el-select v-model="form.mountOptions" multiple filterable allow-create default-first-option :reserve-keyword="false" placeholder="输入后回车添加，例如 nfsvers=4.1" style="width: 100%" />
            </el-form-item>
          </el-form>
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
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Check, CircleCheckFilled } from '@element-plus/icons-vue'
import YamlEditor from '@/components/YamlEditor.vue'
import KVRowsTable from '@/components/KVRowsTable.vue'
import { pvApi, type PVSourceType } from '@/api/pv'
import { storageClassApi, type StorageClassVO } from '@/api/storageclass'
import { clusterApi } from '@/api/cluster'
import { ACCESS_MODE_OPTIONS, CAPACITY_UNITS, rowsToMap, type CapacityUnit, type ParamRow } from '@/utils/storageclass'
import { jsonToYaml } from '@/utils/yaml'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    clusterId: string
    // 打开时预选的存储类(从存储类详情页创建时使用)
    defaultStorageClass?: string
  }>(),
  { defaultStorageClass: '' }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: []
}>()

const steps = [
  { title: '基本信息', desc: '名称、别名与标签' },
  { title: '存储来源', desc: 'NFS、本地盘或 CSI' },
  { title: '容量与访问', desc: '容量、访问模式与回收策略' }
]

const SOURCES: Array<{ value: PVSourceType; label: string; short: string; color: string; desc: string }> = [
  { value: 'nfs', label: 'NFS', short: 'NFS', color: 'linear-gradient(135deg, #34d399 0%, #059669 100%)', desc: '已有的 NFS 共享目录，支持多节点读写' },
  { value: 'local', label: '本地存储', short: 'LOC', color: 'linear-gradient(135deg, #94a3b8 0%, #475569 100%)', desc: '节点上的本地磁盘，性能最好，需指定节点' },
  { value: 'csi', label: 'CSI', short: 'CSI', color: 'linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%)', desc: '由 CSI 驱动管理的已有卷，如云硬盘' },
  { value: 'hostPath', label: 'HostPath', short: 'HP', color: 'linear-gradient(135deg, #fbbf24 0%, #d97706 100%)', desc: '节点上的目录，仅适合单节点测试' }
]

const HOST_PATH_TYPES = [
  { value: '', label: '不检查' },
  { value: 'DirectoryOrCreate', label: '目录，不存在时自动创建' },
  { value: 'Directory', label: '目录，必须已存在' },
  { value: 'FileOrCreate', label: '文件，不存在时自动创建' },
  { value: 'File', label: '文件，必须已存在' }
]

const current = ref(0)
const yamlMode = ref(false)
const submitting = ref(false)
const basicRef = ref<FormInstance>()

const storageClasses = ref<StorageClassVO[]>([])
const scLoading = ref(false)
const nodes = ref<string[]>([])
const nodesLoading = ref(false)

const createForm = () => ({
  name: '',
  alias: '',
  describe: '',
  labels: [] as ParamRow[],
  sourceType: 'nfs' as PVSourceType,
  nfs: { server: '', path: '', readOnly: false },
  hostPath: { path: '', type: 'DirectoryOrCreate' },
  local: { path: '', node: '', fsType: '' },
  csi: { driver: '', volumeHandle: '', fsType: '', readOnly: false, attributes: [] as ParamRow[] },
  capacityValue: 10,
  capacityUnit: 'Gi' as CapacityUnit,
  storageClassName: '',
  accessModes: ['ReadWriteOnce'] as string[],
  reclaimPolicy: 'Retain' as 'Retain' | 'Delete' | 'Recycle',
  volumeMode: 'Filesystem' as 'Filesystem' | 'Block',
  mountOptions: [] as string[]
})

const form = reactive(createForm())

const rules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    {
      pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/,
      message: '只能包含小写字母、数字、"-" 和 "."，且以字母或数字开头和结尾',
      trigger: 'blur'
    }
  ]
}

const supportsRecycle = computed(() => form.sourceType === 'nfs' || form.sourceType === 'hostPath')
const supportsBlock = computed(() => form.sourceType === 'local' || form.sourceType === 'csi')

const reclaimHint = computed(() => {
  switch (form.reclaimPolicy) {
    case 'Delete':
      return 'PVC 删除后同时删除持久卷和底层存储（需要存储插件支持）'
    case 'Recycle':
      return 'PVC 删除后清空数据，持久卷可以再次被绑定（已废弃，不建议使用）'
    default:
      return 'PVC 删除后保留持久卷和数据，需要手动清理后才能再次使用'
  }
})

// 切换存储来源后, 去掉不支持的选项
watch(
  () => form.sourceType,
  () => {
    if (!supportsRecycle.value && form.reclaimPolicy === 'Recycle') form.reclaimPolicy = 'Retain'
    if (!supportsBlock.value && form.volumeMode === 'Block') form.volumeMode = 'Filesystem'
  }
)

const selectSource = (value: PVSourceType) => {
  form.sourceType = value
  if (value === 'local' && !nodes.value.length) loadNodes()
}

const loadStorageClasses = async () => {
  scLoading.value = true
  try {
    const resp = await storageClassApi.getList({ clusterId: props.clusterId, page: 1, pageSize: 1000 })
    storageClasses.value = resp.data?.items || []
  } catch {
    storageClasses.value = []
  } finally {
    scLoading.value = false
  }
}

const loadNodes = async () => {
  nodesLoading.value = true
  try {
    const resp = await clusterApi.getNodes(props.clusterId, { page: 1, pageSize: 1000 })
    nodes.value = (resp.data?.items || []).map((n: any) => n.hostname || n.name).filter(Boolean)
  } catch {
    nodes.value = []
  } finally {
    nodesLoading.value = false
  }
}

const reset = () => {
  Object.assign(form, createForm())
  form.storageClassName = props.defaultStorageClass || ''
  current.value = 0
  yamlMode.value = false
  loadStorageClasses()
}

// 校验当前步骤, 不通过时提示并返回 false
const validateStep = async (step: number) => {
  if (step === 0) {
    if (!basicRef.value) return !!form.name
    try {
      await basicRef.value.validate()
      return true
    } catch {
      return false
    }
  }
  if (step === 1) {
    const missing =
      (form.sourceType === 'nfs' && (!form.nfs.server.trim() || !form.nfs.path.trim()) && '请填写 NFS 服务器地址和共享路径') ||
      (form.sourceType === 'hostPath' && !form.hostPath.path.trim() && '请填写节点目录') ||
      (form.sourceType === 'local' && (!form.local.path.trim() || !form.local.node) && '请填写磁盘路径并选择节点') ||
      (form.sourceType === 'csi' && (!form.csi.driver.trim() || !form.csi.volumeHandle.trim()) && '请填写 CSI 驱动名称和卷 ID')
    if (missing) {
      ElMessage.warning(missing)
      return false
    }
    return true
  }
  if (!form.accessModes.length) {
    ElMessage.warning('请至少选择一种访问模式')
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

const capacity = computed(() => `${form.capacityValue}${form.capacityUnit}`)

const buildAnnotations = () => {
  const ann: Record<string, string> = {}
  if (form.alias.trim()) ann.alias = form.alias.trim()
  if (form.describe.trim()) ann.describe = form.describe.trim()
  return ann
}

// 预览与后端 buildPersistentVolume 生成的对象一致
const previewYaml = computed(() => {
  const metadata: Record<string, unknown> = { name: form.name || '<名称>' }
  const labels = rowsToMap(form.labels)
  if (Object.keys(labels).length) metadata.labels = labels
  const ann = buildAnnotations()
  if (Object.keys(ann).length) metadata.annotations = ann

  const spec: Record<string, unknown> = {
    capacity: { storage: capacity.value },
    accessModes: form.accessModes,
    persistentVolumeReclaimPolicy: form.reclaimPolicy
  }
  if (form.storageClassName) spec.storageClassName = form.storageClassName
  spec.volumeMode = form.volumeMode
  if (form.mountOptions.length) spec.mountOptions = form.mountOptions
  switch (form.sourceType) {
    case 'nfs':
      spec.nfs = { server: form.nfs.server || '<服务器地址>', path: form.nfs.path || '<共享路径>', ...(form.nfs.readOnly ? { readOnly: true } : {}) }
      break
    case 'hostPath':
      spec.hostPath = { path: form.hostPath.path || '<节点目录>', ...(form.hostPath.type ? { type: form.hostPath.type } : {}) }
      break
    case 'local':
      spec.local = { path: form.local.path || '<磁盘路径>', ...(form.local.fsType ? { fsType: form.local.fsType } : {}) }
      spec.nodeAffinity = {
        required: {
          nodeSelectorTerms: [{ matchExpressions: [{ key: 'kubernetes.io/hostname', operator: 'In', values: [form.local.node || '<节点>'] }] }]
        }
      }
      break
    case 'csi': {
      const attrs = rowsToMap(form.csi.attributes)
      spec.csi = {
        driver: form.csi.driver || '<驱动名称>',
        volumeHandle: form.csi.volumeHandle || '<卷 ID>',
        ...(form.csi.fsType ? { fsType: form.csi.fsType } : {}),
        ...(form.csi.readOnly ? { readOnly: true } : {}),
        ...(Object.keys(attrs).length ? { volumeAttributes: attrs } : {})
      }
      break
    }
  }
  return jsonToYaml({ apiVersion: 'v1', kind: 'PersistentVolume', metadata, spec }) + '\n'
})

const submit = async () => {
  for (let step = 0; step < steps.length; step++) {
    // YAML 预览模式下表单没有渲染, 第 1 步只能检查名称是否填写
    const ok = step === 0 && (yamlMode.value || current.value !== 0) ? !!form.name.trim() : await validateStep(step)
    if (!ok) {
      if (step === 0) ElMessage.warning('请填写名称')
      yamlMode.value = false
      current.value = step
      return
    }
  }

  submitting.value = true
  try {
    await pvApi.create({
      clusterId: props.clusterId,
      name: form.name.trim(),
      alias: form.alias.trim(),
      describe: form.describe.trim(),
      capacity: capacity.value,
      accessModes: form.accessModes,
      reclaimPolicy: form.reclaimPolicy,
      storageClassName: form.storageClassName,
      volumeMode: form.volumeMode,
      mountOptions: form.mountOptions,
      labels: rowsToMap(form.labels),
      sourceType: form.sourceType,
      nfs: form.sourceType === 'nfs' ? { ...form.nfs } : undefined,
      hostPath: form.sourceType === 'hostPath' ? { ...form.hostPath } : undefined,
      local: form.sourceType === 'local' ? { ...form.local } : undefined,
      csi:
        form.sourceType === 'csi'
          ? {
              driver: form.csi.driver,
              volumeHandle: form.csi.volumeHandle,
              fsType: form.csi.fsType,
              readOnly: form.csi.readOnly,
              volumeAttributes: rowsToMap(form.csi.attributes)
            }
          : undefined
    })
    ElMessage.success('创建成功')
    emit('update:modelValue', false)
    emit('success')
  } catch {
    // 请求拦截器已提示错误
  } finally {
    submitting.value = false
  }
}
</script>
