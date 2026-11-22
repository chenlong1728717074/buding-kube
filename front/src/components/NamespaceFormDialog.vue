<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="800px"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--compact"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="initForm"
    @closed="formRef?.clearValidate()"
  >
    <template #header>
      <div class="dialog-header">
        <div>
          <h3 class="dialog-title">{{ title }}</h3>
        </div>
      </div>
    </template>
    <div class="config-editor">
      <div class="config-content">
        <el-form ref="formRef" :model="form" :rules="rules" label-width="96px" class="namespace-form">
          <el-form-item label="命名空间" prop="namespace">
            <el-input
              v-model="form.namespace"
              placeholder="请输入命名空间名称，如：my-app"
              :disabled="isEdit"
              maxlength="63"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="别名" prop="alias">
            <el-input v-model="form.alias" placeholder="请输入别名" maxlength="50" show-word-limit />
          </el-form-item>

          <el-form-item label="描述" prop="describe">
            <el-input
              v-model="form.describe"
              type="textarea"
              :rows="3"
              placeholder="请输入描述"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>
        </el-form>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { namespaceApi } from '@/api/namespace'

// 命名空间添加 / 编辑弹窗, 列表页和详情页共用
const props = withDefaults(
  defineProps<{
    modelValue: boolean
    clusterId: string
    // 传入时为编辑模式
    namespace?: { name: string; alias?: string; describe?: string } | null
  }>(),
  { namespace: null }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: []
}>()

const isEdit = computed(() => !!props.namespace)
const title = computed(() => (isEdit.value ? '编辑命名空间' : '添加命名空间'))

const formRef = ref<FormInstance>()
const submitting = ref(false)
const form = reactive({ namespace: '', alias: '', describe: '' })

const rules: FormRules = {
  namespace: [
    { required: true, message: '请输入命名空间名称', trigger: 'blur' },
    {
      pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
      message: '只能包含小写字母、数字和 "-"，且以字母或数字开头和结尾',
      trigger: 'blur'
    }
  ]
}

const initForm = () => {
  form.namespace = props.namespace?.name || ''
  form.alias = props.namespace?.alias || ''
  form.describe = props.namespace?.describe || ''
}

const handleSubmit = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  submitting.value = true
  try {
    const data = { clusterId: props.clusterId, namespace: form.namespace, alias: form.alias, describe: form.describe }
    if (isEdit.value) {
      await namespaceApi.update(data)
      ElMessage.success('更新成功')
    } else {
      await namespaceApi.create(data)
      ElMessage.success('创建成功')
    }
    emit('update:modelValue', false)
    emit('success')
  } catch {
    // 请求拦截器已提示错误
  } finally {
    submitting.value = false
  }
}
</script>
