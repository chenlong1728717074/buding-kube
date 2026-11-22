import { useRouter } from 'vue-router'
import { useTagsStore } from '@/stores/tags'

// 详情页的返回按钮: 回到当前标签内的上一页(通常是列表页)
export function useTabBack() {
  const router = useRouter()
  const tagsStore = useTagsStore()
  return () => tagsStore.back(router)
}
