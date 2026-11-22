<template>
  <el-dialog
    :model-value="modelValue"
    title="创建存储类"
    width="1080px"
    top="6vh"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--wide"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="reset"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">创建存储类</h3>
        <div class="dialog-actions">
          <el-switch v-model="yamlMode" active-text="YAML 预览" />
        </div>
      </div>
    </template>

    <div class="config-editor">
      <div v-if="!yamlMode" class="step-tabs">
        <div
          v-for="(step, index) in steps"
          :key="step.title"
          :class="['step-tab', { active: current === index, completed: index < current }]"
          @click="goStep(index)"
        >
          <div class="step-icon">
            <el-icon v-if="index < current"><Check /></el-icon>
            <span v-else>{{ index + 1 }}</span>
          </div>
          <div class="step-info">
            <div class="step-title">{{ step.title }}</div>
            <div class="step-desc">{{ step.desc }}</div>
          </div>
        </div>
      </div>

      <div class="config-content">
        <!-- YAML 预览 -->
        <div v-if="yamlMode" class="yaml-view">
          <YamlEditor :model-value="previewYaml" :readonly="true" height="100%" title="存储类 YAML 预览" :filename="`${form.name || 'storageclass'}.yaml`" />
        </div>

        <!-- 第 1 步: 存储系统 -->
        <div v-else-if="current === 0" class="sc-systems">
          <p class="sc-tip">选择存储系统后，会自动填好供应者和常用参数，后面的步骤里仍然可以修改。</p>
          <div class="sc-system-grid">
            <button
              v-for="tpl in PROVISIONER_TEMPLATES"
              :key="tpl.key"
              type="button"
              :class="['sc-system', { 'is-active': form.template === tpl.key }]"
              @click="selectTemplate(tpl)"
            >
              <span class="sc-system__logo" :style="{ background: tpl.color }">{{ tpl.short }}</span>
              <span class="sc-system__body">
                <span class="sc-system__head">
                  <span class="sc-system__name">{{ tpl.name }}</span>
                  <span v-if="tpl.storageType" class="sc-system__type">{{ storageTypeLabel(tpl.storageType) }}</span>
                </span>
                <span class="sc-system__desc">{{ tpl.desc }}</span>
                <code class="sc-system__provisioner">{{ tpl.provisioner || '自行填写供应者' }}</code>
              </span>
              <el-icon class="sc-system__check"><CircleCheckFilled /></el-icon>
            </button>
          </div>
        </div>

        <!-- 第 2 步: 基本信息 -->
        <div v-else-if="current === 1" class="basic-form">
          <el-form ref="basicRef" :model="form" :rules="rules" label-width="100px" label-position="right">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="例如 nfs-client" maxlength="253" show-word-limit />
              <div class="field-hint">小写字母、数字、"-" 和 "."，创建后不可修改</div>
            </el-form-item>
            <el-form-item label="别名" prop="alias">
              <el-input v-model="form.alias" placeholder="可选，用于列表展示" maxlength="50" show-word-limit />
            </el-form-item>
            <el-form-item label="存储类型" prop="storageType">
              <el-select v-model="form.storageType" placeholder="自动识别" clearable style="width: 100%">
                <el-option v-for="opt in STORAGE_TYPE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value">
                  <div class="sc-option">
                    <span>{{ opt.label }}</span>
                    <span class="sc-option__desc">{{ opt.desc }}</span>
                  </div>
                </el-option>
              </el-select>
              <div class="field-hint">不选时根据供应者自动识别</div>
            </el-form-item>
            <el-form-item label="描述" prop="describe">
              <el-input v-model="form.describe" type="textarea" :rows="3" placeholder="可选，说明用途" maxlength="200" show-word-limit />
            </el-form-item>
          </el-form>
        </div>

        <!-- 第 3 步: 存储类设置 -->
        <div v-else class="basic-form">
          <el-form ref="settingRef" :model="form" :rules="rules" label-width="100px" label-position="right">
            <el-form-item label="供应者" prop="provisioner" class="is-full">
              <el-input v-model="form.provisioner" :disabled="form.template !== 'custom'" placeholder="例如 diskplugin.csi.alibabacloud.com" />
              <div class="field-hint">{{ form.template === 'custom' ? '填写集群中已安装的 CSI 驱动名称' : '由所选存储系统决定，如需修改请选择「自定义」' }}</div>
            </el-form-item>
            <el-form-item label="回收策略">
              <el-radio-group v-model="form.reclaimPolicy">
                <el-radio-button value="Delete">删除</el-radio-button>
                <el-radio-button value="Retain">保留</el-radio-button>
              </el-radio-group>
              <div class="field-hint">删除 PVC 后，{{ form.reclaimPolicy === 'Delete' ? '同时删除底层存储卷' : '保留底层存储卷和数据' }}</div>
            </el-form-item>
            <el-form-item label="绑定模式">
              <el-radio-group v-model="form.volumeBindingMode">
                <el-radio-button value="Immediate">立即绑定</el-radio-button>
                <el-radio-button value="WaitForFirstConsumer">首次使用时绑定</el-radio-button>
              </el-radio-group>
              <div class="field-hint">{{ form.volumeBindingMode === 'Immediate' ? '创建 PVC 后立即创建并绑定存储卷' : '等 Pod 调度到节点后再创建存储卷，适合本地盘和多可用区' }}</div>
            </el-form-item>
            <el-form-item label="卷扩展">
              <el-switch v-model="form.allowVolumeExpansion" active-text="允许" inactive-text="禁止" inline-prompt />
              <span class="sc-inline-hint">允许 PVC 在使用中扩容</span>
            </el-form-item>
            <el-form-item label="卷克隆">
              <el-switch v-model="form.allowVolumeClone" active-text="允许" inactive-text="禁止" inline-prompt />
              <span class="sc-inline-hint">允许以已有 PVC 为数据源创建新 PVC</span>
            </el-form-item>
            <el-form-item label="访问模式" class="is-full">
              <el-checkbox-group v-model="form.accessModes" class="sc-access-modes">
                <el-checkbox v-for="mode in ACCESS_MODE_OPTIONS" :key="mode.value" :value="mode.value" border>
                  <span class="sc-access-mode">{{ mode.label }}</span>
                  <span class="sc-access-mode__desc">{{ mode.desc }}</span>
                </el-checkbox>
              </el-checkbox-group>
              <div class="field-hint">限制使用该存储类的 PVC 可以选择的访问模式，不选表示不限制</div>
            </el-form-item>
            <el-form-item label="挂载选项" class="is-full">
              <el-select v-model="form.mountOptions" multiple filterable allow-create default-first-option :reserve-keyword="false" placeholder="输入后回车添加，例如 nfsvers=4.1" style="width: 100%" />
            </el-form-item>
            <el-form-item label="默认存储类">
              <el-switch v-model="form.isDefault" />
              <span class="sc-inline-hint">未指定存储类的 PVC 会使用它，同时会取消其他存储类的默认标记</span>
            </el-form-item>
          </el-form>

          <div class="form-section">
            <div class="section-header-row">
              <div class="section-title">参数</div>
              <el-button size="small" @click="form.params.push({ key: '', value: '' })">添加参数</el-button>
            </div>
            <el-table :data="form.params" style="width: 100%">
              <el-table-column label="键" width="380">
                <template #default="{ row }">
                  <el-input v-model="row.key" placeholder="key" />
                </template>
              </el-table-column>
              <el-table-column label="值">
                <template #default="{ row }">
                  <el-input v-model="row.value" placeholder="value" />
                </template>
              </el-table-column>
              <el-table-column label="操作" width="90" align="center">
                <template #default="{ $index }">
                  <el-button link type="danger" @click="form.params.splice($index, 1)">删除</el-button>
                </template>
              </el-table-column>
              <template #empty>
                <span>暂无参数，供应者不需要参数时可以留空</span>
              </template>
            </el-table>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <template v-if="!yamlMode">
          <el-button v-if="current > 0" @click="current--">上一步</el-button>
          <el-button v-if="current < steps.length - 1" type="primary" @click="next">下一步</el-button>
          <el-button v-else type="primary" :loading="submitting" @click="submit">创建</el-button>
        </template>
        <el-button v-else type="primary" :loading="submitting" @click="submit">创建</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Check, CircleCheckFilled } from '@element-plus/icons-vue'
import YamlEditor from '@/components/YamlEditor.vue'
import { storageClassApi, type StorageTypeKey } from '@/api/storageclass'
import {
  ACCESS_MODE_OPTIONS,
  PROVISIONER_TEMPLATES,
  SC_ANNOTATIONS,
  STORAGE_TYPE_OPTIONS,
  storageTypeLabel,
  type ParamRow,
  type ProvisionerTemplate
} from '@/utils/storageclass'
import { jsonToYaml } from '@/utils/yaml'

const props = defineProps<{
  modelValue: boolean
  clusterId: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: []
}>()

const steps = [
  { title: '存储系统', desc: '选择存储后端' },
  { title: '基本信息', desc: '名称、别名与存储类型' },
  { title: '存储类设置', desc: '回收、绑定、扩展与参数' }
]

const current = ref(0)
const yamlMode = ref(false)
const submitting = ref(false)
const basicRef = ref<FormInstance>()
const settingRef = ref<FormInstance>()

const form = reactive({
  template: '',
  name: '',
  alias: '',
  describe: '',
  storageType: '' as StorageTypeKey,
  provisioner: '',
  reclaimPolicy: 'Delete' as 'Delete' | 'Retain',
  volumeBindingMode: 'Immediate' as 'Immediate' | 'WaitForFirstConsumer',
  allowVolumeExpansion: false,
  allowVolumeClone: false,
  accessModes: [] as string[],
  mountOptions: [] as string[],
  isDefault: false,
  params: [] as ParamRow[]
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    {
      pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/,
      message: '只能包含小写字母、数字、"-" 和 "."，且以字母或数字开头和结尾',
      trigger: 'blur'
    }
  ],
  provisioner: [{ required: true, message: '请输入供应者', trigger: 'blur' }]
}

const selectTemplate = (tpl: ProvisionerTemplate) => {
  form.template = tpl.key
  form.provisioner = tpl.provisioner
  form.storageType = tpl.storageType
  form.reclaimPolicy = tpl.reclaimPolicy
  form.volumeBindingMode = tpl.volumeBindingMode
  form.allowVolumeExpansion = tpl.allowVolumeExpansion
  form.allowVolumeClone = tpl.allowVolumeClone
  form.accessModes = [...tpl.accessModes]
  form.mountOptions = [...tpl.mountOptions]
  form.params = tpl.params.map(p => ({ ...p }))
  if (!form.name && tpl.key !== 'custom') {
    form.name = tpl.key
  }
}

const reset = () => {
  current.value = 0
  yamlMode.value = false
  form.template = ''
  form.name = ''
  form.alias = ''
  form.describe = ''
  form.isDefault = false
  selectTemplate(PROVISIONER_TEMPLATES[0]!)
}

// 校验当前步骤, 通过后返回 true
const validateCurrent = async () => {
  if (current.value === 0) {
    if (!form.template) {
      ElMessage.warning('请选择存储系统')
      return false
    }
    return true
  }
  const formEl = current.value === 1 ? basicRef.value : settingRef.value
  if (!formEl) return true
  try {
    await formEl.validate()
    return true
  } catch {
    return false
  }
}

const next = async () => {
  if (await validateCurrent()) current.value++
}

// 点击步骤条: 后退随时可以, 前进需要先通过当前步骤的校验
const goStep = async (index: number) => {
  if (index <= current.value) {
    current.value = index
    return
  }
  while (current.value < index) {
    if (!(await validateCurrent())) return
    current.value++
  }
}

const collectParams = () => {
  const params: Record<string, string> = {}
  for (const row of form.params) {
    const key = row.key.trim()
    if (key) params[key] = row.value
  }
  return params
}

const buildAnnotations = () => {
  const annotations: Record<string, string> = {}
  if (form.alias.trim()) annotations.alias = form.alias.trim()
  if (form.describe.trim()) annotations.describe = form.describe.trim()
  if (form.storageType) annotations[SC_ANNOTATIONS.storageType] = form.storageType
  annotations[SC_ANNOTATIONS.allowClone] = String(form.allowVolumeClone)
  if (form.accessModes.length) annotations[SC_ANNOTATIONS.accessModes] = JSON.stringify(form.accessModes)
  if (form.isDefault) annotations[SC_ANNOTATIONS.isDefault] = 'true'
  return annotations
}

// 预览与后端 buildStorageClass 生成的对象一致
const previewYaml = computed(() => {
  const manifest: Record<string, unknown> = {
    apiVersion: 'storage.k8s.io/v1',
    kind: 'StorageClass',
    metadata: { name: form.name || '<名称>', annotations: buildAnnotations() },
    provisioner: form.provisioner || '<供应者>'
  }
  const params = collectParams()
  if (Object.keys(params).length) manifest.parameters = params
  manifest.reclaimPolicy = form.reclaimPolicy
  manifest.volumeBindingMode = form.volumeBindingMode
  manifest.allowVolumeExpansion = form.allowVolumeExpansion
  if (form.mountOptions.length) manifest.mountOptions = form.mountOptions
  return jsonToYaml(manifest) + '\n'
})

const submit = async () => {
  if (!form.name.trim()) {
    yamlMode.value = false
    current.value = 1
    ElMessage.warning('请填写名称')
    return
  }
  if (!form.provisioner.trim()) {
    yamlMode.value = false
    current.value = 2
    ElMessage.warning('请填写供应者')
    return
  }
  if (!yamlMode.value && !(await validateCurrent())) return
  const emptyParam = form.params.find(p => p.key.trim() && p.value === '')
  if (emptyParam) {
    yamlMode.value = false
    current.value = 2
    ElMessage.warning(`参数 ${emptyParam.key} 的值不能为空，不需要的参数请删除`)
    return
  }

  submitting.value = true
  try {
    await storageClassApi.create({
      clusterId: props.clusterId,
      name: form.name.trim(),
      alias: form.alias.trim(),
      describe: form.describe.trim(),
      storageType: form.storageType,
      provisioner: form.provisioner.trim(),
      parameters: collectParams(),
      reclaimPolicy: form.reclaimPolicy,
      volumeBindingMode: form.volumeBindingMode,
      allowVolumeExpansion: form.allowVolumeExpansion,
      allowVolumeClone: form.allowVolumeClone,
      accessModes: form.accessModes,
      mountOptions: form.mountOptions,
      isDefault: form.isDefault
    })
    ElMessage.success('创建成功')
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
.sc-tip {
  margin: 0 0 16px;
  font-size: 13px;
  color: #64748b;
}

.sc-system-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 14px;
}

.sc-system {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 14px;
  min-width: 0;
  padding: 16px;
  border: 1px solid #e3eaf5;
  border-radius: 14px;
  background: #ffffff;
  font: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.15s ease, box-shadow 0.15s ease, background 0.15s ease;
}

.sc-system:hover {
  border-color: #b9cdf3;
  box-shadow: 0 6px 18px rgba(37, 99, 235, 0.08);
}

.sc-system:focus-visible {
  outline: 2px solid #3b74f2;
  outline-offset: 2px;
}

.sc-system.is-active {
  border-color: #3b74f2;
  background: #f5f9ff;
  box-shadow: 0 0 0 3px rgba(59, 116, 242, 0.12);
}

.sc-system__logo {
  display: grid;
  place-items: center;
  flex: none;
  width: 46px;
  height: 46px;
  border-radius: 12px;
  color: #ffffff;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.5px;
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.12);
}

.sc-system__body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.sc-system__head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sc-system__name {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
}

.sc-system__type {
  padding: 1px 8px;
  border-radius: 999px;
  background: #eff5ff;
  color: #2563eb;
  font-size: 12px;
}

.sc-system__desc {
  font-size: 13px;
  line-height: 1.5;
  color: #64748b;
}

.sc-system__provisioner {
  overflow: hidden;
  margin-top: 2px;
  color: #94a3b8;
  font-family: 'SF Mono', Consolas, Monaco, monospace;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sc-system__check {
  position: absolute;
  top: 12px;
  right: 12px;
  font-size: 18px;
  color: #3b74f2;
  opacity: 0;
  transform: scale(0.6);
  transition: all 0.15s ease;
}

.sc-system.is-active .sc-system__check {
  opacity: 1;
  transform: scale(1);
}

.basic-form :deep(.el-form-item.is-full) {
  grid-column: 1 / -1;
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

.sc-access-modes {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.sc-access-modes :deep(.el-checkbox) {
  height: 36px;
  margin-right: 0;
  border-radius: 8px;
}

.sc-access-mode {
  font-family: 'SF Mono', Consolas, Monaco, monospace;
  font-weight: 600;
}

.sc-access-mode__desc {
  margin-left: 6px;
  font-size: 12px;
  color: #94a3b8;
}

@media (max-width: 900px) {
  .sc-inline-hint {
    display: block;
    margin-left: 0;
  }
}
</style>
