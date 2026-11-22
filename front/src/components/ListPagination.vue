<template>
  <div class="list-footer">
    <div class="list-footer__total">
      <el-icon class="list-footer__info"><InfoFilled /></el-icon>
      <span>共 {{ total }} 条数据</span>
    </div>

    <div class="list-footer__pager">
      <el-pagination
        :current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        :pager-count="7"
        layout="prev, pager, next"
        @update:current-page="changePage"
      />
      <span class="list-footer__label">每页</span>
      <el-select :model-value="pageSize" class="list-footer__size" @update:model-value="changeSize">
        <el-option v-for="size in pageSizes" :key="size" :label="`${size} 条`" :value="size" />
      </el-select>
      <span class="list-footer__label">跳至</span>
      <el-input v-model="jumpValue" class="list-footer__jump" @keyup.enter="jump" @blur="jump" />
      <span class="list-footer__label">页</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { InfoFilled } from '@element-plus/icons-vue'

// 用法与 el-pagination 保持一致: v-model:current-page / v-model:page-size / @current-change / @size-change
const props = withDefaults(
  defineProps<{
    currentPage?: number
    pageSize?: number
    total?: number
    pageSizes?: number[]
  }>(),
  {
    currentPage: 1,
    pageSize: 10,
    total: 0,
    pageSizes: () => [10, 20, 50, 100]
  }
)

const emit = defineEmits<{
  'update:currentPage': [page: number]
  'update:pageSize': [size: number]
  'current-change': [page: number]
  'size-change': [size: number]
}>()

const pageCount = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

const jumpValue = ref(String(props.currentPage))
watch(
  () => props.currentPage,
  page => {
    jumpValue.value = String(page)
  }
)

const changePage = (page: number) => {
  if (page === props.currentPage) return
  emit('update:currentPage', page)
  emit('current-change', page)
}

const changeSize = (size: number) => {
  if (size === props.pageSize) return
  emit('update:pageSize', size)
  // 与 el-pagination 一致: 当前页超出新的总页数时回到最后一页(先更新页码, 再触发 size-change 让页面用正确的页码请求)
  const maxPage = Math.max(1, Math.ceil(props.total / size))
  if (props.currentPage > maxPage) {
    emit('update:currentPage', maxPage)
  }
  emit('size-change', size)
}

const jump = () => {
  const page = Number.parseInt(jumpValue.value, 10)
  if (Number.isNaN(page)) {
    jumpValue.value = String(props.currentPage)
    return
  }
  const target = Math.min(Math.max(page, 1), pageCount.value)
  jumpValue.value = String(target)
  changePage(target)
}
</script>

<style scoped>
.list-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px 16px;
  padding: 18px 12px 6px;
}

.list-footer__total {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #475569;
}

.list-footer__info {
  font-size: 16px;
  color: #3b82f6;
}

.list-footer__pager {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.list-footer__label {
  font-size: 14px;
  color: #475569;
}

.list-footer__size {
  width: 96px;
}

.list-footer__jump {
  width: 56px;
}

.list-footer :deep(.el-pagination) {
  padding: 0;
}

.list-footer :deep(.el-pagination button),
.list-footer :deep(.el-pager li) {
  min-width: 34px;
  height: 34px;
  margin: 0 4px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
  color: #475569;
  font-weight: 500;
  transition: all 0.15s ease;
}

.list-footer :deep(.el-pager li.is-active) {
  color: #ffffff;
  border-color: transparent;
  background: linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%);
  box-shadow: 0 4px 10px rgba(59, 116, 242, 0.3);
}

.list-footer :deep(.el-pager li:not(.is-active):hover),
.list-footer :deep(.el-pagination button:not(:disabled):hover) {
  color: #2563eb;
  border-color: #bfd4fb;
}

.list-footer :deep(.el-pagination button:disabled) {
  color: #cbd5e1;
  background: #f8fafc;
}

.list-footer :deep(.list-footer__size .el-select__wrapper) {
  min-height: 34px;
  border-radius: 8px !important;
}

.list-footer :deep(.list-footer__jump .el-input__wrapper) {
  height: 34px;
  border-radius: 8px !important;
}

.list-footer :deep(.list-footer__jump .el-input__inner) {
  text-align: center;
}

@media (max-width: 768px) {
  .list-footer {
    justify-content: center;
  }
}
</style>
