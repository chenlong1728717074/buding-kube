<template>
  <el-dialog
    :model-value="modelValue"
    title="编辑存储类"
    width="800px"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--compact"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="initForm"
  >
    <template #header>
      <div class="dialog-header">
        <div>
          <h3 class="dialog-title">编辑存储类</h3>
        </div>
      </div>
    </template>
    <div class="config-editor">
      <div class="config-content">
        <!-- 创建后不可修改的设置, 只读展示 -->
        <div v-if="storageClass" class="sc-fixed">
          <div class="sc-fixed__item">
            <span class="sc-fixed__label">供应者</span>
            <code class="sc-fixed__value">{{ storageClass.provisioner || '-' }}</code>
          </div>
          <div class="sc-fixed__item">
            <span class="sc-fixed__label">回收策略</span>
            <span class="sc-fixed__value">{{ reclaimText }}</span>
          </div>
          <div class="sc-fixed__item">
            <span class="sc-fixed__label">绑定模式</span>
            <span class="sc-fixed__value">{{ bindingText }}</span>
          </div>
          <p class="sc-fixed__tip">以上设置创建后不能修改，如需调整请重新创建存储类</p>
        </div>

        <el-form :model="form" label-width="96px">
          <el-form-item label="名称">
            <el-input :model-value="storageClass?.name || ''" disabled />
          </el-form-item>
          <el-form-item label="别名">
            <el-input v-model="form.alias" placeholder="请输入别名" maxlength="50" show-word-limit />
          </el-form-item>
          <el-form-item label="存储类型">
            <el-select v-model="form.storageType" style="width: 100%">
              <el-option :label="autoLabel" value="" />
              <el-option v-for="opt in STORAGE_TYPE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value">
                <div class="sc-option">
                  <span>{{ opt.label }}</span>
                  <span class="sc-option__desc">{{ opt.desc }}</span>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="描述">
            <el-input v-model="form.describe" type="textarea" :rows="3" placeholder="请输入描述" maxlength="200" show-word-limit />
          </el-form-item>
          <el-form-item label="卷扩展">
            <el-switch v-model="form.allowVolumeExpansion" active-text="允许" inactive-text="禁止" inline-prompt />
          </el-form-item>
          <el-form-item label="卷克隆">
            <el-switch v-model="form.allowVolumeClone" active-text="允许" inactive-text="禁止" inline-prompt />
          </el-form-item>
          <el-form-item label="默认存储类">
            <el-switch v-model="form.isDefault" />
            <span class="sc-inline-hint">设为默认后会取消其他存储类的默认标记</span>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { storageClassApi, type StorageClassVO, type StorageTypeKey } from '@/api/storageclass'
import { STORAGE_TYPE_OPTIONS } from '@/utils/storageclass'

const props = defineProps<{
  modelValue: boolean
  clusterId: string
  storageClass: StorageClassVO | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: []
}>()

const submitting = ref(false)
const form = reactive({
  alias: '',
  describe: '',
  storageType: '' as StorageTypeKey,
  allowVolumeExpansion: false,
  allowVolumeClone: false,
  isDefault: false
})

// 「自动识别」选项里显示当前按供应者推断出的类型
const autoLabel = computed(() => {
  const sc = props.storageClass
  if (sc?.storageTypeAuto) return `自动识别（当前为${sc.storageType || '未分类'}）`
  return '自动识别（按供应者判断）'
})

const reclaimText = computed(() => {
  const policy = props.storageClass?.reclaimPolicy
  return policy === 'Retain' ? '保留（Retain）' : policy === 'Delete' ? '删除（Delete）' : policy || '-'
})

const bindingText = computed(() => {
  const mode = props.storageClass?.volumeBindingMode
  if (mode === 'WaitForFirstConsumer') return '首次使用时绑定'
  if (mode === 'Immediate') return '立即绑定'
  return mode || '-'
})

const initForm = () => {
  const sc = props.storageClass
  form.alias = sc?.alias || ''
  form.describe = sc?.describe || ''
  form.storageType = sc && !sc.storageTypeAuto ? sc.storageTypeKey || '' : ''
  form.allowVolumeExpansion = !!sc?.allowVolumeExpansion
  form.allowVolumeClone = !!sc?.allowVolumeClone
  form.isDefault = !!sc?.isDefault
}

const submit = async () => {
  if (!props.storageClass) return
  submitting.value = true
  try {
    await storageClassApi.update({
      clusterId: props.clusterId,
      name: props.storageClass.name,
      alias: form.alias.trim(),
      describe: form.describe.trim(),
      storageType: form.storageType,
      allowVolumeExpansion: form.allowVolumeExpansion,
      allowVolumeClone: form.allowVolumeClone,
      isDefault: form.isDefault
    })
    ElMessage.success('更新成功')
    emit('update:modelValue', false)
    emit('success')
  } catch {
    // 请求拦截器已提示错误
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.sc-fixed {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin: 0 0 20px 24px;
  padding: 12px 14px;
  border: 1px solid #e8eef8;
  border-radius: 12px;
  background: #f8fbff;
}

.sc-fixed__item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.sc-fixed__label {
  font-size: 12px;
  color: #94a3b8;
}

.sc-fixed__value {
  overflow: hidden;
  color: #1e293b;
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

code.sc-fixed__value {
  font-family: 'SF Mono', Consolas, Monaco, monospace;
  font-weight: 500;
}

.sc-fixed__tip {
  grid-column: 1 / -1;
  margin: 2px 0 0;
  font-size: 12px;
  color: #94a3b8;
}

.sc-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.sc-option__desc {
  font-size: 12px;
  color: #94a3b8;
}

.sc-inline-hint {
  margin-left: 12px;
  font-size: 12px;
  color: #94a3b8;
}

@media (max-width: 768px) {
  .sc-fixed {
    grid-template-columns: 1fr;
    margin-left: 0;
  }
}
</style>
