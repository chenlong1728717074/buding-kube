<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="800px"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--compact"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="load"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">{{ title }}</h3>
      </div>
    </template>
    <div v-loading="loading" class="config-editor">
      <div class="config-content">
        <el-form :model="form" label-width="120px">
          <el-form-item label="名称">
            <el-input :model-value="target ? `${target.namespace} / ${target.name}` : ''" disabled />
          </el-form-item>
          <el-form-item label="别名">
            <el-input v-model="form.alias" placeholder="请输入别名" maxlength="50" show-word-limit />
          </el-form-item>
          <el-form-item label="描述">
            <el-input v-model="form.describe" type="textarea" :rows="3" placeholder="请输入描述" maxlength="200" show-word-limit />
          </el-form-item>

          <template v-if="kind === 'job'">
            <el-form-item label="并行数量">
              <el-input-number v-model="form.parallelism" :min="0" :max="10000" controls-position="right" />
              <span class="fd-inline-hint">设为 0 可以暂停创建新的容器组</span>
            </el-form-item>
            <el-form-item label="最大运行时间">
              <el-input-number v-model="form.activeDeadlineSeconds" :min="0" :max="31536000" controls-position="right" />
              <span class="fd-inline-hint">秒，0 表示不限制</span>
            </el-form-item>
            <el-form-item label="结束后自动删除">
              <el-switch v-model="form.autoClean" />
              <template v-if="form.autoClean">
                <el-input-number v-model="form.ttlSecondsAfterFinished" :min="0" :max="31536000" controls-position="right" style="margin-left: 12px" />
                <span class="fd-inline-hint">秒</span>
              </template>
            </el-form-item>
            <p class="je-tip">容器、完成数量、重试次数等设置在任务创建后不能修改；如需调整，请编辑 YAML 后重新执行，或重新创建任务。</p>
          </template>

          <template v-else>
            <el-form-item label="调度规则" required>
              <el-input v-model="form.schedule" placeholder="分 时 日 月 周，例如 0 2 * * *" />
              <div class="field-hint">{{ scheduleDesc ? `执行时间：${scheduleDesc}` : '分(0-59) 时(0-23) 日(1-31) 月(1-12) 周(0-6)' }}</div>
            </el-form-item>
            <el-form-item label="时区">
              <el-select v-model="form.timeZone" filterable allow-create clearable placeholder="不设置时使用集群控制器所在时区" style="width: 100%">
                <el-option v-for="tz in TIMEZONE_OPTIONS" :key="tz" :label="tz" :value="tz" />
              </el-select>
            </el-form-item>
            <el-form-item label="并发策略">
              <el-radio-group v-model="form.concurrencyPolicy">
                <el-radio-button v-for="p in CONCURRENCY_POLICIES" :key="p.value" :value="p.value">{{ p.label }}</el-radio-button>
              </el-radio-group>
              <div class="field-hint">{{ CONCURRENCY_POLICIES.find(p => p.value === form.concurrencyPolicy)?.desc }}</div>
            </el-form-item>
            <el-form-item label="启动截止时间">
              <el-input-number v-model="form.startingDeadlineSeconds" :min="0" :max="86400" controls-position="right" />
              <span class="fd-inline-hint">秒，0 表示不限制</span>
            </el-form-item>
            <el-form-item label="保留成功任务">
              <el-input-number v-model="form.successfulJobsHistoryLimit" :min="0" :max="100" controls-position="right" />
            </el-form-item>
            <el-form-item label="保留失败任务">
              <el-input-number v-model="form.failedJobsHistoryLimit" :min="0" :max="100" controls-position="right" />
            </el-form-item>
            <el-form-item label="暂停调度">
              <el-switch v-model="form.suspend" />
            </el-form-item>
            <p class="je-tip">容器等任务模板请通过「编辑YAML」修改，只对之后创建的任务生效。</p>
          </template>
        </el-form>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="submitting" :disabled="loading || !loaded" @click="submit">保存</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { jobApi } from '@/api/job'
import { cronJobApi, type ConcurrencyPolicy } from '@/api/cronjob'
import { CONCURRENCY_POLICIES, TIMEZONE_OPTIONS, describeCron } from '@/utils/job'

const props = defineProps<{
  modelValue: boolean
  clusterId: string
  kind: 'job' | 'cronjob'
  target: { namespace: string; name: string } | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: []
}>()

const title = computed(() => (props.kind === 'job' ? '编辑任务' : '编辑定时任务'))
const loading = ref(false)
const loaded = ref(false)
const submitting = ref(false)

const form = reactive({
  alias: '',
  describe: '',
  parallelism: 1,
  activeDeadlineSeconds: 0,
  autoClean: false,
  ttlSecondsAfterFinished: 3600,
  schedule: '',
  timeZone: '',
  concurrencyPolicy: 'Allow' as ConcurrencyPolicy,
  startingDeadlineSeconds: 0,
  successfulJobsHistoryLimit: 3,
  failedJobsHistoryLimit: 1,
  suspend: false
})

const scheduleDesc = computed(() => describeCron(form.schedule))

// 打开时读取最新详情, 列表项里没有完整的设置
const load = async () => {
  loaded.value = false
  if (!props.target) return
  loading.value = true
  const params = { clusterId: props.clusterId, namespace: props.target.namespace, name: props.target.name }
  try {
    if (props.kind === 'job') {
      const { data } = await jobApi.getDetail(params)
      Object.assign(form, {
        alias: data.alias || '',
        describe: data.describe || '',
        parallelism: data.parallelism,
        activeDeadlineSeconds: data.activeDeadlineSeconds || 0,
        autoClean: data.ttlSecondsAfterFinished !== null && data.ttlSecondsAfterFinished !== undefined,
        ttlSecondsAfterFinished: data.ttlSecondsAfterFinished ?? 3600
      })
    } else {
      const { data } = await cronJobApi.getDetail(params)
      Object.assign(form, {
        alias: data.alias || '',
        describe: data.describe || '',
        schedule: data.schedule,
        timeZone: data.timeZone || '',
        concurrencyPolicy: data.concurrencyPolicy || 'Allow',
        startingDeadlineSeconds: data.startingDeadlineSeconds || 0,
        successfulJobsHistoryLimit: data.successfulJobsHistoryLimit,
        failedJobsHistoryLimit: data.failedJobsHistoryLimit,
        suspend: data.suspend
      })
    }
    loaded.value = true
  } catch {
    emit('update:modelValue', false)
  } finally {
    loading.value = false
  }
}

const submit = async () => {
  if (!props.target) return
  const base = {
    clusterId: props.clusterId,
    namespace: props.target.namespace,
    name: props.target.name,
    alias: form.alias.trim(),
    describe: form.describe.trim()
  }
  if (props.kind === 'cronjob') {
    const v = form.schedule.trim()
    if (!v || (!v.startsWith('@') && v.split(/\s+/).length !== 5)) {
      ElMessage.warning('调度规则需要 5 个字段：分 时 日 月 周')
      return
    }
  }
  submitting.value = true
  try {
    if (props.kind === 'job') {
      await jobApi.update({
        ...base,
        parallelism: form.parallelism,
        activeDeadlineSeconds: form.activeDeadlineSeconds,
        ttlSecondsAfterFinished: form.autoClean ? form.ttlSecondsAfterFinished : -1
      })
    } else {
      await cronJobApi.update({
        ...base,
        schedule: form.schedule.trim(),
        timeZone: form.timeZone || '',
        concurrencyPolicy: form.concurrencyPolicy,
        startingDeadlineSeconds: form.startingDeadlineSeconds,
        successfulJobsHistoryLimit: form.successfulJobsHistoryLimit,
        failedJobsHistoryLimit: form.failedJobsHistoryLimit,
        suspend: form.suspend
      })
    }
    ElMessage.success('保存成功')
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
.je-tip {
  margin: 4px 0 0 120px;
  color: #94a3b8;
  font-size: 12px;
  line-height: 1.6;
}
</style>
