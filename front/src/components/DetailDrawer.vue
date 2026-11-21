<template>
  <el-drawer
    :model-value="modelValue"
    size="60%"
    direction="rtl"
    :with-header="false"
    :append-to-body="true"
    class="ks-drawer-root"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="ks-drawer">
      <!-- 头部: 资源类型 + 名称 + 元信息 -->
      <div class="ks-drawer-header">
        <div class="ks-drawer-head-main">
          <div class="ks-drawer-title-row">
            <span class="ks-kind">{{ kind }}</span>
            <span class="ks-name">{{ name }}</span>
          </div>
          <div v-if="meta?.length" class="ks-drawer-meta">
            <span v-for="m in meta" :key="m.label" class="ks-meta-item">
              <span class="ks-meta-label">{{ m.label }}</span>
              <span class="ks-meta-value">{{ m.value || '-' }}</span>
            </span>
          </div>
        </div>
        <el-button text circle class="ks-close" @click="emit('update:modelValue', false)">
          <el-icon :size="18"><Close /></el-icon>
        </el-button>
      </div>

      <el-divider class="ks-header-divider" />

      <!-- 内容区 -->
      <div class="ks-drawer-body">
        <slot />
      </div>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { Close } from '@element-plus/icons-vue'

// KubeSphere 风格资源详情抽屉: 右侧滑出, 大标题 + 元信息 + 自由内容区(Tab)
defineProps<{
  modelValue: boolean
  /** 资源类型, 如 ConfigMap / Secret */
  kind: string
  /** 资源名称 */
  name?: string
  /** 头部元信息行 */
  meta?: { label: string; value?: string }[]
}>()

const emit = defineEmits(['update:modelValue'])
</script>

<style scoped>
.ks-drawer {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fff;
}

.ks-drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 20px 24px 0;
}

.ks-drawer-head-main {
  min-width: 0;
}

.ks-drawer-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.ks-kind {
  flex-shrink: 0;
  font-size: 12px;
  color: #2563eb;
  background: #eff6ff;
  border: 1px solid #dbeafe;
  padding: 3px 10px;
  border-radius: 4px;
  font-weight: 500;
}

.ks-name {
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
  word-break: break-all;
}

.ks-drawer-meta {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px 24px;
}

.ks-meta-item {
  font-size: 13px;
}

.ks-meta-label {
  color: #9ca3af;
  margin-right: 6px;
}

.ks-meta-value {
  color: #4b5563;
}

.ks-close {
  flex-shrink: 0;
  color: #9ca3af;
}

.ks-header-divider {
  margin: 14px 0 0;
}

.ks-drawer-body {
  flex: 1;
  overflow: auto;
  padding: 8px 24px 24px;
}
</style>

<!-- 槽位内容由父页面编译, 共享样式需全局作用域 -->
<style>
.ks-drawer-root .el-drawer__body {
  padding: 0;
}

.ks-tabs .el-tabs__header {
  margin-bottom: 16px;
}

.ks-tabs .el-tabs__item {
  font-size: 14px;
}

.ks-data-hint {
  font-size: 12px;
  color: #9ca3af;
  margin-bottom: 12px;
}

/* 基本信息行 */
.ks-info-row {
  display: flex;
  padding: 11px 0;
  border-bottom: 1px dashed #f0f1f3;
}

.ks-info-row:last-child {
  border-bottom: none;
}

.ks-info-label {
  width: 120px;
  flex-shrink: 0;
  color: #9ca3af;
  font-size: 13px;
}

.ks-info-value {
  flex: 1;
  min-width: 0;
  color: #1f2937;
  font-size: 13px;
  word-break: break-all;
}

/* 标签/注解 芯片 */
.ks-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ks-chip {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  padding: 4px 10px;
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  font-size: 12px;
  color: #475569;
}

.ks-chip-key {
  color: #1f2937;
  margin-right: 4px;
}

.ks-chip-value {
  color: #6b7280;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* KV 数据卡片 */
.ks-kv-item {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 12px 14px;
  margin-bottom: 12px;
}

.ks-kv-key {
  font-weight: 600;
  font-size: 13px;
  color: #1f2937;
  margin-bottom: 8px;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  word-break: break-all;
}

.ks-kv-value {
  font-size: 12px;
  color: #475569;
  background: #f8fafc;
  border-radius: 6px;
  padding: 10px 12px;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 240px;
  overflow: auto;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  line-height: 1.6;
}

/* 分组标题 */
.ks-section-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
  margin: 4px 0 12px;
}
</style>
