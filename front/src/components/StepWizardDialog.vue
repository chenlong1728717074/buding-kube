<template>
  <el-dialog
    v-model="visible"
    :title="title"
    :width="width"
    :close-on-click-modal="false"
    destroy-on-close
    class="step-wizard-dialog"
  >
    <div class="header-line">
      <div class="title">{{ title }}</div>
      <div class="actions">
        <el-switch v-model="yamlMode" active-text="编辑 YAML" inactive-text="编辑 YAML" />
      </div>
    </div>
    <div class="steps" v-if="!yamlMode">
      <div v-for="(s, i) in steps" :key="s.key" class="step" :class="{ active: i === activeIndex }" @click="setStep(i)">
        <span class="dot" />
        <span class="label">{{ s.label }}</span>
        <span class="status">{{ i === activeIndex ? '当前' : '未设置' }}</span>
      </div>
    </div>
    <div class="body">
      <slot v-if="!yamlMode" />
      <YamlEditor v-else v-model="internalYaml" :readonly="false" height="500px" />
    </div>
    <template #footer>
      <div class="footer-bar">
        <el-button @click="emitClose">取消</el-button>
        <el-button v-if="!yamlMode && activeIndex < steps.length - 1" type="primary" @click="next">下一步</el-button>
        <el-button v-else type="primary" :loading="loading" @click="confirm">{{ confirmText }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import YamlEditor from '@/components/YamlEditor.vue'

interface StepItem { key: string; label: string }
interface Props {
  modelValue: boolean
  title: string
  width?: string
  steps: StepItem[]
  active?: number
  loading?: boolean
  confirmText?: string
  yaml?: string
}

const props = withDefaults(defineProps<Props>(), {
  width: '800px',
  active: 0,
  loading: false,
  confirmText: '创建',
  yaml: ''
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:active': [index: number]
  'update:yaml': [yaml: string]
  'confirm': []
}>()

const visible = ref(false)
const activeIndex = ref(props.active)
const yamlMode = ref(false)
const internalYaml = ref(props.yaml)

watch(() => props.modelValue, v => (visible.value = v), { immediate: true })
watch(visible, v => emit('update:modelValue', v))
watch(() => props.active, v => (activeIndex.value = v))
watch(() => props.yaml, v => (internalYaml.value = v))
watch(internalYaml, v => emit('update:yaml', v))

const setStep = (i: number) => { activeIndex.value = i; emit('update:active', i) }
const next = () => setStep(Math.min(activeIndex.value + 1, props.steps.length - 1))
const emitClose = () => { visible.value = false }
const confirm = () => emit('confirm')

const confirmText = computed(() => props.confirmText)
const steps = computed(() => props.steps)
</script>

<style scoped>
.step-wizard-dialog :deep(.el-dialog__header) { padding: 0; }
.header-line { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; background: #fff; border-bottom: 1px solid rgba(59,130,246,0.08); }
.title { font-size: 15px; font-weight: 600; color: #2E3A59; }
.actions :deep(.el-switch) { --el-switch-on-color: #e5e7eb; --el-switch-off-color: #e5e7eb; }
.steps { display: flex; gap: 16px; padding: 8px 16px; border-bottom: 1px solid rgba(59,130,246,0.08); background: #fff; }
.step { display: flex; align-items: center; gap: 8px; color: #8A93A6; cursor: pointer; }
.step.active { color: #2E3A59; }
.dot { width: 8px; height: 8px; border-radius: 50%; background: #9ca3af; }
.step.active .dot { background: #22c55e; }
.label { font-weight: 500; }
.status { font-size: 12px; }
.body { padding: 12px 16px; }
.footer-bar { display: flex; justify-content: flex-end; gap: 8px; padding: 12px 16px; }
</style>
