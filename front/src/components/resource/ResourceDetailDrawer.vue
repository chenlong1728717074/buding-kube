<template>
  <DetailDrawer :model-value="modelValue" :kind="kindLabel" :name="name" :meta="meta" @update:model-value="(v: boolean) => emit('update:modelValue', v)">
    <div v-loading="loading">
      <el-tabs v-model="tab" class="ks-tabs">
        <el-tab-pane label="概览" name="info">
          <template v-if="obj">
            <div class="rd-grid">
              <div class="rd-item"><label>名称</label><span>{{ obj.metadata?.name }}</span></div>
              <div v-if="obj.metadata?.namespace" class="rd-item"><label>命名空间</label><span>{{ obj.metadata.namespace }}</span></div>
              <div class="rd-item"><label>API 版本</label><span class="mono">{{ obj.apiVersion }}</span></div>
              <div class="rd-item"><label>创建时间</label><span>{{ formatDateTime(obj.metadata?.creationTimestamp) }}</span></div>
              <div v-if="owner" class="rd-item"><label>所属</label><span>{{ owner }}</span></div>
              <div class="rd-item"><label>UID</label><span class="mono">{{ obj.metadata?.uid }}</span></div>
            </div>
            <slot name="summary" :obj="obj" />
            <template v-if="conditions.length">
              <h4 class="rd-title">状态条件</h4>
              <el-table :data="conditions" size="small" style="width: 100%">
                <el-table-column label="类型" prop="type" min-width="160" />
                <el-table-column label="状态" width="90">
                  <template #default="{ row }">
                    <el-tag :type="row.status === 'True' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="原因" prop="reason" min-width="140" />
                <el-table-column label="消息" prop="message" min-width="260" />
              </el-table>
            </template>
            <h4 class="rd-title">元数据</h4>
            <MetadataPanel :labels="obj.metadata?.labels" :annotations="obj.metadata?.annotations" />
          </template>
        </el-tab-pane>
        <el-tab-pane :label="`事件（${events.length}）`" name="events">
          <EventsTable :events="events" />
        </el-tab-pane>
        <el-tab-pane label="YAML" name="yaml">
          <div class="rd-yaml-actions">
            <el-button size="small" type="primary" @click="emit('edit')">编辑YAML</el-button>
          </div>
          <YamlEditor :model-value="yaml" :readonly="true" height="calc(100vh - 320px)" />
        </el-tab-pane>
      </el-tabs>
    </div>
  </DetailDrawer>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import DetailDrawer from '@/components/DetailDrawer.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import MetadataPanel from '@/components/job/MetadataPanel.vue'
import EventsTable from '@/components/job/EventsTable.vue'
import { fetchResourceYaml, resourceApi, type KubeObject, type ResourceRef } from '@/api/resource'
import { eventApi } from '@/api/event'
import type { EventVO } from '@/api/job'
import { formatDateTime } from '@/utils/job'

// 通用资源详情: 概览(元数据、状态条件)、事件、YAML
const props = defineProps<{
  modelValue: boolean
  clusterId: string
  resource: ResourceRef
  kindLabel: string
  namespace?: string
  name: string
}>()

const emit = defineEmits<{ 'update:modelValue': [value: boolean]; edit: [] }>()

const tab = ref('info')
const loading = ref(false)
const obj = ref<KubeObject | null>(null)
const events = ref<EventVO[]>([])
const yaml = ref('')

const conditions = computed<{ type: string; status: string; reason?: string; message?: string }[]>(() =>
  Array.isArray(obj.value?.status?.conditions) ? obj.value!.status.conditions : []
)
const owner = computed(() => {
  const ref = (obj.value?.metadata?.ownerReferences || []).find((r: KubeObject) => r.controller) || obj.value?.metadata?.ownerReferences?.[0]
  return ref ? `${ref.kind} ${ref.name}` : ''
})
const meta = computed(() => [
  { label: '命名空间', value: props.namespace || '集群级' },
  { label: '创建时间', value: formatDateTime(obj.value?.metadata?.creationTimestamp) }
])

const load = async () => {
  const target = { ...props.resource, clusterId: props.clusterId, namespace: props.namespace || undefined, name: props.name }
  tab.value = 'info'
  loading.value = true
  obj.value = null
  events.value = []
  yaml.value = ''
  try {
    const [objResp, yamlText] = await Promise.all([resourceApi.object(target), fetchResourceYaml(target)])
    obj.value = objResp.data
    yaml.value = yamlText
    const uid = objResp.data?.metadata?.uid
    if (uid) {
      const ev = await eventApi.list({ clusterId: props.clusterId, namespace: props.namespace || undefined, uid, page: 1, pageSize: 200 })
      events.value = ev.data?.items || []
    }
  } catch {
    // 请求拦截器已提示错误
  } finally {
    loading.value = false
  }
}

watch(
  () => props.modelValue,
  v => {
    if (v) load()
  }
)

defineExpose({ reload: load })
</script>

<style scoped>
.rd-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 24px;
  margin-bottom: 18px;
}

.rd-item {
  display: flex;
  gap: 12px;
  font-size: 13px;
}

.rd-item label {
  flex: 0 0 72px;
  color: #64748b;
}

.rd-item span {
  color: #1e293b;
  word-break: break-all;
}

.rd-title {
  margin: 18px 0 10px;
  color: #334155;
  font-size: 14px;
  font-weight: 600;
}

.rd-yaml-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 8px;
}

.mono {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12px;
}
</style>
