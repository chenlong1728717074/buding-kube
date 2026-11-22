<template>
  <el-dialog
    :model-value="modelValue"
    title="创建持久卷声明"
    width="1080px"
    top="6vh"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--wide"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="reset"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">创建持久卷声明</h3>
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
          <YamlEditor :model-value="previewYaml" :readonly="true" height="100%" title="持久卷声明 YAML 预览" :filename="`${form.name || 'pvc'}.yaml`" />
        </div>

        <!-- 第 1 步: 基本信息 -->
        <div v-else-if="current === 0" class="basic-form">
          <el-form ref="basicRef" :model="form" :rules="rules" label-width="100px" label-position="right">
            <el-form-item label="命名空间" prop="namespace">
              <el-select v-model="form.namespace" filterable placeholder="请选择命名空间" :loading="nsLoading" style="width: 100%">
                <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="例如 mysql-data" maxlength="253" show-word-limit />
              <div class="field-hint">小写字母、数字、"-" 和 "."，创建后不可修改</div>
            </el-form-item>
            <el-form-item label="别名">
              <el-input v-model="form.alias" placeholder="可选，用于列表展示" maxlength="50" show-word-limit />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="form.describe" type="textarea" :rows="3" placeholder="可选，说明用途" maxlength="200" show-word-limit />
            </el-form-item>
          </el-form>
        </div>

        <!-- 第 2 步: 存储设置 -->
        <div v-else-if="current === 1" class="basic-form">
          <p class="fd-subtitle">创建方式</p>
          <div class="fd-choice-grid is-two">
            <button type="button" :class="['fd-choice', { 'is-active': form.mode === 'class' }]" @click="form.mode = 'class'">
              <span class="fd-choice__icon" style="background: linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%)">SC</span>
              <span class="fd-choice__body">
                <span class="fd-choice__title"><span>通过存储类创建</span></span>
                <span class="fd-choice__desc">由存储类自动创建持久卷，最常用的方式</span>
              </span>
              <el-icon class="fd-choice__check"><CircleCheckFilled /></el-icon>
            </button>
            <button type="button" :class="['fd-choice', { 'is-active': form.mode === 'bind' }]" @click="selectBindMode">
              <span class="fd-choice__icon" style="background: linear-gradient(135deg, #34d399 0%, #059669 100%)">PV</span>
              <span class="fd-choice__body">
                <span class="fd-choice__title"><span>绑定已有持久卷</span></span>
                <span class="fd-choice__desc">选择一个状态为 Available 的持久卷直接使用</span>
              </span>
              <el-icon class="fd-choice__check"><CircleCheckFilled /></el-icon>
            </button>
          </div>

          <template v-if="form.mode === 'class'">
            <p class="fd-subtitle">存储类<small>{{ storageClasses.length ? '' : '集群中还没有存储类' }}</small></p>
            <div v-loading="scLoading" class="fd-choice-grid">
              <button
                v-for="sc in storageClasses"
                :key="sc.name"
                type="button"
                :class="['fd-choice', { 'is-active': form.storageClassName === sc.name }]"
                @click="selectStorageClass(sc)"
              >
                <span class="fd-choice__icon" :style="{ background: typeColor(sc.storageTypeKey) }">{{ (sc.storageType || '存储').slice(0, 1) }}</span>
                <span class="fd-choice__body">
                  <span class="fd-choice__title">
                    <span>{{ sc.name }}</span>
                    <span v-if="sc.isDefault" class="fd-choice__badge is-primary">默认</span>
                    <span class="fd-choice__badge">{{ sc.storageType || '未分类' }}</span>
                  </span>
                  <span v-if="sc.alias" class="fd-choice__desc">{{ sc.alias }}</span>
                  <span class="fd-choice__meta">{{ sc.provisioner }}</span>
                </span>
                <el-icon class="fd-choice__check"><CircleCheckFilled /></el-icon>
              </button>
            </div>
          </template>

          <template v-else>
            <p class="fd-subtitle">持久卷<small>只列出状态为 Available 的持久卷</small></p>
            <el-table
              v-loading="pvLoading"
              :data="availablePVs"
              highlight-current-row
              class="fd-pick-table"
              style="width: 100%; margin-bottom: 22px"
              @current-change="selectPV"
            >
              <el-table-column width="48" align="center">
                <template #default="{ row }">
                  <el-radio :model-value="form.volumeName" :value="row.name" @click.stop="selectPV(row)">&nbsp;</el-radio>
                </template>
              </el-table-column>
              <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
              <el-table-column label="容量" width="100">
                <template #default="{ row }">{{ row.capacity?.storage || '-' }}</template>
              </el-table-column>
              <el-table-column label="访问模式" width="150">
                <template #default="{ row }">{{ formatAccessModes(row.accessModes) }}</template>
              </el-table-column>
              <el-table-column label="存储类" min-width="140" show-overflow-tooltip>
                <template #default="{ row }">{{ row.storageClassName || '不指定' }}</template>
              </el-table-column>
              <el-table-column label="来源" width="110">
                <template #default="{ row }">{{ PV_SOURCE_LABELS[row.sourceType] || row.sourceType || '-' }}</template>
              </el-table-column>
              <template #empty>
                <span>没有可用的持久卷，可以先到「PersistentVolume管理」创建</span>
              </template>
            </el-table>
          </template>

          <el-form :model="form" label-width="100px" label-position="right">
            <el-form-item label="访问模式" required class="is-full">
              <el-radio-group v-model="form.accessMode" class="fd-access-modes">
                <el-radio v-for="mode in allowedAccessModes" :key="mode.value" :value="mode.value" border>
                  <span class="fd-access-mode">{{ mode.label }}</span>
                  <span class="fd-access-mode__desc">{{ mode.desc }}</span>
                </el-radio>
              </el-radio-group>
              <div v-if="accessModeHint" class="field-hint">{{ accessModeHint }}</div>
            </el-form-item>
            <el-form-item label="容量" required>
              <div class="fd-capacity">
                <el-input-number v-model="form.capacityValue" :min="1" :max="1048576" controls-position="right" />
                <el-select v-model="form.capacityUnit">
                  <el-option v-for="unit in CAPACITY_UNITS" :key="unit" :label="unit" :value="unit" />
                </el-select>
              </div>
              <div v-if="selectedPV" class="field-hint">不能超过持久卷的容量 {{ selectedPV.capacity?.storage }}</div>
            </el-form-item>
            <el-form-item label="卷模式">
              <el-radio-group v-model="form.volumeMode">
                <el-radio-button value="Filesystem">文件系统</el-radio-button>
                <el-radio-button value="Block">块设备</el-radio-button>
              </el-radio-group>
            </el-form-item>
          </el-form>
        </div>

        <!-- 第 3 步: 高级设置 -->
        <div v-else class="basic-form">
          <KVRowsTable v-model="form.labels" title="标签" add-text="添加标签" empty-text="暂无标签" />
          <KVRowsTable v-model="form.annotations" title="注解" add-text="添加注解" empty-text="暂无注解" />
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
import { pvcApi } from '@/api/pvc'
import { pvApi, type PVVO } from '@/api/pv'
import { namespaceApi } from '@/api/namespace'
import { storageClassApi, type StorageClassVO } from '@/api/storageclass'
import {
  ACCESS_MODE_OPTIONS,
  CAPACITY_UNITS,
  PV_SOURCE_LABELS,
  capacityToMi,
  formatAccessModes,
  parseCapacity,
  rowsToMap,
  type CapacityUnit,
  type ParamRow
} from '@/utils/storageclass'
import { jsonToYaml } from '@/utils/yaml'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    clusterId: string
    defaultNamespace?: string
    // 默认选中的存储类(从存储类详情页创建时传入)
    defaultStorageClass?: string
  }>(),
  { defaultNamespace: '', defaultStorageClass: '' }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: []
}>()

const steps = [
  { title: '基本信息', desc: '命名空间、名称与描述' },
  { title: '存储设置', desc: '存储类、访问模式与容量' },
  { title: '高级设置', desc: '标签与注解' }
]

const current = ref(0)
const yamlMode = ref(false)
const submitting = ref(false)
const basicRef = ref<FormInstance>()

const namespaces = ref<string[]>([])
const nsLoading = ref(false)
const storageClasses = ref<StorageClassVO[]>([])
const scLoading = ref(false)
const availablePVs = ref<PVVO[]>([])
const pvLoading = ref(false)

const createForm = () => ({
  namespace: '',
  name: '',
  alias: '',
  describe: '',
  mode: 'class' as 'class' | 'bind',
  storageClassName: '',
  volumeName: '',
  accessMode: 'ReadWriteOnce',
  capacityValue: 10,
  capacityUnit: 'Gi' as CapacityUnit,
  volumeMode: 'Filesystem' as 'Filesystem' | 'Block',
  labels: [] as ParamRow[],
  annotations: [] as ParamRow[]
})

const form = reactive(createForm())

const rules: FormRules = {
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    {
      pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/,
      message: '只能包含小写字母、数字、"-" 和 "."，且以字母或数字开头和结尾',
      trigger: 'blur'
    }
  ]
}

const selectedSC = computed(() => storageClasses.value.find(sc => sc.name === form.storageClassName) || null)
const selectedPV = computed(() => availablePVs.value.find(pv => pv.name === form.volumeName) || null)

// 访问模式受存储类(注解中声明的支持范围)或所选持久卷限制
const allowedAccessModes = computed(() => {
  const limit = form.mode === 'bind' ? selectedPV.value?.accessModes : selectedSC.value?.accessModes
  if (!limit || !limit.length) return ACCESS_MODE_OPTIONS
  const filtered = ACCESS_MODE_OPTIONS.filter(o => limit.includes(o.value))
  return filtered.length ? filtered : ACCESS_MODE_OPTIONS
})

const accessModeHint = computed(() => {
  if (form.mode === 'bind') return selectedPV.value ? '只能选择持久卷支持的访问模式' : ''
  return selectedSC.value?.accessModes?.length ? '只能选择存储类支持的访问模式' : ''
})

watch(allowedAccessModes, modes => {
  if (!modes.some(m => m.value === form.accessMode)) form.accessMode = modes[0]?.value || 'ReadWriteOnce'
})

const typeColor = (key?: string) =>
  ({
    block: 'linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%)',
    file: 'linear-gradient(135deg, #34d399 0%, #059669 100%)',
    object: 'linear-gradient(135deg, #fbbf24 0%, #d97706 100%)'
  })[key || ''] || 'linear-gradient(135deg, #94a3b8 0%, #475569 100%)'

const selectStorageClass = (sc: StorageClassVO) => {
  form.storageClassName = sc.name
}

const selectPV = (pv?: PVVO | null) => {
  if (!pv) return
  form.volumeName = pv.name
  const parsed = parseCapacity(pv.capacity?.storage)
  if (parsed) {
    form.capacityValue = parsed.value
    form.capacityUnit = parsed.unit
  }
  if (pv.volumeMode === 'Block' || pv.volumeMode === 'Filesystem') form.volumeMode = pv.volumeMode
}

const selectBindMode = () => {
  form.mode = 'bind'
  if (!availablePVs.value.length) loadPVs()
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

const loadStorageClasses = async () => {
  scLoading.value = true
  try {
    const resp = await storageClassApi.getList({ clusterId: props.clusterId, page: 1, pageSize: 1000 })
    storageClasses.value = resp.data?.items || []
    // 默认选中集群的默认存储类
    if (!form.storageClassName) {
      form.storageClassName =
        (storageClasses.value.find(sc => sc.name === props.defaultStorageClass) ||
          storageClasses.value.find(sc => sc.isDefault) ||
          storageClasses.value[0])?.name || ''
    }
  } catch {
    storageClasses.value = []
  } finally {
    scLoading.value = false
  }
}

const loadPVs = async () => {
  pvLoading.value = true
  try {
    const resp = await pvApi.getPVList({ clusterId: props.clusterId, page: 1, pageSize: 10000 })
    availablePVs.value = (resp.data?.items || []).filter(pv => pv.phase === 'Available')
  } catch {
    availablePVs.value = []
  } finally {
    pvLoading.value = false
  }
}

const reset = () => {
  Object.assign(form, createForm())
  form.namespace = props.defaultNamespace || ''
  availablePVs.value = []
  current.value = 0
  yamlMode.value = false
  loadNamespaces()
  loadStorageClasses()
}

const capacity = computed(() => `${form.capacityValue}${form.capacityUnit}`)

const validateStep = async (step: number) => {
  if (step === 0) {
    if (!basicRef.value) return !!(form.namespace && form.name)
    try {
      await basicRef.value.validate()
      return true
    } catch {
      return false
    }
  }
  if (step === 1) {
    if (form.mode === 'class' && !form.storageClassName) {
      ElMessage.warning('请选择存储类')
      return false
    }
    if (form.mode === 'bind') {
      if (!selectedPV.value) {
        ElMessage.warning('请选择要绑定的持久卷')
        return false
      }
      if (capacityToMi(capacity.value) > capacityToMi(selectedPV.value.capacity?.storage)) {
        ElMessage.warning(`容量不能超过持久卷的容量 ${selectedPV.value.capacity?.storage}`)
        return false
      }
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

// 绑定持久卷时存储类必须与持久卷一致; 持久卷没有存储类时显式设为空, 防止被默认存储类接管
const effectiveStorageClass = computed(() =>
  form.mode === 'bind' ? selectedPV.value?.storageClassName || '' : form.storageClassName
)

const buildAnnotations = () => {
  const ann = rowsToMap(form.annotations)
  if (form.alias.trim()) ann.alias = form.alias.trim()
  if (form.describe.trim()) ann.describe = form.describe.trim()
  return ann
}

const previewYaml = computed(() => {
  const metadata: Record<string, unknown> = { name: form.name || '<名称>', namespace: form.namespace || '<命名空间>' }
  const labels = rowsToMap(form.labels)
  if (Object.keys(labels).length) metadata.labels = labels
  const ann = buildAnnotations()
  if (Object.keys(ann).length) metadata.annotations = ann
  const spec: Record<string, unknown> = {
    accessModes: [form.accessMode],
    resources: { requests: { storage: capacity.value } },
    storageClassName: effectiveStorageClass.value,
    volumeMode: form.volumeMode
  }
  if (form.mode === 'bind') spec.volumeName = form.volumeName || '<持久卷>'
  return jsonToYaml({ apiVersion: 'v1', kind: 'PersistentVolumeClaim', metadata, spec }) + '\n'
})

const submit = async () => {
  for (let step = 0; step < steps.length; step++) {
    const ok = step === 0 && (yamlMode.value || current.value !== 0) ? !!(form.namespace && form.name.trim()) : await validateStep(step)
    if (!ok) {
      if (step === 0) ElMessage.warning('请选择命名空间并填写名称')
      yamlMode.value = false
      current.value = step
      return
    }
  }

  submitting.value = true
  try {
    await pvcApi.create({
      clusterId: props.clusterId,
      namespace: form.namespace,
      name: form.name.trim(),
      alias: form.alias.trim(),
      describe: form.describe.trim(),
      capacity: capacity.value,
      accessModes: [form.accessMode],
      storageClassName: effectiveStorageClass.value,
      volumeName: form.mode === 'bind' ? form.volumeName : '',
      volumeMode: form.volumeMode,
      labels: rowsToMap(form.labels),
      annotations: rowsToMap(form.annotations)
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
