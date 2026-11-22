<template>
  <div class="kv-split" :style="{ height }">
    <div class="kv-left">
      <el-scrollbar class="kv-list">
        <div
          v-for="(item, idx) in internalRows"
          :key="idx + '-' + item.key"
          class="kv-item"
          :class="{ active: idx === selected }"
          @click="select(idx)"
        >
          <el-icon class="kv-icon"><Document /></el-icon>
          <span class="kv-key" :title="item.key || '(未命名)'">{{ item.key || '(未命名)' }}</span>
          <el-button link type="danger" size="small" @click.stop="remove(idx)">删除</el-button>
        </div>
      </el-scrollbar>
    </div>
    <div class="kv-right" v-if="hasCurrent">
      <div class="kv-right-header">
        <el-input v-model="current.key" placeholder="键名" />
      </div>
      <el-input v-model="current.value" class="kv-value" type="textarea" placeholder="值内容" />
      <div v-if="createMode" class="kv-right-actions">
        <el-text v-if="draftKeyError" type="danger">{{ draftKeyError }}</el-text>
        <div class="kv-right-actions__buttons">
          <el-button size="small" @click="cancelCreate">取消</el-button>
          <el-button size="small" type="primary" :disabled="!canConfirmCreate" @click="confirmCreate">确认</el-button>
        </div>
      </div>
    </div>
    <div class="kv-right kv-right-empty" v-else>
      <div class="kv-empty-title">暂无键值</div>
      <div class="kv-empty-sub">点击上方“添加数据”开始编辑</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import { Document } from '@element-plus/icons-vue'

interface Row { key: string; value: string }
interface Props {
  modelValue: Row[]
  height?: string
}

const props = withDefaults(defineProps<Props>(), {
  height: '360px'
})

const emit = defineEmits<{ 'update:modelValue': [rows: Row[]] }>()

const internalRows = ref<Row[]>([])
const selected = ref(-1)
const createMode = ref(false)
const draft = ref<Row>({ key: '', value: '' })
const prevSelected = ref(-1)
let syncingFromProps = false

watch(() => props.modelValue, async (v) => {
  syncingFromProps = true
  internalRows.value = Array.isArray(v) ? v.map(x => ({ ...x })) : []
  if (createMode.value) {
    createMode.value = false
    draft.value = { key: '', value: '' }
    prevSelected.value = -1
  }
  if (internalRows.value.length === 0) {
    selected.value = -1
    await nextTick()
    syncingFromProps = false
    return
  }
  selected.value = Math.min(Math.max(selected.value, 0), internalRows.value.length - 1)
  await nextTick()
  syncingFromProps = false
}, { immediate: true })

watch(internalRows, (v) => {
  if (syncingFromProps) return
  emit('update:modelValue', v.map(x => ({ ...x })))
}, { deep: true })

const select = (idx: number) => {
  if (createMode.value) {
    createMode.value = false
    draft.value = { key: '', value: '' }
    prevSelected.value = -1
  }
  selected.value = idx
}
const hasCurrent = computed(() => createMode.value || (selected.value >= 0 && selected.value < internalRows.value.length))
const current = computed({
  get: () => (createMode.value ? draft.value : (internalRows.value[selected.value] ?? { key: '', value: '' })),
  set: (val: Row) => {
    if (createMode.value) {
      draft.value = val
      return
    }
    if (selected.value < 0 || selected.value >= internalRows.value.length) return
    internalRows.value[selected.value] = val
  }
})

const remove = (idx: number) => {
  internalRows.value.splice(idx, 1)
  if (internalRows.value.length === 0) {
    selected.value = -1
    return
  }
  selected.value = Math.min(selected.value, internalRows.value.length - 1)
}

const addBlank = () => {
  prevSelected.value = selected.value
  createMode.value = true
  draft.value = { key: '', value: '' }
}

const draftKeyTrimmed = computed(() => (draft.value.key || '').trim())
const draftKeyExists = computed(() => internalRows.value.some(r => r.key === draftKeyTrimmed.value))
const canConfirmCreate = computed(() => draftKeyTrimmed.value.length > 0 && !draftKeyExists.value)
const draftKeyError = computed(() => {
  if (!createMode.value) return ''
  if (draftKeyTrimmed.value.length === 0) return '键名不能为空'
  if (draftKeyExists.value) return '键名已存在'
  return ''
})

const cancelCreate = () => {
  createMode.value = false
  draft.value = { key: '', value: '' }
  if (prevSelected.value >= 0 && prevSelected.value < internalRows.value.length) {
    selected.value = prevSelected.value
  } else if (internalRows.value.length === 0) {
    selected.value = -1
  } else {
    selected.value = Math.min(Math.max(selected.value, 0), internalRows.value.length - 1)
  }
  prevSelected.value = -1
}

const confirmCreate = () => {
  if (!createMode.value) return
  if (!canConfirmCreate.value) return
  internalRows.value.push({ key: draftKeyTrimmed.value, value: draft.value.value || '' })
  selected.value = internalRows.value.length - 1
  createMode.value = false
  draft.value = { key: '', value: '' }
  prevSelected.value = -1
}

defineExpose({
  addBlank
})
</script>

<style scoped>
.kv-split {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 16px;
}

/* 左侧键列表 */
.kv-left {
  display: flex;
  flex-direction: column;
  min-height: 0;
  border: 1px solid #e3ecfb;
  border-radius: 12px;
  background: #f8fbff;
  overflow: hidden;
}

.kv-list {
  padding: 8px;
}

.kv-item {
  position: relative;
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
  padding: 10px 12px;
  border: 1px solid transparent;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.15s ease, border-color 0.15s ease;
}

.kv-item:hover {
  background: #eef4ff;
}

.kv-item.active {
  border-color: #d4e2fb;
  background: #ffffff;
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.08);
}

.kv-item.active::before {
  content: '';
  position: absolute;
  top: 10px;
  bottom: 10px;
  left: 0;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: #3b74f2;
}

.kv-icon {
  color: #94a3b8;
}

.kv-item.active .kv-icon {
  color: #3b74f2;
}

.kv-key {
  overflow: hidden;
  color: #334155;
  font-family: 'SF Mono', Consolas, Monaco, monospace;
  font-size: 13px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.kv-item.active .kv-key {
  color: #1d4ed8;
  font-weight: 600;
}

/* 删除按钮只在悬停 / 选中时显示, 列表更干净 */
.kv-item :deep(.el-button) {
  opacity: 0;
  transition: opacity 0.15s ease;
}

.kv-item:hover :deep(.el-button),
.kv-item.active :deep(.el-button),
.kv-item :deep(.el-button:focus-visible) {
  opacity: 1;
}

/* 右侧编辑区 */
.kv-right {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
  min-height: 0;
}

.kv-right-header {
  display: flex;
}

.kv-right-header :deep(.el-input__wrapper) {
  border-radius: 10px !important;
}

.kv-right-header :deep(.el-input__inner) {
  font-family: 'SF Mono', Consolas, Monaco, monospace;
  font-weight: 600;
}

.kv-value {
  flex: 1;
  min-height: 0;
}

.kv-value :deep(.el-textarea) {
  height: 100%;
}

.kv-value :deep(.el-textarea__inner) {
  height: 100% !important;
  padding: 12px 14px;
  border-radius: 10px !important;
  background: #fbfcfe;
  font-family: 'SF Mono', Consolas, Monaco, monospace;
  font-size: 13px;
  line-height: 1.65;
  resize: none;
}

.kv-right-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.kv-right-actions__buttons {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.kv-right-actions__buttons :deep(.el-button) {
  border-radius: 8px !important;
}

/* 空状态 */
.kv-right-empty {
  align-items: center;
  justify-content: center;
  padding: 18px;
  border: 1px dashed #cbdaf3;
  border-radius: 12px;
  background: #fbfcff;
  text-align: center;
}

.kv-empty-title {
  margin-bottom: 6px;
  font-size: 15px;
  font-weight: 600;
  color: #334155;
}

.kv-empty-sub {
  font-size: 13px;
  color: #94a3b8;
}

@media (max-width: 900px) {
  .kv-split {
    grid-template-columns: 1fr;
    height: auto !important;
  }

  .kv-left {
    max-height: 220px;
  }

  .kv-right {
    min-height: 320px;
  }
}
</style>
