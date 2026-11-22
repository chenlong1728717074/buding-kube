<template>
  <div :class="['wp-probe', { 'is-on': probe.enabled }]">
    <div class="wp-probe__head">
      <el-switch v-model="probe.enabled" />
      <span class="wp-probe__title">{{ title }}</span>
      <span class="wp-probe__desc">{{ desc }}</span>
    </div>
    <div v-if="probe.enabled" class="wp-probe__body">
      <el-radio-group v-model="probe.type" size="small">
        <el-radio-button value="http">HTTP 请求</el-radio-button>
        <el-radio-button value="tcp">TCP 端口</el-radio-button>
        <el-radio-button value="exec">执行命令</el-radio-button>
      </el-radio-group>
      <div v-if="probe.type !== 'exec'" class="wp-probe__row">
        <el-select v-if="probe.type === 'http'" v-model="probe.scheme" style="width: 100px">
          <el-option label="HTTP" value="HTTP" />
          <el-option label="HTTPS" value="HTTPS" />
        </el-select>
        <el-select v-model="probe.port" filterable allow-create default-first-option placeholder="端口" style="width: 130px" @change="onPortChange">
          <el-option v-for="p in portOptions" :key="p" :label="String(p)" :value="p" />
        </el-select>
        <el-input v-if="probe.type === 'http'" v-model="probe.path" placeholder="/healthz" style="flex: 1" />
      </div>
      <el-input
        v-else
        v-model="probe.commandText"
        type="textarea"
        :rows="2"
        placeholder="每行一项，例如&#10;cat&#10;/tmp/healthy"
        class="mono-input"
      />
      <div class="wp-probe__grid">
        <label>初始延迟<el-input-number v-model="probe.initialDelaySeconds" :min="0" :max="3600" size="small" controls-position="right" /><em>秒</em></label>
        <label>检查间隔<el-input-number v-model="probe.periodSeconds" :min="1" :max="3600" size="small" controls-position="right" /><em>秒</em></label>
        <label>超时时间<el-input-number v-model="probe.timeoutSeconds" :min="1" :max="3600" size="small" controls-position="right" /><em>秒</em></label>
        <label>失败阈值<el-input-number v-model="probe.failureThreshold" :min="1" :max="100" size="small" controls-position="right" /><em>次</em></label>
        <label v-if="allowSuccess">成功阈值<el-input-number v-model="probe.successThreshold" :min="1" :max="100" size="small" controls-position="right" /><em>次</em></label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProbeForm } from '@/utils/workload'

// probe 由父组件创建, 这里直接编辑它的字段
const props = withDefaults(
  defineProps<{
    probe: ProbeForm
    title: string
    desc: string
    allowSuccess?: boolean
    portOptions?: number[]
  }>(),
  { allowSuccess: false, portOptions: () => [] }
)

// 允许手动输入端口
const onPortChange = (value: unknown) => {
  const n = Number(value)
  props.probe.port = Number.isInteger(n) && n > 0 ? n : undefined
}
</script>

<style scoped>
.wp-probe {
  padding: 10px 14px;
  border: 1px solid #e6ecf5;
  border-radius: 10px;
  background: #fff;
}

.wp-probe + .wp-probe {
  margin-top: 10px;
}

.wp-probe.is-on {
  border-color: #c9dafc;
  background: #f8fbff;
}

.wp-probe__head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.wp-probe__title {
  color: #1e293b;
  font-weight: 600;
}

.wp-probe__desc {
  color: #94a3b8;
  font-size: 12px;
}

.wp-probe__body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 10px;
}

.wp-probe__row {
  display: flex;
  gap: 8px;
}

.wp-probe__grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 18px;
}

.wp-probe__grid label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #64748b;
  font-size: 13px;
}

.wp-probe__grid em {
  color: #94a3b8;
  font-style: normal;
  font-size: 12px;
}

.wp-probe__grid :deep(.el-input-number) {
  width: 96px;
}

.mono-input :deep(textarea) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
