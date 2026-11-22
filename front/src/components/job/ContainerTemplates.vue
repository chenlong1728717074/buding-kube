<template>
  <el-empty v-if="!containers.length" description="没有容器" :image-size="72" />
  <div v-else class="jd-containers">
    <div v-for="c in containers" :key="(c.init ? 'init-' : '') + c.name" class="jd-container">
      <div class="jd-container__head">
        <span class="jd-container__name">{{ c.name }}</span>
        <el-tag v-if="c.init" size="small" type="warning">初始化容器</el-tag>
        <el-tag v-else size="small" type="primary">工作容器</el-tag>
      </div>
      <div class="jd-container__row"><label>镜像</label><span class="mono">{{ c.image }}</span></div>
      <div class="jd-container__row"><label>拉取策略</label><span>{{ c.imagePullPolicy || '-' }}</span></div>
      <div class="jd-container__row"><label>命令</label><span class="mono">{{ joinArgs(c.command) || '镜像默认' }}</span></div>
      <div v-if="c.args?.length" class="jd-container__row"><label>参数</label><span class="mono">{{ joinArgs(c.args) }}</span></div>
      <div v-if="c.workingDir" class="jd-container__row"><label>工作目录</label><span class="mono">{{ c.workingDir }}</span></div>
      <div class="jd-container__row"><label>资源预留</label><span>{{ formatResources(c.requests) }}</span></div>
      <div class="jd-container__row"><label>资源上限</label><span>{{ formatResources(c.limits) }}</span></div>
      <div v-if="c.ports?.length" class="jd-container__row">
        <label>端口</label>
        <span class="mono">{{ c.ports.map(p => (p.name ? p.name + ' ' : '') + p.containerPort + '/' + p.protocol).join('，') }}</span>
      </div>
      <div v-if="c.readinessProbe" class="jd-container__row"><label>就绪检查</label><span>{{ c.readinessProbe }}</span></div>
      <div v-if="c.livenessProbe" class="jd-container__row"><label>存活检查</label><span>{{ c.livenessProbe }}</span></div>
      <div v-if="c.startupProbe" class="jd-container__row"><label>启动检查</label><span>{{ c.startupProbe }}</span></div>
      <div v-if="c.volumeMounts?.length" class="jd-container__row">
        <label>存储挂载</label>
        <span class="mono">{{ c.volumeMounts.map(m => `${m.name} → ${m.mountPath}${m.subPath ? '（' + m.subPath + '）' : ''}${m.readOnly ? ' 只读' : ''}`).join('\n') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ContainerTemplateVO } from '@/api/job'

defineProps<{ containers: ContainerTemplateVO[] }>()

// 含空白的参数加引号, 方便看清每一项的边界
const joinArgs = (items?: string[] | null) =>
  (items || []).map(s => (/\s/.test(s) && !s.includes('\n') ? JSON.stringify(s) : s)).join(' ')

const formatResources = (res?: Record<string, string> | null) => {
  if (!res || !Object.keys(res).length) return '不限制'
  const names: Record<string, string> = { cpu: 'CPU', memory: '内存', 'ephemeral-storage': '临时存储' }
  return Object.entries(res)
    .map(([k, v]) => `${names[k] || k} ${v}`)
    .join('，')
}
</script>
