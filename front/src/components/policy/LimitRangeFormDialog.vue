<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="820px"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--compact"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="reset"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">{{ title }}</h3>
      </div>
    </template>
    <div v-loading="loading" class="config-editor">
      <div class="config-content">
        <p class="lf-tip">
          容器没有声明资源预留 / 上限时，自动使用这里的默认值；最小值、最大值限制容器可以声明的范围。留空表示不设置。
        </p>
        <el-form label-width="120px">
          <el-form-item v-if="!namespaceFixed" label="命名空间" required>
            <el-select v-model="namespace" filterable :disabled="isEdit" placeholder="请选择命名空间" style="width: 100%">
              <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
            </el-select>
          </el-form-item>
          <el-form-item label="名称" required>
            <el-input v-model="name" :disabled="isEdit" maxlength="253" />
          </el-form-item>
        </el-form>
        <div class="lf-table">
          <div class="lf-row lf-row--head">
            <span />
            <span v-for="f in FIELDS" :key="f.key">{{ f.label }}</span>
          </div>
          <div v-for="r in RESOURCES" :key="r.key" class="lf-row">
            <span class="lf-res">{{ r.label }}</span>
            <el-input v-for="f in FIELDS" :key="f.key" v-model="values[f.key][r.key]" :placeholder="r.placeholder[f.key]" clearable class="mono-input" />
          </div>
        </div>
        <el-alert v-if="otherTypes" type="info" :closable="false" show-icon class="lf-alert">
          该对象还包含 {{ otherTypes }} 项容器组 / 持久卷声明的限制，保存时会保留，如需修改请使用「编辑YAML」。
        </el-alert>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="saving" :disabled="loading" @click="save">保存</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { editableCopy, resourceApi, saveResourceObject, type KubeObject } from '@/api/resource'
import { namespaceApi } from '@/api/namespace'
import { DEFAULT_LIMIT_RANGE_NAME } from '@/utils/policy'
import { isQuantity, parseQuantity } from '@/utils/resource'

type FieldKey = 'defaultRequest' | 'default' | 'min' | 'max'

const props = withDefaults(
  defineProps<{ modelValue: boolean; clusterId: string; namespace?: string; editName?: string }>(),
  { namespace: '', editName: '' }
)
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; success: [] }>()

const LR_REF = { kind: 'limitrange' as const }
const FIELDS: { key: FieldKey; label: string }[] = [
  { key: 'defaultRequest', label: '默认预留' },
  { key: 'default', label: '默认上限' },
  { key: 'min', label: '最小值' },
  { key: 'max', label: '最大值' }
]
const RESOURCES = [
  { key: 'cpu', label: 'CPU', placeholder: { defaultRequest: '如 100m', default: '如 500m', min: '如 10m', max: '如 4' } },
  { key: 'memory', label: '内存', placeholder: { defaultRequest: '如 128Mi', default: '如 512Mi', min: '如 16Mi', max: '如 8Gi' } }
]

const isEdit = computed(() => !!props.editName)
const namespaceFixed = computed(() => !!props.namespace)
const title = computed(() => (isEdit.value ? '编辑默认资源限制' : '设置默认资源限制'))

const loading = ref(false)
const saving = ref(false)
const namespace = ref('')
const name = ref(DEFAULT_LIMIT_RANGE_NAME)
const namespaces = ref<string[]>([])
const original = ref<KubeObject | null>(null)
const otherTypes = ref(0)
const createValues = () => ({ defaultRequest: {}, default: {}, min: {}, max: {} }) as Record<FieldKey, Record<string, string>>
const values = reactive(createValues())

const reset = async () => {
  Object.assign(values, createValues())
  original.value = null
  otherTypes.value = 0
  namespace.value = props.namespace
  name.value = props.editName || DEFAULT_LIMIT_RANGE_NAME
  if (!namespaceFixed.value) {
    namespaceApi
      .getList({ clusterId: props.clusterId, page: 1, pageSize: 10000 })
      .then(resp => (namespaces.value = (resp.data?.items || []).map(ns => ns.name)))
      .catch(() => (namespaces.value = []))
  }
  if (!isEdit.value) return
  loading.value = true
  try {
    const { data } = await resourceApi.object({ ...LR_REF, clusterId: props.clusterId, namespace: namespace.value, name: props.editName })
    original.value = data
    const limits: KubeObject[] = data.spec?.limits || []
    otherTypes.value = limits.filter(l => l.type !== 'Container').length
    const container = limits.find(l => l.type === 'Container') || {}
    FIELDS.forEach(f => RESOURCES.forEach(r => (values[f.key][r.key] = container[f.key]?.[r.key] ? String(container[f.key][r.key]) : '')))
  } catch {
    emit('update:modelValue', false)
  } finally {
    loading.value = false
  }
}

const validate = (): string => {
  for (const r of RESOURCES) {
    const v: Record<string, number | null> = {}
    for (const f of FIELDS) {
      const text = (values[f.key][r.key] || '').trim()
      if (text && !isQuantity(text)) return `${r.label}的${f.label} ${text} 格式不正确`
      v[f.key] = text ? parseQuantity(text) : null
    }
    const lte = (a: FieldKey, b: FieldKey) => v[a] === null || v[b] === null || (v[a] as number) <= (v[b] as number)
    if (!lte('defaultRequest', 'default')) return `${r.label}的默认预留不能大于默认上限`
    if (!lte('min', 'defaultRequest') || !lte('min', 'default')) return `${r.label}的默认值不能小于最小值`
    if (!lte('defaultRequest', 'max') || !lte('default', 'max')) return `${r.label}的默认值不能大于最大值`
  }
  return ''
}

const save = async () => {
  if (!namespace.value) {
    ElMessage.warning('请选择命名空间')
    return
  }
  const message = validate()
  if (message) {
    ElMessage.warning(message)
    return
  }
  const container: KubeObject = { type: 'Container' }
  FIELDS.forEach(f => {
    const entries = RESOURCES.map(r => [r.key, (values[f.key][r.key] || '').trim()]).filter(([, v]) => v)
    if (entries.length) container[f.key] = Object.fromEntries(entries)
  })
  if (Object.keys(container).length === 1) {
    ElMessage.warning('请至少设置一项')
    return
  }
  const base: KubeObject = original.value
    ? editableCopy(original.value)
    : { apiVersion: 'v1', kind: 'LimitRange', metadata: { name: name.value.trim(), namespace: namespace.value }, spec: {} }
  const kept = ((base.spec?.limits || []) as KubeObject[]).filter(l => l.type !== 'Container')
  // 保留容器类型中表单不支持的字段(例如 maxLimitRequestRatio)
  const oldContainer = ((base.spec?.limits || []) as KubeObject[]).find(l => l.type === 'Container')
  if (oldContainer?.maxLimitRequestRatio) container.maxLimitRequestRatio = oldContainer.maxLimitRequestRatio
  base.spec = { ...(base.spec || {}), limits: [container, ...kept] }
  saving.value = true
  try {
    await saveResourceObject(props.clusterId, LR_REF, base, !isEdit.value)
    ElMessage.success('默认资源限制已保存')
    emit('update:modelValue', false)
    emit('success')
  } catch {
    // 请求拦截器已提示错误
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.lf-tip {
  margin: 0 0 14px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.6;
}

.lf-table {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 4px 0 14px;
}

.lf-row {
  display: grid;
  grid-template-columns: 80px repeat(4, minmax(0, 1fr));
  align-items: center;
  gap: 10px;
}

.lf-row--head span {
  color: #64748b;
  font-size: 12px;
}

.lf-res {
  color: #1e293b;
  font-weight: 600;
}

.lf-alert {
  margin-top: 6px;
}

.mono-input :deep(input) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
