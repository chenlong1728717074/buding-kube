import { computed, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

// 前端分页: 详情页里的表格数据一次性返回, 在前端切分
export function useClientPage<T>(source: MaybeRefOrGetter<T[] | null | undefined>, defaultSize = 10) {
  const list = computed<T[]>(() => toValue(source) || [])
  const page = ref(1)
  const pageSize = ref(defaultSize)
  const total = computed(() => list.value.length)
  const rows = computed(() => list.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value))

  // 数据变少(例如刷新后)导致当前页越界时回到最后一页
  watch([total, pageSize], () => {
    const last = Math.max(1, Math.ceil(total.value / pageSize.value))
    if (page.value > last) page.value = last
  })

  return { page, pageSize, total, rows }
}
