<template>
  <el-select
    v-model="internalValue"
    :placeholder="placeholder"
    :style="style"
    :clearable="clearable"
    :filterable="filterable"
    :remote="remote"
    :remote-method="handleRemoteMethod"
    :loading="loading"
    @change="handleChange"
    @visible-change="handleVisibleChange"
    @clear="handleClear"
  >
    <el-option
      v-for="item in displayOptions"
      :key="getOptionKey(item)"
      :label="getOptionLabel(item)"
      :value="getOptionValue(item)"
    />
    <div
      v-if="hasMore && !loading"
      ref="loadMoreTrigger"
      class="load-more-trigger"
      @click="loadMore"
    >
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
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { Loading } from '@element-plus/icons-vue'

interface Props {
  modelValue?: any
  placeholder?: string
  style?: string | object
  clearable?: boolean
  filterable?: boolean
  remote?: boolean
  // 获取数据的函数，返回 Promise<{ data: any[], total: number }>
  fetchData: (params: { page: number; pageSize: number; keyword?: string }) => Promise<{ data: any[]; total: number }>
  // 选项的键值映射
  optionKey?: string | ((item: any) => string | number)
  optionLabel?: string | ((item: any) => string)
  optionValue?: string | ((item: any) => any)
  // 每页大小
  pageSize?: number
}

interface Emits {
  (e: 'update:modelValue', value: any): void
  (e: 'change', value: any): void
  (e: 'clear'): void
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '请选择',
  clearable: true,
  filterable: false,
  remote: false,
  optionKey: 'id',
  optionLabel: 'name',
  optionValue: 'id',
  pageSize: 20
})

const emit = defineEmits<Emits>()

// 内部值，用于处理v-model
const internalValue = ref(props.modelValue)

// 监听外部值变化
watch(() => props.modelValue, (newValue) => {
  internalValue.value = newValue
})

// 监听内部值变化，向外发射事件
watch(internalValue, (newValue) => {
  emit('update:modelValue', newValue)
})

const loading = ref(false)
const options = ref<any[]>([])
const currentPage = ref(1)
const total = ref(0)
const keyword = ref('')
const loadMoreTrigger = ref<HTMLElement>()

const displayOptions = computed(() => options.value)
const hasMore = computed(() => options.value.length < total.value)

// 获取选项的键值
const getOptionKey = (item: any) => {
  if (typeof props.optionKey === 'function') {
    return props.optionKey(item)
  }
  return item[props.optionKey]
}

// 获取选项的标签
const getOptionLabel = (item: any) => {
  if (typeof props.optionLabel === 'function') {
    return props.optionLabel(item)
  }
  return item[props.optionLabel]
}

// 获取选项的值
const getOptionValue = (item: any) => {
  if (typeof props.optionValue === 'function') {
    return props.optionValue(item)
  }
  return item[props.optionValue]
}

// 加载数据
const loadData = async (reset = false) => {
  if (loading.value) return
  
  loading.value = true
  try {
    const page = reset ? 1 : currentPage.value
    const result = await props.fetchData({
      page,
      pageSize: props.pageSize,
      keyword: keyword.value
    })
    
    if (reset) {
      options.value = result.data
      currentPage.value = 1
    } else {
      options.value.push(...result.data)
    }
    
    total.value = result.total
    currentPage.value = page + 1
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    loading.value = false
  }
}

// 加载更多
const loadMore = () => {
  if (hasMore.value && !loading.value) {
    loadData(false)
  }
}

// 处理远程搜索
const handleRemoteMethod = (query: string) => {
  keyword.value = query
  loadData(true)
}

// 处理下拉框显示状态变化
const handleVisibleChange = (visible: boolean) => {
  if (visible && options.value.length === 0) {
    loadData(true)
  }
  
  if (visible) {
    // 监听滚动事件
    nextTick(() => {
      const selectDropdown = document.querySelector('.el-select-dropdown__list')
      if (selectDropdown) {
        selectDropdown.addEventListener('scroll', handleScroll)
      }
    })
  } else {
    // 移除滚动事件监听
    const selectDropdown = document.querySelector('.el-select-dropdown__list')
    if (selectDropdown) {
      selectDropdown.removeEventListener('scroll', handleScroll)
    }
  }
}

// 处理滚动事件
const handleScroll = (event: Event) => {
  const target = event.target as HTMLElement
  const { scrollTop, scrollHeight, clientHeight } = target
  
  // 滚动到底部时加载更多
  if (scrollTop + clientHeight >= scrollHeight - 10 && hasMore.value && !loading.value) {
    loadMore()
  }
}

// 处理值变化
const handleChange = (value: any) => {
  emit('change', value)
}

// 处理清空
const handleClear = () => {
  emit('clear')
}

// 重置数据
const reset = () => {
  options.value = []
  currentPage.value = 1
  total.value = 0
  keyword.value = ''
}

// 刷新数据
const refresh = () => {
  reset()
  loadData(true)
}

// 暴露方法
defineExpose({
  reset,
  refresh,
  loadMore
})

// 监听 fetchData 变化，重新加载数据
watch(() => props.fetchData, () => {
  refresh()
}, { immediate: false })

// 组件挂载时自动加载初始数据
onMounted(() => {
  loadData(true)
})
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