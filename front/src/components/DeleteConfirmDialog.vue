<template>
  <el-dialog
    v-model="visible"
    :title="title"
    :width="dialogWidth"
    :before-close="handleClose"
    class="delete-confirm-dialog"
  >
    <div class="delete-confirm-content">
      <div class="content-row">
        <div class="danger-icon">
          <el-icon :size="20" color="#ffffff">
            <CloseBold />
          </el-icon>
        </div>
        <p class="main-message">{{ message }}</p>
      </div>

      <div class="input-section">
        <p class="input-label">请输入 <strong>{{ displayName }}</strong> 来确认删除：</p>
        <el-input
          v-model="confirmInput"
          :placeholder="`请输入 ${displayName}`"
          @keyup.enter="handleConfirm"
          ref="inputRef"
        />
      </div>
    </div>
    
    <template #footer>
      <div class="delete-confirm-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="danger"
          :disabled="!canConfirm"
          :loading="loading"
          @click="handleConfirm"
        >
          确定
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import { CloseBold } from '@element-plus/icons-vue'

interface Props {
  modelValue: boolean
  title?: string
  message?: string
  itemName: string
  width?: string
  loading?: boolean
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm'): void
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  title: '确认删除',
  message: '确定要删除此项吗？',
  width: '600px',
  loading: false
})

const emit = defineEmits<Emits>()

const visible = ref(false)
const confirmInput = ref('')
const inputRef = ref()

const dialogWidth = computed(() => props.width || '600px')
const displayName = computed(() => props.itemName || '资源名称')
const canConfirm = computed(() => Boolean(props.itemName) && confirmInput.value === props.itemName)

// 监听 modelValue 变化
watch(
  () => props.modelValue,
  (newVal) => {
    visible.value = newVal
    if (newVal) {
      confirmInput.value = ''
      // 延迟聚焦到输入框
      nextTick(() => {
        inputRef.value?.focus()
      })
    }
  },
  { immediate: true }
)

// 监听 visible 变化
watch(visible, (newVal) => {
  emit('update:modelValue', newVal)
})

const handleClose = () => {
  visible.value = false
  emit('cancel')
}

const handleConfirm = () => {
  if (canConfirm.value) {
    emit('confirm')
  }
}
</script>

<style scoped>
.delete-confirm-content {
  padding: 0;
}

.content-row {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 16px;
}

.danger-icon {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #F56C6C;
  display: flex;
  align-items: center;
  justify-content: center;
}

.main-message {
  margin: 0;
  font-size: 16px;
  color: #303133;
  line-height: 1.5;
  padding-top: 4px;
}

.input-section {
  text-align: left;
}

.input-label {
  margin-bottom: 10px;
  font-size: 15px;
  color: #606266;
  line-height: 1.4;
}

.input-label strong {
  color: #303133;
  font-weight: 600;
}

.input-section :deep(.el-input__wrapper) {
  padding: 10px 12px;
}

.input-section :deep(.el-input__inner) {
  font-size: 15px;
}

.delete-confirm-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

:deep(.delete-confirm-dialog .el-dialog__header) {
  padding: 18px 24px 14px 24px;
}

:deep(.delete-confirm-dialog .el-dialog__title) {
  font-size: 18px;
  font-weight: 700;
}

:deep(.el-dialog__body) {
  padding: 18px 24px 18px 24px;
}

:deep(.el-dialog__footer) {
  padding: 16px 24px;
  background: #f3f4f6;
  border-top: 1px solid #e5e7eb;
}

.delete-confirm-footer :deep(.el-button) {
  min-width: 96px;
}
</style>
