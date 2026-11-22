<template>
  <el-dialog
    :model-value="modelValue"
    :width="width"
    top="4vh"
    :close-on-click-modal="false"
    :before-close="handleBeforeClose"
    class="yaml-view-dialog"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
  >
    <template #header>
      <div class="yv-header">
        <span class="yv-header__icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="m16 18 6-6-6-6" />
            <path d="m8 6-6 6 6 6" />
          </svg>
        </span>
        <div class="yv-header__text">
          <h3 class="yv-header__title">{{ title }}</h3>
          <p v-if="subtitle" class="yv-header__subtitle">{{ subtitle }}</p>
        </div>
        <span class="yv-header__mode" :class="isReadonly ? 'is-readonly' : dirty ? 'is-dirty' : 'is-editable'">
          {{ isReadonly ? '只读' : dirty ? '已修改' : '可编辑' }}
        </span>
      </div>
    </template>

    <div class="yv-body">
      <div v-if="visibleInfo.length" class="yv-info">
        <div v-for="item in visibleInfo" :key="item.label" class="yv-info__item">
          <span class="yv-info__label">{{ item.label }}</span>
          <span class="yv-info__value">{{ item.value }}</span>
        </div>
      </div>

      <div v-if="$slots.extra" class="yv-extra">
        <slot name="extra" />
      </div>

      <div v-loading="loading" class="yv-editor">
        <YamlEditor
          :model-value="local"
          :readonly="isReadonly"
          :title="editorTitle"
          :filename="filename"
          height="100%"
          @update:model-value="handleInput"
        />
      </div>
    </div>

    <template #footer>
      <div class="yv-footer">
        <span v-if="dirty" class="yv-footer__dirty">
          <i class="yv-footer__dot" />
          有未保存的修改
        </span>
        <span v-if="tip" class="yv-footer__tip">{{ tip }}</span>
        <div class="yv-footer__actions">
          <el-button @click="handleCancel">{{ isReadonly ? '关闭' : '取消' }}</el-button>
          <el-button v-if="!isReadonly" type="primary" :loading="submitting || saving" @click="handleConfirm">
            {{ confirmText || (applyTarget ? '保存' : '应用') }}
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { resourceApi, type ResourceYamlTarget } from '@/api/resource'

// 全站统一的 YAML 查看 / 编辑 / 添加弹窗. 三种用法:
//  1. 编辑已有资源: 传 apply-target, 弹窗自己调用保存接口, 成功后触发 @applied(页面刷新列表即可)
//  2. 页面自己处理提交: 传 :readonly="false" 并监听 @confirm(如 YAML 添加)
//  3. 只读查看: 什么都不传
// 额外的表单(如选择命名空间)放到 #extra 插槽
export interface YamlInfoItem {
  label: string
  value?: string | number | null
}

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    content?: string
    title?: string
    subtitle?: string
    info?: YamlInfoItem[]
    readonly?: boolean
    loading?: boolean
    submitting?: boolean
    confirmText?: string
    editorTitle?: string
    filename?: string
    tip?: string
    width?: string
    // 编辑的目标资源; 传入后弹窗可编辑, 并由弹窗负责保存
    applyTarget?: ResourceYamlTarget | null
  }>(),
  {
    content: '',
    title: '查看YAML',
    subtitle: '',
    info: () => [],
    readonly: true,
    loading: false,
    submitting: false,
    confirmText: '',
    editorTitle: 'YAML配置',
    filename: 'config.yaml',
    tip: '',
    width: 'min(1560px, 94vw)',
    applyTarget: null
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:content': [value: string]
  change: [value: string]
  confirm: []
  applied: []
}>()

const isReadonly = computed(() => (props.applyTarget ? false : props.readonly))

// 编辑器里的内容, 以及打开时的原始内容(用于判断是否有未保存的修改)
const local = ref(props.content ?? '')
const original = ref(props.content ?? '')
const saving = ref(false)
const dirty = computed(() => !isReadonly.value && local.value !== original.value)

const resetContent = (value: string) => {
  local.value = value
  original.value = value
}

watch(
  () => props.content,
  value => {
    // 编辑时页面通过 v-model:content 回写的值, 不当成新内容
    if ((value ?? '') !== local.value) resetContent(value ?? '')
  }
)

watch(
  () => props.modelValue,
  open => {
    if (open) resetContent(props.content ?? '')
  }
)

const handleInput = (value: string) => {
  local.value = value
  emit('update:content', value)
  emit('change', value)
}

const visibleInfo = computed(() =>
  props.info.map(item => ({
    label: item.label,
    value: item.value === undefined || item.value === null || item.value === '' ? '-' : String(item.value)
  }))
)

const close = () => emit('update:modelValue', false)

// 有未保存的修改时, 关闭前确认一次
const confirmDiscard = async () => {
  if (!dirty.value) return true
  try {
    await ElMessageBox.confirm('修改的内容还没有保存，确定要关闭吗？', '放弃修改', {
      type: 'warning',
      confirmButtonText: '放弃修改',
      cancelButtonText: '继续编辑'
    })
    return true
  } catch {
    return false
  }
}

const handleBeforeClose = async (done: () => void) => {
  if (await confirmDiscard()) done()
}

const handleCancel = async () => {
  if (await confirmDiscard()) close()
}

const handleConfirm = async () => {
  if (!props.applyTarget) {
    emit('confirm')
    return
  }
  if (!local.value.trim()) {
    ElMessage.warning('YAML 内容不能为空')
    return
  }
  saving.value = true
  try {
    await resourceApi.updateYaml({ ...props.applyTarget, yaml: local.value })
    ElMessage.success('保存成功')
    original.value = local.value
    emit('applied')
    close()
  } catch {
    // 请求拦截器已提示错误(包含 apiserver 返回的原因, 例如修改了不可变字段)
  } finally {
    saving.value = false
  }
}
</script>

<!-- el-dialog 会被移动到 body 下, scoped 样式选不到, 这里用带前缀的全局样式 -->
<style>
.el-dialog.yaml-view-dialog {
  --yv-primary: #3b74f2;
  display: flex;
  flex-direction: column;
  max-width: calc(100vw - 48px);
  height: min(92vh, 1200px);
  margin-bottom: 0;
  padding: 0 !important;
  border: 1px solid #e8eefa;
  border-radius: 18px !important;
  background: #ffffff;
  box-shadow: 0 24px 64px rgba(30, 64, 175, 0.16), 0 2px 8px rgba(15, 23, 42, 0.06) !important;
  overflow: hidden;
}

/* ---------- 头部 ---------- */
.el-dialog.yaml-view-dialog .el-dialog__header {
  position: relative;
  margin: 0 !important;
  padding: 14px 64px 13px 24px !important;
  border-bottom: 1px solid #edf1f7 !important;
  background: linear-gradient(180deg, #fafcff 0%, #ffffff 100%);
}

.el-dialog.yaml-view-dialog .el-dialog__header::after {
  content: '';
  position: absolute;
  left: 24px;
  bottom: -1px;
  width: 40px;
  height: 2px;
  border-radius: 2px;
  background: var(--yv-primary);
}

.yaml-view-dialog .yv-header {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.yaml-view-dialog .yv-header__icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 30px;
  height: 30px;
  border-radius: 9px;
  background: linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%);
  color: #ffffff;
  box-shadow: 0 4px 10px rgba(59, 116, 242, 0.3);
}

.yaml-view-dialog .yv-header__icon svg {
  width: 16px;
  height: 16px;
}

.yaml-view-dialog .yv-header__text {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 2px 12px;
  min-width: 0;
}

.yaml-view-dialog .yv-header__title {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  line-height: 26px;
  color: #1e293b;
}

.yaml-view-dialog .yv-header__subtitle {
  margin: 0;
  font-size: 13px;
  color: #94a3b8;
  word-break: break-all;
}

.yaml-view-dialog .yv-header__mode {
  flex: none;
  margin-left: auto;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.yaml-view-dialog .yv-header__mode.is-readonly {
  background: #f1f5f9;
  color: #64748b;
}

.yaml-view-dialog .yv-header__mode.is-editable {
  background: #eaf2ff;
  color: #2563eb;
}

.yaml-view-dialog .yv-header__mode.is-dirty {
  background: #fff4e5;
  color: #d97706;
}

.el-dialog.yaml-view-dialog .el-dialog__headerbtn {
  top: 10px;
  right: 12px;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  transition: background 0.15s ease;
}

.el-dialog.yaml-view-dialog .el-dialog__headerbtn:hover {
  background: #eff5ff;
}

.el-dialog.yaml-view-dialog .el-dialog__headerbtn .el-dialog__close {
  font-size: 18px;
  color: var(--yv-primary);
}

/* ---------- 内容: 编辑器撑满剩余高度 ---------- */
.el-dialog.yaml-view-dialog .el-dialog__body {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  padding: 0 !important;
  overflow: hidden;
}

.yaml-view-dialog .yv-body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  padding: 14px 20px 18px;
  background: #f8fbff;
}

.yaml-view-dialog .yv-info {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.yaml-view-dialog .yv-info__item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  padding: 4px 12px;
  border: 1px solid #e3ecfb;
  border-radius: 999px;
  background: #ffffff;
  font-size: 13px;
  line-height: 20px;
}

.yaml-view-dialog .yv-info__label {
  flex: none;
  color: #94a3b8;
}

.yaml-view-dialog .yv-info__value {
  overflow: hidden;
  color: #1e293b;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.yaml-view-dialog .yv-extra .el-form-item {
  margin-bottom: 0;
}

.yaml-view-dialog .yv-extra .el-select__wrapper,
.yaml-view-dialog .yv-extra .el-input__wrapper {
  border-radius: 8px !important;
}

.yaml-view-dialog .yv-editor {
  display: flex;
  flex: 1;
  min-height: 0;
  border-radius: 12px;
}

.yaml-view-dialog .yv-editor > .yaml-editor {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  height: 100%;
}


/* ---------- 底部 ---------- */
.el-dialog.yaml-view-dialog .el-dialog__footer {
  padding: 12px 20px !important;
  border-top: 1px solid #edf1f7;
  background: #ffffff;
}

.yaml-view-dialog .yv-footer {
  display: flex;
  align-items: center;
  gap: 12px;
}

.yaml-view-dialog .yv-footer__tip {
  font-size: 12px;
  color: #94a3b8;
}

.yaml-view-dialog .yv-footer__dirty {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #d97706;
}

.yaml-view-dialog .yv-footer__dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #f59e0b;
}

.yaml-view-dialog .yv-footer__actions {
  display: flex;
  gap: 12px;
  margin-left: auto;
}

.yaml-view-dialog .yv-footer__actions .el-button {
  min-width: 72px;
  height: 34px;
  margin: 0 !important;
  padding: 0 18px !important;
  border-radius: 10px !important;
  font-weight: 500;
}

.yaml-view-dialog .yv-footer__actions .el-button--default {
  color: #334155 !important;
  background: #ffffff !important;
  border: 1px solid #d5dce8 !important;
}

.yaml-view-dialog .yv-footer__actions .el-button--default:hover {
  color: #2563eb !important;
  background: #f5f9ff !important;
  border-color: #b9cdf3 !important;
}

.yaml-view-dialog .yv-footer__actions .el-button--primary {
  color: #ffffff !important;
  background: linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%) !important;
  border: none !important;
  box-shadow: 0 6px 16px rgba(59, 116, 242, 0.32) !important;
}

@media (max-width: 768px) {
  .el-dialog.yaml-view-dialog {
    height: 92vh;
  }

  .yaml-view-dialog .yv-body {
    padding: 12px;
  }
}
</style>
