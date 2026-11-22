<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="860px"
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
        <p class="qf-tip">
          限制命名空间中资源的总量，留空表示不限制。设置了 CPU / 内存配额后，没有声明对应资源的容器组会被拒绝创建，建议同时设置「默认资源限制」。
        </p>
        <el-form label-width="150px">
          <el-form-item v-if="!namespaceFixed" label="命名空间" required>
            <InfiniteSelect v-model="namespace" :fetch-data="namespaceFetcher" v-bind="namespaceSelectConfig" :clearable="false" :disabled="isEdit" placeholder="请选择命名空间" style="width: 100%" />
          </el-form-item>
          <el-form-item label="名称" required>
            <el-input v-model="name" :disabled="isEdit" maxlength="253" />
          </el-form-item>
          <template v-for="g in QUOTA_GROUPS" :key="g.title">
            <p class="fd-subtitle">{{ g.title }}</p>
            <div class="qf-grid">
              <el-form-item v-for="item in g.items" :key="item.key" :label="item.label">
                <el-input v-model="values[item.key]" :placeholder="item.placeholder" clearable class="mono-input" />
              </el-form-item>
            </div>
          </template>
        </el-form>
        <KVRowsTable v-model="extra" title="其他配额" add-text="添加配额项" empty-text="没有其他配额项" key-placeholder="资源，如 count/ingresses.networking.k8s.io" value-placeholder="数量" />
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
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import KVRowsTable from '@/components/KVRowsTable.vue'
import { editableCopy, resourceApi, saveResourceObject, type KubeObject } from '@/api/resource'
import { DEFAULT_QUOTA_NAME, QUOTA_GROUPS } from '@/utils/policy'
import { isQuantity } from '@/utils/resource'
import type { KVRow } from '@/utils/job'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    clusterId: string
    // 命名空间详情中使用时固定
    namespace?: string
    editName?: string
  }>(),
  { namespace: '', editName: '' }
)
// 命名空间下拉框分页加载
const namespaceFetcher = computed(() => useNamespaceFetcher(props.clusterId))

const emit = defineEmits<{ 'update:modelValue': [value: boolean]; success: [] }>()

const QUOTA_REF = { kind: 'resourcequota' as const }
const FORM_KEYS = new Set(QUOTA_GROUPS.flatMap(g => g.items.map(i => i.key)))

const isEdit = computed(() => !!props.editName)
const namespaceFixed = computed(() => !!props.namespace)
const title = computed(() => (isEdit.value ? '编辑资源配额' : '设置资源配额'))

const loading = ref(false)
const saving = ref(false)
const namespace = ref('')
const name = ref(DEFAULT_QUOTA_NAME)
const values = reactive<Record<string, string>>({})
const extra = ref<KVRow[]>([])
const original = ref<KubeObject | null>(null)

const reset = async () => {
  Object.keys(values).forEach(k => delete values[k])
  extra.value = []
  original.value = null
  namespace.value = props.namespace
  name.value = props.editName || DEFAULT_QUOTA_NAME
  if (!isEdit.value) return
  loading.value = true
  try {
    const { data } = await resourceApi.object({ ...QUOTA_REF, clusterId: props.clusterId, namespace: namespace.value, name: props.editName })
    original.value = data
    Object.entries((data.spec?.hard || {}) as Record<string, string>).forEach(([k, v]) => {
      if (FORM_KEYS.has(k)) values[k] = String(v)
      else extra.value.push({ key: k, value: String(v) })
    })
  } catch {
    emit('update:modelValue', false)
  } finally {
    loading.value = false
  }
}

const save = async () => {
  if (!namespace.value) {
    ElMessage.warning('请选择命名空间')
    return
  }
  if (!/^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$/.test(name.value.trim())) {
    ElMessage.warning('名称只能包含小写字母、数字、"-" 和 "."')
    return
  }
  const hard: Record<string, string> = {}
  for (const [k, v] of [...Object.entries(values), ...extra.value.map(r => [r.key.trim(), r.value] as [string, string])]) {
    const value = (v || '').trim()
    if (!k || !value) continue
    if (!isQuantity(value)) {
      ElMessage.warning(`${k} 的值 ${value} 格式不正确，例如 4、500m、8Gi`)
      return
    }
    hard[k] = value
  }
  if (!Object.keys(hard).length) {
    ElMessage.warning('请至少设置一项配额')
    return
  }
  const base: KubeObject = original.value
    ? editableCopy(original.value)
    : { apiVersion: 'v1', kind: 'ResourceQuota', metadata: { name: name.value.trim(), namespace: namespace.value }, spec: {} }
  base.spec = { ...(base.spec || {}), hard }
  saving.value = true
  try {
    await saveResourceObject(props.clusterId, QUOTA_REF, base, !isEdit.value)
    ElMessage.success('资源配额已保存')
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
.qf-tip {
  margin: 0 0 14px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.6;
}

.qf-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.mono-input :deep(input) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
