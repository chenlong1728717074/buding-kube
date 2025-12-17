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
.kv-split { display: grid; grid-template-columns: 340px 1fr; gap: 16px; }
.kv-left { display: flex; flex-direction: column; border: 1px solid #cfe5ff; border-radius: 12px; background: #f0f9ff; }
.kv-list { padding: 6px; }
.kv-item { display: grid; grid-template-columns: 20px 1fr auto; align-items: center; gap: 8px; padding: 10px 12px; border-radius: 10px; cursor: pointer; }
.kv-item.active { background: #eef6ff; }
.kv-key { color: #1e3a8a; font-weight: 600; max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.kv-icon { color: #3b82f6; }
.kv-right { display: flex; flex-direction: column; gap: 10px; min-height: 0; }
.kv-value { flex: 1; }
.kv-value :deep(.el-textarea) { height: 100%; }
.kv-value :deep(.el-textarea__inner) { height: 100% !important; resize: none; }
.kv-right-header { display: flex; }
.kv-right-actions { display: flex; justify-content: space-between; align-items: center; gap: 10px; }
.kv-right-actions__buttons { display: flex; gap: 8px; }
.kv-right-empty { border: 1px dashed #cfe5ff; border-radius: 12px; padding: 18px; background: #ffffff; justify-content: center; }
.kv-empty-title { font-weight: 700; color: #0f172a; margin-bottom: 6px; }
.kv-empty-sub { color: #64748b; }
</style>
