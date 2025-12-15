<template>
  <div class="tags-view">
    <div class="tags-list">
      <el-dropdown
        v-for="tag in tags"
        :key="tag.fullPath"
        trigger="contextmenu"
        @command="(cmd:string)=>handleAction(cmd, tag)"
      >
        <el-tag
          :type="tag.active ? 'success' : 'info'"
          closable
          @close="closeTag(tag)"
          @click="activateTag(tag)"
          class="tag-item"
        >
          {{ tag.title }}
        </el-tag>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="refresh">刷新当前页面</el-dropdown-item>
            <el-dropdown-item command="close-left">关闭左边所有</el-dropdown-item>
            <el-dropdown-item command="close-right">关闭右边所有</el-dropdown-item>
            <el-dropdown-item command="close-all" divided>关闭所有</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowDown } from '@element-plus/icons-vue'

interface TagItem {
  title: string
  fullPath: string
  active: boolean
}

const route = useRoute()
const router = useRouter()
const tags = ref<TagItem[]>([])

const normalizeFullPath = (fp: string): string => {
  try {
    const url = new URL(fp, window.location.origin)
    const params = new URLSearchParams(url.search)
    params.delete('__refresh')
    const qs = params.toString()
    return url.pathname + (qs ? `?${qs}` : '')
  } catch {
    return fp
  }
}

const addTag = () => {
  const title = (route.meta.title as string) || route.name?.toString() || route.path
  const fullPath = normalizeFullPath(route.fullPath)
  const exists = tags.value.find(t => t.fullPath === fullPath)
  tags.value.forEach(t => (t.active = false))
  if (exists) {
    exists.active = true
  } else {
    tags.value.push({ title, fullPath, active: true })
  }
  persist()
}

const activateTag = (tag: TagItem) => {
  tags.value.forEach(t => (t.active = t.fullPath === tag.fullPath))
  router.push(tag.fullPath)
}

const closeTag = (tag: TagItem) => {
  const idx = tags.value.findIndex(t => t.fullPath === tag.fullPath)
  if (idx >= 0) {
    const wasActive = tags.value[idx].active
    tags.value.splice(idx, 1)
    if (wasActive) {
      const next = tags.value[idx - 1] || tags.value[idx] || null
      if (next) {
        activateTag(next)
      }
    }
  }
  persist()
}

const handleAction = (command: string, tag: TagItem) => {
  switch (command) {
    case 'refresh':
      refreshTag(tag)
      break
    case 'close-left':
      closeLeft(tag)
      break
    case 'close-right':
      closeRight(tag)
      break
    case 'close-all':
      closeAll()
      break
  }
}

const refreshTag = (tag: TagItem) => {
  try {
    const url = new URL(tag.fullPath, window.location.origin)
    const params = new URLSearchParams(url.search)
    params.set('__refresh', String(Date.now()))
    const newFull = url.pathname + (params.toString() ? `?${params.toString()}` : '')
    router.push(newFull)
  } catch {
    router.push({ path: tag.fullPath })
  }
}

const closeAll = () => {
  tags.value = []
  persist()
  router.push('/dashboard')
}

const closeLeft = (tag?: TagItem) => {
  const idx = tags.value.findIndex(t => (tag ? t.fullPath === tag.fullPath : t.active))
  if (idx <= 0) return
  tags.value.splice(0, idx)
  persist()
}

const closeRight = (tag?: TagItem) => {
  const idx = tags.value.findIndex(t => (tag ? t.fullPath === tag.fullPath : t.active))
  if (idx < 0) return
  tags.value.splice(idx + 1)
  persist()
}

const persist = () => {
  try {
    localStorage.setItem('tagsView', JSON.stringify(tags.value))
  } catch {}
}

const restore = () => {
  try {
    const raw = localStorage.getItem('tagsView')
    if (raw) {
      const arr = JSON.parse(raw) as TagItem[]
      tags.value = arr.map(t => ({ ...t, active: false }))
      // 激活当前路由对应的标签
      const cur = tags.value.find(t => t.fullPath === normalizeFullPath(route.fullPath))
      if (cur) cur.active = true
    }
  } catch {}
}

watch(() => route.fullPath, addTag, { immediate: true })
restore()
</script>

<style scoped>
.tags-view {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--gap-3) var(--gap-4);
  background: #ffffff;
  border-bottom: 1px solid rgba(59,130,246,0.12);
}

.tags-list {
  display: flex;
  gap: var(--gap-2);
  overflow-x: auto;
  white-space: nowrap;
}

.tag-item {
  cursor: pointer;
  padding: 8px 14px;
  font-size: 14px;
  border-radius: 999px;
  border-color: #cfe0ff !important;
}

.tag-item:hover {
  filter: brightness(1.02);
}


</style>
