<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="1080px"
    top="6vh"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--wide"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="reset"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">{{ title }}</h3>
        <div class="dialog-actions">
          <el-switch v-model="yamlMode" active-text="YAML 预览" />
        </div>
      </div>
    </template>

    <div class="config-editor">
      <div v-if="!yamlMode" class="step-tabs">
        <div
          v-for="(step, index) in steps"
          :key="step.key"
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
        <div v-if="yamlMode" class="yaml-view">
          <YamlEditor
            :model-value="previewYaml"
            :readonly="true"
            height="100%"
            :title="`${kindLabel} YAML 预览`"
            :filename="`${form.name || kind}.yaml`"
          />
        </div>

        <!-- 基本信息 -->
        <div v-else-if="stepKey === 'basic'" class="basic-form">
          <el-form ref="basicRef" :model="form" :rules="rules" label-width="100px" label-position="right">
            <el-form-item label="命名空间" prop="namespace">
              <InfiniteSelect v-model="form.namespace" :fetch-data="namespaceFetcher" v-bind="namespaceSelectConfig" :clearable="false" placeholder="请选择命名空间" style="width: 100%" />
            </el-form-item>
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" :placeholder="kind === 'job' ? '例如 db-backup' : '例如 nightly-report'" :maxlength="nameMax" show-word-limit />
              <div class="field-hint">小写字母、数字和 "-"，以字母或数字开头和结尾，创建后不可修改</div>
            </el-form-item>
            <el-form-item label="别名">
              <el-input v-model="form.alias" placeholder="可选，用于列表展示" maxlength="50" show-word-limit />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="form.describe" type="textarea" :rows="3" placeholder="可选，说明用途" maxlength="200" show-word-limit />
            </el-form-item>
          </el-form>
        </div>

        <!-- 定时设置 -->
        <div v-else-if="stepKey === 'schedule'" class="basic-form">
          <el-form :model="cron" label-width="120px" label-position="right">
            <el-form-item label="调度规则" required>
              <div class="jc-presets">
                <button
                  v-for="p in SCHEDULE_PRESETS"
                  :key="p.value"
                  type="button"
                  :class="['jc-preset', { 'is-active': cron.schedule.trim() === p.value }]"
                  @click="cron.schedule = p.value"
                >
                  {{ p.label }}
                </button>
              </div>
              <el-input v-model="cron.schedule" placeholder="分 时 日 月 周，例如 0 2 * * *" class="mono-input" />
              <div class="field-hint">
                <template v-if="scheduleDesc">执行时间：{{ scheduleDesc }}</template>
                <template v-else>Cron 表达式：分(0-59) 时(0-23) 日(1-31) 月(1-12) 周(0-6，0 为周日)</template>
              </div>
            </el-form-item>
            <el-form-item label="时区">
              <el-select v-model="cron.timeZone" filterable allow-create clearable placeholder="不设置时使用集群控制器所在时区" style="width: 100%">
                <el-option v-for="tz in TIMEZONE_OPTIONS" :key="tz" :label="tz" :value="tz" />
              </el-select>
            </el-form-item>
            <el-form-item label="并发策略" class="is-full">
              <div class="fd-choice-grid jc-grid-3">
                <button
                  v-for="p in CONCURRENCY_POLICIES"
                  :key="p.value"
                  type="button"
                  :class="['fd-choice', { 'is-active': cron.concurrencyPolicy === p.value }]"
                  @click="cron.concurrencyPolicy = p.value"
                >
                  <span class="fd-choice__body">
                    <span class="fd-choice__title"><span>{{ p.label }}</span><span class="fd-choice__badge">{{ p.value }}</span></span>
                    <span class="fd-choice__desc">{{ p.desc }}</span>
                  </span>
                  <el-icon class="fd-choice__check"><CircleCheckFilled /></el-icon>
                </button>
              </div>
            </el-form-item>
            <el-form-item label="启动截止时间">
              <el-input-number v-model="cron.startingDeadlineSeconds" :min="0" :max="86400" controls-position="right" />
              <span class="fd-inline-hint">秒。错过调度时间后，超过这个时间就不再启动；0 表示不限制</span>
            </el-form-item>
            <el-form-item label="保留成功任务">
              <el-input-number v-model="cron.successfulJobsHistoryLimit" :min="0" :max="100" controls-position="right" />
              <span class="fd-inline-hint">个，更早的成功任务会被自动删除</span>
            </el-form-item>
            <el-form-item label="保留失败任务">
              <el-input-number v-model="cron.failedJobsHistoryLimit" :min="0" :max="100" controls-position="right" />
              <span class="fd-inline-hint">个</span>
            </el-form-item>
            <el-form-item label="创建后暂停">
              <el-switch v-model="cron.suspend" />
              <span class="fd-inline-hint">开启后不会按计划执行，可以稍后在列表或详情中启用</span>
            </el-form-item>
          </el-form>
        </div>

        <!-- 容器设置 -->
        <div v-else-if="stepKey === 'containers'" class="basic-form">
          <div v-for="(c, index) in containers" :key="index" class="jc-container">
            <div class="jc-container__head">
              <span class="jc-container__title">容器 {{ index + 1 }}<small v-if="c.name">{{ c.name }}</small></span>
              <el-button v-if="containers.length > 1" link type="danger" @click="containers.splice(index, 1)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
            <el-form :model="c" label-width="100px" label-position="right">
              <div class="jc-row">
                <el-form-item label="容器名称" required>
                  <el-input v-model="c.name" placeholder="例如 main" maxlength="63" />
                </el-form-item>
                <el-form-item label="拉取策略">
                  <el-select v-model="c.imagePullPolicy" style="width: 100%">
                    <el-option label="本地没有时拉取（IfNotPresent）" value="IfNotPresent" />
                    <el-option label="每次都拉取（Always）" value="Always" />
                    <el-option label="只使用本地镜像（Never）" value="Never" />
                  </el-select>
                </el-form-item>
              </div>
              <el-form-item label="镜像" required>
                <el-input v-model="c.image" placeholder="例如 busybox:1.36、registry.example.com/team/app:v1" class="mono-input" />
              </el-form-item>
              <el-form-item label="启动命令">
                <el-radio-group v-model="c.commandMode">
                  <el-radio-button value="shell">Shell 脚本</el-radio-button>
                  <el-radio-button value="exec">命令与参数</el-radio-button>
                  <el-radio-button value="default">使用镜像默认</el-radio-button>
                </el-radio-group>
              </el-form-item>
              <el-form-item v-if="c.commandMode === 'shell'" label="脚本">
                <el-input v-model="c.script" type="textarea" :rows="4" placeholder="例如：echo start && sleep 10 && echo done" class="mono-input" />
                <div class="field-hint">以 /bin/sh -c 执行，镜像中需要有 sh；留空则使用镜像默认命令</div>
              </el-form-item>
              <template v-else-if="c.commandMode === 'exec'">
                <el-form-item label="命令">
                  <el-input v-model="c.commandText" type="textarea" :rows="2" placeholder="每行一项，例如&#10;python&#10;main.py" class="mono-input" />
                </el-form-item>
                <el-form-item label="参数">
                  <el-input v-model="c.argsText" type="textarea" :rows="2" placeholder="每行一项，例如&#10;--date&#10;today" class="mono-input" />
                  <div class="field-hint">每行是一个独立参数，不需要加引号</div>
                </el-form-item>
              </template>
              <el-form-item label="工作目录">
                <el-input v-model="c.workingDir" placeholder="可选，例如 /app" class="mono-input" />
              </el-form-item>
              <div class="jc-row jc-row--4">
                <el-form-item label="CPU 预留">
                  <el-input v-model="c.cpuRequest" placeholder="如 100m" />
                </el-form-item>
                <el-form-item label="CPU 上限">
                  <el-input v-model="c.cpuLimit" placeholder="如 1" />
                </el-form-item>
                <el-form-item label="内存预留">
                  <el-input v-model="c.memoryRequest" placeholder="如 128Mi" />
                </el-form-item>
                <el-form-item label="内存上限">
                  <el-input v-model="c.memoryLimit" placeholder="如 512Mi" />
                </el-form-item>
              </div>
            </el-form>
            <KVRowsTable v-model="c.env" title="环境变量" add-text="添加环境变量" empty-text="暂无环境变量" key-placeholder="变量名" value-placeholder="值" />
          </div>
          <el-button class="jc-add" @click="addContainer">
            <el-icon><Plus /></el-icon>
            添加容器
          </el-button>
        </div>

        <!-- 任务设置 -->
        <div v-else class="basic-form">
          <el-form :model="settings" label-width="130px" label-position="right">
            <el-form-item label="完成数量">
              <el-input-number v-model="settings.completions" :min="1" :max="10000" controls-position="right" />
              <span class="fd-inline-hint">需要成功运行完成的容器组数量</span>
            </el-form-item>
            <el-form-item label="并行数量">
              <el-input-number v-model="settings.parallelism" :min="1" :max="10000" controls-position="right" />
              <span class="fd-inline-hint">同时运行的容器组数量</span>
            </el-form-item>
            <el-form-item label="最大重试次数">
              <el-input-number v-model="settings.backoffLimit" :min="0" :max="100" controls-position="right" />
              <span class="fd-inline-hint">失败超过这个次数后任务标记为失败</span>
            </el-form-item>
            <el-form-item label="最大运行时间">
              <el-input-number v-model="settings.activeDeadlineSeconds" :min="0" :max="31536000" controls-position="right" />
              <span class="fd-inline-hint">秒，超时后停止任务并标记为失败；0 表示不限制</span>
            </el-form-item>
            <el-form-item label="重启策略">
              <el-radio-group v-model="settings.restartPolicy">
                <el-radio-button value="Never">不重启（Never）</el-radio-button>
                <el-radio-button value="OnFailure">失败时重启（OnFailure）</el-radio-button>
              </el-radio-group>
              <div class="field-hint">
                {{ settings.restartPolicy === 'Never' ? '容器失败时创建新的容器组重试，失败的容器组会保留，方便查看日志' : '容器失败时在原容器组内重启容器' }}
              </div>
            </el-form-item>
            <el-form-item label="结束后自动删除">
              <el-switch v-model="settings.autoClean" />
              <template v-if="settings.autoClean">
                <el-input-number v-model="settings.ttlSecondsAfterFinished" :min="0" :max="31536000" controls-position="right" style="margin-left: 12px" />
                <span class="fd-inline-hint">秒后删除任务及其容器组</span>
              </template>
            </el-form-item>
          </el-form>
          <KVRowsTable v-model="labels" title="标签" add-text="添加标签" empty-text="暂无标签" />
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
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { computed, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Check, CircleCheckFilled, Delete, Plus } from '@element-plus/icons-vue'
import YamlEditor from '@/components/YamlEditor.vue'
import KVRowsTable from '@/components/KVRowsTable.vue'
import { jobApi } from '@/api/job'
import { cronJobApi, type ConcurrencyPolicy } from '@/api/cronjob'
import { jsonToYaml } from '@/utils/yaml'
import {
  CONCURRENCY_POLICIES,
  DNS_LABEL,
  SCHEDULE_PRESETS,
  TIMEZONE_OPTIONS,
  buildJobTemplateDTO,
  createContainerForm,
  createJobSettingsForm,
  describeCron,
  jobSpecPreview,
  type ContainerForm,
  type KVRow
} from '@/utils/job'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    clusterId: string
    kind: 'job' | 'cronjob'
    defaultNamespace?: string
  }>(),
  { defaultNamespace: '' }
)
// 命名空间下拉框分页加载
const namespaceFetcher = computed(() => useNamespaceFetcher(props.clusterId))

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  // 创建成功, 参数为命名空间和名称
  success: [namespace: string, name: string]
}>()

type StepKey = 'basic' | 'schedule' | 'containers' | 'job'

const kindLabel = computed(() => (props.kind === 'job' ? 'Job' : 'CronJob'))
const title = computed(() => (props.kind === 'job' ? '创建任务（Job）' : '创建定时任务（CronJob）'))
// 定时任务名称最长 52: 控制器创建任务时会在名称后追加 11 位
const nameMax = computed(() => (props.kind === 'job' ? 63 : 52))

const steps = computed<{ key: StepKey; title: string; desc: string }[]>(() => [
  { key: 'basic', title: '基本信息', desc: '命名空间、名称与描述' },
  ...(props.kind === 'cronjob' ? [{ key: 'schedule' as StepKey, title: '定时设置', desc: '调度规则、时区与并发策略' }] : []),
  { key: 'containers', title: '容器设置', desc: '镜像、命令、环境变量与资源' },
  { key: 'job', title: '任务设置', desc: '完成数、并行数、重试与清理' }
])

const current = ref(0)
const stepKey = computed(() => steps.value[current.value]?.key)
const yamlMode = ref(false)
const submitting = ref(false)
const basicRef = ref<FormInstance>()


const form = reactive({ namespace: '', name: '', alias: '', describe: '' })

const createCronForm = () => ({
  schedule: '0 2 * * *',
  timeZone: 'Asia/Shanghai',
  concurrencyPolicy: 'Forbid' as ConcurrencyPolicy,
  startingDeadlineSeconds: 0,
  successfulJobsHistoryLimit: 3,
  failedJobsHistoryLimit: 1,
  suspend: false
})
const cron = reactive(createCronForm())
const containers = ref<ContainerForm[]>([createContainerForm()])
const settings = reactive(createJobSettingsForm())
const labels = ref<KVRow[]>([])

const scheduleDesc = computed(() => describeCron(cron.schedule))

const rules = computed<FormRules>(() => ({
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    {
      validator: (_rule, value: string, callback) => {
        const v = (value || '').trim()
        if (!DNS_LABEL.test(v)) callback(new Error('只能包含小写字母、数字和 "-"，且以字母或数字开头和结尾'))
        else if (v.length > nameMax.value) callback(new Error(`名称不能超过 ${nameMax.value} 个字符`))
        else callback()
      },
      trigger: 'blur'
    }
  ]
}))

const reset = () => {
  Object.assign(form, { namespace: props.defaultNamespace || '', name: '', alias: '', describe: '' })
  Object.assign(cron, createCronForm())
  Object.assign(settings, createJobSettingsForm())
  // 定时任务通常是脚本, 失败时在原容器组重启即可
  if (props.kind === 'cronjob') {
    settings.restartPolicy = 'OnFailure'
    settings.backoffLimit = 2
  }
  containers.value = [createContainerForm()]
  labels.value = []
  current.value = 0
  yamlMode.value = false
}

const addContainer = () => containers.value.push(createContainerForm(containers.value.length))

const QUANTITY = /^\d+(\.\d+)?(m|k|M|G|T|P|Ki|Mi|Gi|Ti|Pi)?$/

const validateContainers = () => {
  const names = new Set<string>()
  for (let i = 0; i < containers.value.length; i++) {
    const c = containers.value[i]
    const label = `容器 ${i + 1}`
    const name = c.name.trim()
    if (!DNS_LABEL.test(name)) return `${label}的名称不合法：只能包含小写字母、数字和 "-"`
    if (names.has(name)) return `容器名称 ${name} 重复`
    names.add(name)
    if (!c.image.trim()) return `请填写${label}的镜像`
    if (/\s/.test(c.image.trim())) return `${label}的镜像不能包含空格`
    if (c.commandMode === 'exec' && !c.commandText.trim() && !c.argsText.trim()) return `请填写${label}的命令或参数，或选择「使用镜像默认」`
    const resources: [string, string][] = [
      [c.cpuRequest, 'CPU 预留'],
      [c.cpuLimit, 'CPU 上限'],
      [c.memoryRequest, '内存预留'],
      [c.memoryLimit, '内存上限']
    ]
    for (const [value, text] of resources) {
      if (value.trim() && !QUANTITY.test(value.trim())) return `${label}的${text}格式不正确，例如 100m、0.5、128Mi`
    }
    const badEnv = c.env.find(e => e.key.trim() && !/^[-._a-zA-Z][-._a-zA-Z0-9]*$/.test(e.key.trim()))
    if (badEnv) return `${label}的环境变量名 ${badEnv.key} 不合法`
  }
  return ''
}

const validateSchedule = () => {
  const v = cron.schedule.trim()
  if (!v) return '请填写调度规则'
  if (v.startsWith('@')) return ''
  if (v.split(/\s+/).length !== 5) return '调度规则需要 5 个字段：分 时 日 月 周'
  return ''
}

const validateStep = async (index: number) => {
  const key = steps.value[index]?.key
  if (key === 'basic') {
    if (!basicRef.value || stepKey.value !== 'basic') {
      const ok = !!form.namespace && DNS_LABEL.test(form.name.trim()) && form.name.trim().length <= nameMax.value
      if (!ok) ElMessage.warning('请选择命名空间并填写正确的名称')
      return ok
    }
    try {
      await basicRef.value.validate()
      return true
    } catch {
      return false
    }
  }
  const message = key === 'schedule' ? validateSchedule() : key === 'containers' ? validateContainers() : ''
  if (message) {
    ElMessage.warning(message)
    return false
  }
  return true
}

const next = async () => {
  if (await validateStep(current.value)) current.value++
}

const goStep = async (index: number) => {
  if (index <= current.value) {
    current.value = index
    return
  }
  while (current.value < index) {
    if (!(await validateStep(current.value))) return
    current.value++
  }
}

const labelMap = computed(() => {
  const result: Record<string, string> = {}
  labels.value.forEach(r => {
    if (r.key.trim()) result[r.key.trim()] = r.value
  })
  return result
})

const template = computed(() => buildJobTemplateDTO(settings, containers.value))

const previewYaml = computed(() => {
  const annotations: Record<string, string> = {}
  if (form.alias.trim()) annotations.alias = form.alias.trim()
  if (form.describe.trim()) annotations.describe = form.describe.trim()
  const metadata: Record<string, unknown> = { name: form.name.trim() || `example-${props.kind}`, namespace: form.namespace || 'default' }
  if (Object.keys(labelMap.value).length) metadata.labels = labelMap.value
  if (Object.keys(annotations).length) metadata.annotations = annotations
  const jobSpec = jobSpecPreview(template.value, labelMap.value)
  let obj: Record<string, unknown>
  if (props.kind === 'job') {
    obj = { apiVersion: 'batch/v1', kind: 'Job', metadata, spec: jobSpec }
  } else {
    const spec: Record<string, unknown> = { schedule: cron.schedule.trim() }
    if (cron.timeZone) spec.timeZone = cron.timeZone
    spec.concurrencyPolicy = cron.concurrencyPolicy
    if (cron.startingDeadlineSeconds > 0) spec.startingDeadlineSeconds = cron.startingDeadlineSeconds
    spec.successfulJobsHistoryLimit = cron.successfulJobsHistoryLimit
    spec.failedJobsHistoryLimit = cron.failedJobsHistoryLimit
    spec.suspend = cron.suspend
    spec.jobTemplate = Object.keys(labelMap.value).length ? { metadata: { labels: labelMap.value }, spec: jobSpec } : { spec: jobSpec }
    obj = { apiVersion: 'batch/v1', kind: 'CronJob', metadata, spec }
  }
  // 去掉 undefined
  return jsonToYaml(JSON.parse(JSON.stringify(obj)))
})

const submit = async () => {
  for (let step = 0; step < steps.value.length; step++) {
    if (!(await validateStep(step))) {
      yamlMode.value = false
      current.value = step
      return
    }
  }

  submitting.value = true
  const base = {
    clusterId: props.clusterId,
    namespace: form.namespace,
    name: form.name.trim(),
    alias: form.alias.trim(),
    describe: form.describe.trim(),
    labels: labelMap.value,
    template: template.value
  }
  try {
    if (props.kind === 'job') {
      await jobApi.create(base)
    } else {
      await cronJobApi.create({
        ...base,
        schedule: cron.schedule.trim(),
        timeZone: cron.timeZone || '',
        concurrencyPolicy: cron.concurrencyPolicy,
        startingDeadlineSeconds: cron.startingDeadlineSeconds > 0 ? cron.startingDeadlineSeconds : undefined,
        successfulJobsHistoryLimit: cron.successfulJobsHistoryLimit,
        failedJobsHistoryLimit: cron.failedJobsHistoryLimit,
        suspend: cron.suspend
      })
    }
    ElMessage.success('创建成功')
    emit('update:modelValue', false)
    emit('success', base.namespace, base.name)
  } catch {
    // 请求拦截器已提示错误
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.jc-presets {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  width: 100%;
  margin-bottom: 10px;
}

.jc-preset {
  height: 30px;
  padding: 0 14px;
  border: 1px solid #d8e1ee;
  border-radius: 15px;
  background: #fff;
  color: #475569;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.jc-preset:hover {
  border-color: #3b74f2;
  color: #3b74f2;
}

.jc-preset.is-active {
  border-color: #3b74f2;
  background: #eef4ff;
  color: #3b74f2;
  font-weight: 600;
}

.jc-grid-3 {
  grid-template-columns: repeat(3, minmax(0, 1fr));
  width: 100%;
  margin-bottom: 0;
}

.jc-container {
  margin-bottom: 18px;
  padding: 16px 18px 4px;
  border: 1px solid #e3eaf4;
  border-radius: 12px;
  background: #fbfcfe;
}

.jc-container__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.jc-container__title {
  color: #1e293b;
  font-size: 15px;
  font-weight: 600;
}

.jc-container__title small {
  margin-left: 8px;
  color: #94a3b8;
  font-weight: 400;
  font-size: 12px;
}

.jc-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.jc-row--4 {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.jc-row--4 :deep(.el-form-item__label) {
  width: 80px !important;
}

.jc-add {
  width: 100%;
  border-style: dashed;
}

.mono-input :deep(input),
.mono-input :deep(textarea) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
