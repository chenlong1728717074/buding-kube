<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="640px"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--compact"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="init"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">{{ title }}</h3>
      </div>
    </template>
    <div class="config-editor">
      <div class="config-content">
        <el-form :model="form" label-width="80px">
          <el-form-item label="名称">
            <el-input :model-value="name" disabled />
          </el-form-item>
          <el-form-item label="别名">
            <el-input v-model="form.alias" placeholder="请输入别名" maxlength="50" show-word-limit />
          </el-form-item>
          <el-form-item label="描述">
            <el-input v-model="form.describe" type="textarea" :rows="3" placeholder="请输入描述" maxlength="200" show-word-limit />
          </el-form-item>
        </el-form>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'

// 编辑别名和描述的通用弹窗, 保存逻辑由调用方传入
const props = withDefaults(
  defineProps<{
    modelValue: boolean
    name: string
    alias?: string
    describe?: string
    title?: string
    submit: (alias: string, describe: string) => Promise<unknown>
  }>(),
  { alias: '', describe: '', title: '编辑信息' }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: []
}>()

const form = reactive({ alias: '', describe: '' })
const saving = ref(false)

const init = () => {
  form.alias = props.alias || ''
  form.describe = props.describe || ''
}

const save = async () => {
  saving.value = true
  try {
    await props.submit(form.alias.trim(), form.describe.trim())
    ElMessage.success('保存成功')
    emit('update:modelValue', false)
    emit('success')
  } catch {
    // 请求拦截器已提示错误
  } finally {
    saving.value = false
  }
}
</script>
