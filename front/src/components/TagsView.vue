<template>
  <div class="tags-view">
    <div class="tags-list">
      <el-dropdown
        v-for="tag in tagsStore.tags"
        :key="tag.key"
        trigger="contextmenu"
        @command="(cmd: string) => handleAction(cmd, tag)"
      >
        <el-tag
          :type="tag.key === tagsStore.activeKey ? 'success' : 'info'"
          closable
          class="tag-item"
          @close="tagsStore.close(router, tag)"
          @click="tagsStore.activate(router, tag)"
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
import { useRouter } from 'vue-router'
import { useTagsStore, type TagItem } from '@/stores/tags'

// 标签的增减和切换逻辑在 stores/tags.ts, 由路由的 afterEach 驱动
const router = useRouter()
const tagsStore = useTagsStore()

const handleAction = (command: string, tag: TagItem) => {
  switch (command) {
    case 'refresh':
      tagsStore.refresh(router, tag)
      break
    case 'close-left':
      tagsStore.closeLeft(tag)
      break
    case 'close-right':
      tagsStore.closeRight(tag)
      break
    case 'close-all':
      tagsStore.closeAll(router)
      break
  }
}
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
