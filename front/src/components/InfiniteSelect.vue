<template>
  <el-select
    :model-value="modelValue"
    :placeholder="placeholder"
    :style="style"
    :clearable="clearable"
    :filterable="filterable"
    :remote="remote"
    :remote-method="handleRemoteMethod"
    :multiple="multiple"
    :loading="loading"
    @update:model-value="handleInput"
    @change="handleChange"
    @visible-change="handleVisibleChange"
    @clear="emit('clear')"
  >
    <el-option v-for="item in options" :key="read(item, optionKey)" :label="read(item, optionLabel)" :value="read(item, optionValue)" />
    <div v-if="hasMore && !loading" class="load-more-trigger" @click="loadMore">
      <span class="load-more-text">点击加载更多...</span>
    </div>
    <div v-if="loading" class="loading-indicator">
      <span class="loading-text">
        <el-icon class="is-loading"><Loading /></el-icon>
        加载中...
      </span>
    </div>
  </el-select>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { Loading } from '@element-plus/icons-vue'

// 分页加载选项的下拉框: 打开时加载第一页, 滚动到底部加载更多, 输入时远程搜索
export type OptionItem = Record<string, unknown>
export type OptionValue = string | number
type Field = string | ((item: OptionItem) => OptionValue)

interface Props {
  modelValue?: OptionValue | OptionValue[] | null
  placeholder?: string
  style?: string | Record<string, string>
  clearable?: boolean
  filterable?: boolean
  remote?: boolean
  multiple?: boolean
  // 获取一页数据
  fetchData: (params: { page: number; pageSize: number; keyword?: string }) => Promise<{ data: OptionItem[]; total: number }>
  optionKey?: Field
  optionLabel?: Field
  optionValue?: Field
  pageSize?: number
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: undefined,
  placeholder: '请选择',
  style: undefined,
  clearable: true,
  filterable: false,
  remote: false,
  multiple: false,
  optionKey: 'id',
  optionLabel: 'name',
  optionValue: 'id',
  pageSize: 20
})

const emit = defineEmits<{
  'update:modelValue': [value: OptionValue | OptionValue[]]
  change: [value: OptionValue | OptionValue[]]
  clear: []
}>()

const loading = ref(false)
const options = ref<OptionItem[]>([])
const currentPage = ref(1)
const total = ref(0)
const keyword = ref('')
const hasMore = computed(() => options.value.length < total.value)

const read = (item: OptionItem, field: Field): OptionValue => {
  if (typeof field === 'function') return field(item)
  const v = item[field]
  return typeof v === 'number' ? v : String(v ?? '')
}

// 清空时 el-select 给出 undefined, 单选统一为空字符串(表示全部), 多选为空数组
const normalize = (value: unknown): OptionValue | OptionValue[] => {
  if (props.multiple) return Array.isArray(value) ? (value as OptionValue[]) : []
  return value === undefined || value === null ? '' : (value as OptionValue)
}

// 必须同步回写 v-model: el-select 先触发 update:modelValue 再触发 change, 页面在 change 里直接读取绑定的值查询
const handleInput = (value: unknown) => emit('update:modelValue', normalize(value))
const handleChange = (value: unknown) => emit('change', normalize(value))

// 每次加载递增, 旧请求(例如快速输入时之前的搜索)的结果直接丢弃
let loadSeq = 0
const loadData = async (reset = false) => {
  if (loading.value && !reset) return
  const seq = ++loadSeq
  loading.value = true
  try {
    const page = reset ? 1 : currentPage.value
    const result = await props.fetchData({ page, pageSize: props.pageSize, keyword: keyword.value })
    if (seq !== loadSeq) return
    options.value = reset ? result.data : [...options.value, ...result.data]
    total.value = result.total
    currentPage.value = page + 1
  } catch {
    // 请求拦截器已提示错误
  } finally {
    if (seq === loadSeq) loading.value = false
  }
}

const loadMore = () => {
  if (hasMore.value && !loading.value) loadData(false)
}

const handleRemoteMethod = (query: string) => {
  keyword.value = query
  loadData(true)
}

const onScroll = (event: Event) => {
  const target = event.target as HTMLElement
  if (target.scrollTop + target.clientHeight >= target.scrollHeight - 10) loadMore()
}

let scrollEl: Element | null = null
const handleVisibleChange = (visible: boolean) => {
  if (visible && options.value.length === 0) loadData(true)
  scrollEl?.removeEventListener('scroll', onScroll)
  scrollEl = null
  if (visible) {
    nextTick(() => {
      // 下拉面板挂载在 body 下, 取当前可见的列表
      const lists = Array.from(document.querySelectorAll('.el-select-dropdown__wrap'))
      scrollEl = lists.find(el => (el as HTMLElement).offsetParent !== null) || null
      scrollEl?.addEventListener('scroll', onScroll)
    })
  }
}

const reset = () => {
  options.value = []
  currentPage.value = 1
  total.value = 0
  keyword.value = ''
}

const refresh = () => {
  reset()
  loadData(true)
}

defineExpose({ reset, refresh, loadMore })

// 数据来源变化(例如切换集群)时重新加载
watch(() => props.fetchData, refresh)

onMounted(() => loadData(true))
</script>

<style scoped>
.load-more-trigger {
  padding: 8px 12px;
  text-align: center;
  cursor: pointer;
  border-top: 1px solid var(--el-border-color-light);
}

.load-more-trigger:hover {
  background-color: var(--el-fill-color-light);
}

.loading-indicator {
  padding: 8px 12px;
  text-align: center;
  border-top: 1px solid var(--el-border-color-light);
}

.loading-indicator .el-icon {
  margin-right: 4px;
}

.load-more-text {
  font-size: 12px;
  color: var(--el-color-info);
}

.loading-text {
  font-size: 12px;
  color: var(--el-color-info);
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
