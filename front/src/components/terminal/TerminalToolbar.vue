<template>
  <div class="tt-bar">
    <el-button-group size="small">
      <el-button title="减小字号" @click="view?.changeFontSize(-1)">A-</el-button>
      <el-button title="增大字号" @click="view?.changeFontSize(1)">A+</el-button>
    </el-button-group>
    <el-button size="small" title="清屏" @click="view?.clear()"><el-icon><Brush /></el-icon></el-button>
    <el-button size="small" @click="emit('reconnect')"><el-icon><RefreshRight /></el-icon>重新连接</el-button>
    <el-button v-if="fullscreen !== undefined" size="small" :title="fullscreen ? '退出全屏' : '全屏'" @click="emit('fullscreen')"><el-icon><FullScreen /></el-icon></el-button>
    <el-button v-if="closable" size="small" title="关闭" @click="emit('close')"><el-icon><Close /></el-icon></el-button>
  </div>
</template>

<script setup lang="ts">
import { Brush, Close, FullScreen, RefreshRight } from '@element-plus/icons-vue'
import type TerminalView from '@/components/terminal/TerminalView.vue'

withDefaults(
  defineProps<{
    view: InstanceType<typeof TerminalView> | null
    // 不传时不显示全屏按钮
    fullscreen?: boolean
    closable?: boolean
  }>(),
  { fullscreen: undefined, closable: true }
)

const emit = defineEmits<{ reconnect: []; fullscreen: []; close: [] }>()
</script>

<style scoped>
.tt-bar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tt-bar .el-button + .el-button {
  margin-left: 0;
}
</style>
