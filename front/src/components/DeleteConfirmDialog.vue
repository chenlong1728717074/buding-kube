<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="400px"
    :before-close="handleClose"
    center
  >
    <div class="delete-confirm-content">
      <div class="content-header">
        <el-icon size="32" color="#F56C6C" class="warning-icon">
          <WarningFilled />
        </el-icon>
        <div class="warning-text">
          <p class="main-message">{{ message }}</p>
          <p class="danger-text">此操作不可恢复！</p>
        </div>
      </div>
      
      <div class="input-section">
        <p class="input-label">请输入 <strong>{{ itemName }}</strong> 来确认删除：</p>
        <el-input
          v-model="confirmInput"
          :placeholder="`请输入 ${itemName}`"
          @keyup.enter="handleConfirm"
          ref="inputRef"
          size="small"
        />
      </div>
    </div>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="danger"
          :disabled="confirmInput !== itemName"
          :loading="loading"
          @click="handleConfirm"
        >
          确认删除
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { WarningFilled } from '@element-plus/icons-vue'

interface Props {
  modelValue: boolean
  title?: string
  message?: string
  itemName: string
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
  loading: false
})

const emit = defineEmits<Emits>()

const visible = ref(false)
const confirmInput = ref('')
const inputRef = ref()

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
  if (confirmInput.value === props.itemName) {
    emit('confirm')
  }
}
</script>

<style scoped>
.delete-confirm-content {
  padding: 8px 0;
}

.content-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.warning-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.warning-text {
  flex: 1;
}

.main-message {
  margin: 0 0 4px 0;
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.danger-text {
  margin: 0;
  font-size: 12px;
  color: #F56C6C !important;
  font-weight: 400;
}

.input-section {
  text-align: left;
}

.input-label {
  margin-bottom: 8px;
  font-size: 13px;
  color: #606266;
  line-height: 1.4;
}

.input-label strong {
  color: #303133;
  font-weight: 600;
}

.dialog-footer {
  text-align: right;
  padding-top: 8px;
}

:deep(.el-dialog__body) {
  padding: 16px 20px 8px 20px;
}

:deep(.el-dialog__footer) {
  padding: 8px 20px 16px 20px;
}
</style>