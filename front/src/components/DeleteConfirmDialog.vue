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
        <div class="warning-icon">
          <el-icon :size="24" color="#f59e0b">
            <WarningFilled />
          </el-icon>
        </div>
        <div class="message-wrap">
          <p class="main-message">{{ message }}</p>
          <p v-if="showTargetName" class="target-name">
            目标对象：<strong>{{ itemName }}</strong>
          </p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="delete-confirm-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="primary"
          :loading="loading"
          @click="handleConfirm"
        >
          确定删除
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { WarningFilled } from '@element-plus/icons-vue'

interface Props {
  modelValue: boolean
  title?: string
  message?: string
  itemName?: string
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
  itemName: '',
  width: '440px',
  loading: false
})

const emit = defineEmits<Emits>()

const visible = ref(false)

const dialogWidth = computed(() => props.width || '440px')
const showTargetName = computed(
  () => Boolean(props.itemName) && !props.message.includes(props.itemName)
)

// 监听 modelValue 变化
watch(
  () => props.modelValue,
  (newVal) => {
    visible.value = newVal
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
  emit('confirm')
}
</script>

<style scoped>
.delete-confirm-content {
  padding: 6px 0;
}

.content-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.warning-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 1px;
}

.message-wrap {
  flex: 1;
  min-width: 0;
}

.main-message {
  margin: 0;
  font-size: 14px;
  color: #303133;
  line-height: 1.6;
}

.target-name {
  margin: 6px 0 0;
  font-size: 13px;
  color: #606266;
  line-height: 1.5;
  word-break: break-all;
}

.target-name strong {
  color: #2563eb;
  font-weight: 600;
}

.delete-confirm-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.delete-confirm-dialog .el-dialog__header) {
  padding: 16px 20px 12px 20px;
}

:deep(.delete-confirm-dialog .el-dialog__title) {
  font-size: 16px;
  font-weight: 600;
}

:deep(.el-dialog__body) {
  padding: 14px 20px 18px 20px;
}

:deep(.el-dialog__footer) {
  padding: 12px 20px 16px;
  border-top: none !important;
}

.delete-confirm-footer :deep(.el-button) {
  min-width: 84px;
}
</style>
