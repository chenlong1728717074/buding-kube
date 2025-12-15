<template>
  <el-dialog
    v-model="visible"
    :width="width"
    :close-on-click-modal="false"
    destroy-on-close
    class="unified-dialog"
  >
    <div class="unified-dialog__header">
      <div class="unified-dialog__title">{{ title }}</div>
      <div v-if="subtitle" class="unified-dialog__subtitle">{{ subtitle }}</div>
    </div>
    <div class="unified-dialog__body">
      <slot />
    </div>
    <template #footer>
      <div class="unified-dialog__footer">
        <slot name="footer" />
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface Props {
  modelValue: boolean
  title: string
  subtitle?: string
  width?: string
}

const props = withDefaults(defineProps<Props>(), {
  width: '600px'
})

const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

const visible = ref(false)

watch(
  () => props.modelValue,
  v => { visible.value = v },
  { immediate: true }
)

watch(visible, v => emit('update:modelValue', v))
</script>

<style scoped>
.unified-dialog :deep(.el-dialog__header) { display: none; }
.unified-dialog :deep(.el-dialog__body) { padding: 0; }
.unified-dialog :deep(.el-dialog__footer) { padding: 12px 16px; }
.unified-dialog__header {
  background: #ffffff;
  color: var(--text-primary);
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
}
.unified-dialog__title { font-size: 15px; font-weight: 600; }
.unified-dialog__subtitle { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }
.unified-dialog__body { padding: 12px 16px; background: #fff; }
.unified-dialog__footer { background: #fff; border-top: 1px solid #e4e7ed; }
</style>
