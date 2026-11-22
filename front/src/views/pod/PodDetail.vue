<template>
  <div class="job-detail detail-page">
    <PageHeader
      :title="pod?.name || name || 'Pod详情'"
      :subtitle="pod ? pod.alias || pod.describe || `命名空间：${pod.namespace}` : 'Pod 详情'"
      icon="pod"
      back
      @back="goBack"
    >
      <template v-if="pod" #meta>
        <el-tag :type="podStatusTag(pod.displayStatus)" effect="dark">{{ podStatusText(pod.displayStatus) }}</el-tag>
        <el-tag v-if="pod.qosClass" type="info" effect="light">{{ pod.qosClass }}</el-tag>
      </template>
      <template v-if="pod">
        <el-button :loading="loading" @click="fetchDetail()">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" :disabled="!canTerminal" @click="openTerminal()">
          <el-icon><Monitor /></el-icon>
          终端
        </el-button>
        <el-button @click="openLogs()">
          <el-icon><Tickets /></el-icon>
          查看日志
        </el-button>
        <el-button @click="infoVisible = true">
          <el-icon><Edit /></el-icon>
          编辑信息
        </el-button>
        <el-button @click="openImageDialog">
          <el-icon><Picture /></el-icon>
          更新镜像
        </el-button>
        <el-button @click="openYaml">
          <el-icon><Document /></el-icon>
          编辑YAML
        </el-button>
        <el-button type="warning" plain :disabled="!!pod.deletionTime" @click="handleExpel">驱逐</el-button>
        <el-button type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon>
          {{ pod.deletionTime ? '强制删除' : '删除' }}
        </el-button>
      </template>
    </PageHeader>

    <div v-if="loading && !pod" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="pod" class="detail-content">
      <el-card class="info-card jd-attr-card">
        <template #header>
          <div class="card-header" @click="attrOpen = !attrOpen">
            <span>属性</span>
            <el-icon :class="['jd-collapse', { 'is-collapsed': !attrOpen }]"><ArrowDown /></el-icon>
          </div>
        </template>
        <el-collapse-transition>
          <div v-show="attrOpen" class="jd-attr-grid">
            <div class="jd-attr"><label>集群</label><span>{{ clusterName || clusterId }}</span></div>
            <div class="jd-attr">
              <label>命名空间</label>
              <span><el-link type="primary" @click="goNamespace">{{ pod.namespace }}</el-link></span>
            </div>
            <div class="jd-attr">
              <label>状态</label>
              <span>
                {{ podStatusText(pod.displayStatus) }}
                <span v-if="pod.message" class="jd-msg">{{ pod.message }}</span>
              </span>
            </div>
            <div class="jd-attr"><label>就绪容器</label><span>{{ pod.ready }}</span></div>
            <div class="jd-attr">
              <label>CPU 用量</label>
              <span><ResourceUsageBar kind="cpu" :used="pod.cpuUsage" :request="pod.cpuRequest" :limit="pod.cpuLimit" /></span>
            </div>
            <div class="jd-attr">
              <label>内存用量</label>
              <span><ResourceUsageBar kind="memory" :used="pod.memoryUsage" :request="pod.memoryRequest" :limit="pod.memoryLimit" /></span>
            </div>
            <div class="jd-attr">
              <label>CPU 预留 / 上限</label>
              <span>{{ pod.cpuRequest ? formatCpu(pod.cpuRequest) : '未设置' }} / {{ pod.cpuLimit ? formatCpu(pod.cpuLimit) : '不限制' }}</span>
            </div>
            <div class="jd-attr">
              <label>内存预留 / 上限</label>
              <span>{{ pod.memoryRequest ? formatMemory(pod.memoryRequest) : '未设置' }} / {{ pod.memoryLimit ? formatMemory(pod.memoryLimit) : '不限制' }}</span>
            </div>
            <div class="jd-attr"><label>重启次数</label><span :class="{ 'pd-warn': pod.restarts > 0 }">{{ pod.restarts }}</span></div>
            <div class="jd-attr">
              <label>所属工作负载</label>
              <span>
                <el-link v-if="ownerPath" type="primary" @click="goOwner">{{ pod.ownerName }}</el-link>
                <template v-else>{{ pod.ownerName || '无（单独创建）' }}</template>
                <span v-if="pod.ownerKind" class="jd-sub">{{ pod.ownerKind }}</span>
              </span>
            </div>
            <div class="jd-attr">
              <label>节点</label>
              <span>
                <el-link v-if="pod.nodeName" type="primary" @click="goNode">{{ pod.nodeName }}</el-link>
                <template v-else>未调度</template>
              </span>
            </div>
            <div class="jd-attr"><label>容器组 IP</label><span class="mono">{{ pod.podIPs.length ? pod.podIPs.join(', ') : pod.podIP || '-' }}</span></div>
            <div class="jd-attr"><label>主机 IP</label><span class="mono">{{ pod.hostIP || '-' }}</span></div>
            <div class="jd-attr"><label>QoS 类别</label><span>{{ QOS_TEXT[pod.qosClass] || pod.qosClass || '-' }}</span></div>
            <div class="jd-attr"><label>重启策略</label><span>{{ RESTART_POLICY_TEXT[pod.restartPolicy] || pod.restartPolicy || '-' }}</span></div>
            <div class="jd-attr"><label>服务账户</label><span>{{ pod.serviceAccountName || 'default' }}</span></div>
            <div class="jd-attr"><label>网络</label><span>{{ pod.hostNetwork ? '节点网络（hostNetwork）' : '容器网络' }}</span></div>
            <div v-if="pod.priorityClassName" class="jd-attr"><label>优先级类</label><span>{{ pod.priorityClassName }}</span></div>
            <div class="jd-attr"><label>启动时间</label><span>{{ formatDateTime(pod.startTime) }}</span></div>
            <div class="jd-attr"><label>创建时间</label><span>{{ formatDateTime(pod.createTime) }}</span></div>
            <div class="jd-attr"><label>创建者</label><span>{{ pod.creator || '-' }}</span></div>
            <div v-if="pod.deletionTime" class="jd-attr"><label>开始删除时间</label><span class="pd-warn">{{ formatDateTime(pod.deletionTime) }}</span></div>
          </div>
        </el-collapse-transition>
      </el-card>

      <div class="jd-tabbar">
        <button v-for="t in tabs" :key="t.key" type="button" :class="['jd-tab', { 'is-active': activeTab === t.key }]" @click="activeTab = t.key">
          {{ t.label }}
          <span v-if="t.count !== undefined" class="jd-tab__count">{{ t.count }}</span>
        </button>
      </div>

      <!-- 资源状态 -->
      <template v-if="activeTab === 'resources'">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>容器</span>
              <el-tag size="small" type="info">{{ containerRows.length }}</el-tag>
              <span class="card-header__hint">文件下载和上传需要容器内有 tar 命令</span>
            </div>
          </template>
          <el-table :data="containerRows" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
            <el-table-column label="容器名称" min-width="170">
              <template #default="{ row }">
                <span class="pd-cname">{{ row.name }}</span>
                <el-tag v-if="row.init" size="small" type="warning" effect="plain">初始化</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="镜像" min-width="240" show-overflow-tooltip>
              <template #default="{ row }"><span class="mono">{{ row.image }}</span></template>
            </el-table-column>
            <el-table-column label="状态" min-width="190">
              <template #default="{ row }">
                <el-tag :type="containerStateTag(row.status)" size="small">{{ containerStateText(row.status) }}</el-tag>
                <span v-if="row.status?.message" class="jd-msg" :title="row.status.message">{{ row.status.message }}</span>
              </template>
            </el-table-column>
            <el-table-column label="CPU" width="140" align="center" header-align="center">
              <template #default="{ row }">
                <ResourceUsageBar kind="cpu" compact :used="row.usage?.cpuUsage" :request="row.usage?.cpuRequest" :limit="row.usage?.cpuLimit" />
              </template>
            </el-table-column>
            <el-table-column label="内存" width="150" align="center" header-align="center">
              <template #default="{ row }">
                <ResourceUsageBar kind="memory" compact :used="row.usage?.memoryUsage" :request="row.usage?.memoryRequest" :limit="row.usage?.memoryLimit" />
              </template>
            </el-table-column>
            <el-table-column label="重启次数" width="96" align="center" header-align="center">
              <template #default="{ row }">{{ row.status?.restartCount ?? 0 }}</template>
            </el-table-column>
            <el-table-column label="上次退出" min-width="200">
              <template #default="{ row }">
                <template v-if="row.status?.lastReason">
                  <span class="pd-warn">{{ row.status.lastReason }}（退出码 {{ row.status.lastExitCode }}）</span>
                  <span class="jd-muted pd-block">{{ formatDateTime(row.status.lastFinishedAt) }}</span>
                </template>
                <span v-else class="jd-muted">-</span>
              </template>
            </el-table-column>
            <el-table-column label="启动时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.status?.startedAt) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="190" align="center" header-align="center">
              <template #default="{ row }">
                <el-button-group size="small">
                  <el-button type="primary" title="进入终端" :disabled="row.status?.state !== 'Running' || !!pod?.deletionTime" @click="openTerminal(row.name)">
                    <el-icon><Monitor /></el-icon>
                  </el-button>
                  <el-button type="primary" title="查看日志" :disabled="!row.status || row.status.state === 'Waiting'" @click="openLogs(row.name)">
                    <el-icon><Tickets /></el-icon>
                  </el-button>
                  <el-button type="success" title="下载文件" :disabled="row.status?.state !== 'Running'" @click="openDownload(row.name)">
                    <el-icon><Download /></el-icon>
                  </el-button>
                  <el-button type="warning" title="上传文件" :disabled="row.status?.state !== 'Running'" @click="openUpload(row.name)">
                    <el-icon><Upload /></el-icon>
                  </el-button>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>容器设置</span>
              <span class="card-header__hint">端口、命令、资源配额、健康检查和存储挂载</span>
            </div>
          </template>
          <ContainerTemplates :containers="pod.containers" />
        </el-card>

        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>存储卷</span>
              <el-tag size="small" type="info">{{ pod.volumes.length }}</el-tag>
            </div>
          </template>
          <el-table :data="pod.volumes" style="width: 100%" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="名称" min-width="200" prop="name" />
            <el-table-column label="类型" width="140">
              <template #default="{ row }">{{ volumeTypeText(row.type) }}</template>
            </el-table-column>
            <el-table-column label="来源" min-width="300">
              <template #default="{ row }"><span class="mono">{{ row.source || '-' }}</span></template>
            </el-table-column>
            <template #empty>
              <span class="jd-muted">没有存储卷</span>
            </template>
          </el-table>
        </el-card>

        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>状态条件</span>
              <el-tag size="small" type="info">{{ pod.conditions.length }}</el-tag>
            </div>
          </template>
          <el-table :data="pod.conditions" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
            <el-table-column label="类型" min-width="170">
              <template #default="{ row }">{{ CONDITION_TEXT[row.type] || row.type }}</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'True' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="原因" min-width="160" prop="reason" />
            <el-table-column label="消息" min-width="300" prop="message" />
            <el-table-column label="更新时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.lastTime) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </template>

      <!-- 元数据 -->
      <template v-else-if="activeTab === 'metadata'">
        <el-card class="info-card">
          <template #header>
            <div class="card-header"><span>元数据</span></div>
          </template>
          <MetadataPanel :labels="pod.labels" :annotations="pod.annotations" />
        </el-card>
        <el-card class="info-card">
          <template #header>
            <div class="card-header"><span>调度</span></div>
          </template>
          <div class="jd-attr-grid">
            <div class="jd-attr"><label>节点选择器</label><span class="mono">{{ formatLabels(pod.nodeSelector) || '不限制' }}</span></div>
            <div class="jd-attr"><label>DNS 策略</label><span>{{ pod.dnsPolicy || '-' }}</span></div>
          </div>
          <h4 class="jd-section__title pd-tol-title">容忍（{{ pod.tolerations.length }}）</h4>
          <div v-if="pod.tolerations.length" class="pd-tags">
            <el-tag v-for="t in pod.tolerations" :key="t" type="info" class="mono">{{ t }}</el-tag>
          </div>
          <span v-else class="jd-muted">没有容忍设置</span>
        </el-card>
      </template>

      <el-card v-else-if="activeTab === 'env'" class="info-card">
        <template #header>
          <div class="card-header"><span>环境变量</span></div>
        </template>
        <ContainerEnvList :containers="pod.containers" />
      </el-card>

      <el-card v-else class="info-card">
        <template #header>
          <div class="card-header">
            <span>事件</span>
            <el-tag size="small" type="info">{{ pod.events.length }}</el-tag>
          </div>
        </template>
        <EventsTable :events="pod.events" />
      </el-card>
    </div>

    <el-empty v-else description="容器组不存在或已被删除" />

    <ResourceInfoDialog v-model="infoVisible" :name="pod?.name || ''" :alias="pod?.alias" :describe="pod?.describe" :submit="saveInfo" @success="fetchDetail()" />

    <el-dialog v-model="imageVisible" title="更新镜像" width="720px" :close-on-click-modal="false" class="config-dialog form-dialog--compact">
      <template #header>
        <div class="dialog-header"><h3 class="dialog-title">更新镜像</h3></div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <el-alert v-if="pod?.ownerKind" type="warning" :closable="false" show-icon class="pd-alert">
            该容器组由 {{ pod.ownerKind }} {{ pod.ownerName }} 管理，直接修改会在容器组重建后失效，建议在工作负载中更新镜像。
          </el-alert>
          <p class="pd-tip">修改后 kubelet 会重启对应的容器，容器组名称和 IP 保持不变。</p>
          <el-form label-width="140px">
            <el-form-item v-for="c in pod?.containers || []" :key="(c.init ? 'init-' : '') + c.name" :label="c.name">
              <el-input v-model="imageForm[c.name]" class="mono-input">
                <template v-if="c.init" #prepend>初始化</template>
              </el-input>
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="imageVisible = false">取消</el-button>
          <el-button type="primary" :loading="imageSaving" @click="saveImages">更新</el-button>
        </div>
      </template>
    </el-dialog>

    <YamlDialog
      v-model="yamlVisible"
      v-model:content="yamlContent"
      title="编辑YAML"
      subtitle="Pod 配置"
      :info="[{ label: '集群', value: clusterName || clusterId }, { label: '命名空间', value: pod?.namespace }, { label: '名称', value: pod?.name }, { label: '类型', value: 'Pod' }]"
      :editor-title="`${pod?.name} - Pod YAML`"
      :filename="`${pod?.name}-pod.yaml`"
      tip="Pod 创建后只能修改镜像、标签、注解等少量字段，其余修改会被拒绝"
      :apply-target="pod ? { clusterId, kind: 'pod', namespace: pod.namespace, name: pod.name } : null"
      @applied="fetchDetail()"
    />

    <!-- 文件下载 -->
    <el-dialog v-model="downloadVisible" title="下载文件" width="800px" :close-on-click-modal="false" class="config-dialog form-dialog--compact">
      <template #header>
        <div class="dialog-header"><h3 class="dialog-title">下载文件</h3></div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <el-alert type="warning" :closable="false" class="pd-alert">
            需要容器镜像中包含 <strong>tar</strong> 命令；文件以 <strong>.tar</strong> 格式下载，可使用 tar 命令解压。
          </el-alert>
          <el-form label-width="96px">
            <el-form-item label="容器">
              <el-input v-model="fileForm.containerName" disabled />
            </el-form-item>
            <el-form-item label="文件路径" required>
              <el-input v-model="fileForm.filePath" placeholder="文件或目录的路径，如 /app/config.json" class="mono-input" @keyup.enter="confirmDownload" />
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="downloadVisible = false">取消</el-button>
          <el-button type="primary" :loading="fileLoading" :disabled="!fileForm.filePath.trim()" @click="confirmDownload">下载</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 文件上传 -->
    <el-dialog v-model="uploadVisible" title="上传文件" width="800px" :close-on-click-modal="false" class="config-dialog form-dialog--compact">
      <template #header>
        <div class="dialog-header"><h3 class="dialog-title">上传文件</h3></div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <el-form label-width="96px">
            <el-form-item label="容器">
              <el-input v-model="fileForm.containerName" disabled />
            </el-form-item>
            <el-form-item label="目标路径" required>
              <el-input v-model="fileForm.filePath" placeholder="目标目录或文件路径，如 /tmp/ 或 /app/config.json" class="mono-input" />
            </el-form-item>
            <el-form-item label="选择文件" required>
              <el-upload ref="uploadRef" :auto-upload="false" :limit="1" :on-change="onFileChange" :on-remove="() => (selectedFile = null)">
                <el-button type="primary">选择文件</el-button>
              </el-upload>
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="uploadVisible = false">取消</el-button>
          <el-button type="primary" :loading="fileLoading" :disabled="!fileForm.filePath.trim() || !selectedFile" @click="confirmUpload">上传</el-button>
        </div>
      </template>
    </el-dialog>

    <PodTerminal v-model="terminalVisible" :cluster-id="clusterId" :namespace="namespace" :pod-name="name" :container="terminalContainer" />

    <PodLogViewer
      v-model="logVisible"
      v-model:containerName="logContainer"
      :title="`查看日志 - ${pod?.name || ''}`"
      :cluster-id="clusterId"
      :namespace="pod?.namespace || ''"
      :pod-name="pod?.name || ''"
      :containers="logContainers"
      :container-selectable="true"
      :allow-all-containers="true"
    />
  </div>
</template>

<script setup lang="ts">
import { formatLabels } from '@/utils/format'
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type UploadFile, type UploadInstance } from 'element-plus'
import { ArrowDown, Delete, Document, Download, Edit, Monitor, Picture, Refresh, Tickets, Upload } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import PodLogViewer from '@/components/PodLogViewer.vue'
import PodTerminal from '@/components/pod/PodTerminal.vue'
import ResourceUsageBar from '@/components/ResourceUsageBar.vue'
import ResourceInfoDialog from '@/components/ResourceInfoDialog.vue'
import ContainerTemplates from '@/components/job/ContainerTemplates.vue'
import ContainerEnvList from '@/components/job/ContainerEnvList.vue'
import MetadataPanel from '@/components/job/MetadataPanel.vue'
import EventsTable from '@/components/job/EventsTable.vue'
import { podApi, type PodDetailVO } from '@/api/pod'
import { fetchResourceYaml } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useTabBack } from '@/composables/useTabBack'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime, podStatusTag, podStatusText } from '@/utils/job'
import { ownerDetailPath, volumeTypeText } from '@/utils/workload'
import { QOS_TEXT, RESTART_POLICY_TEXT, containerStateTag, containerStateText } from '@/utils/pod'
import { formatCpu, formatMemory } from '@/utils/resource'
import '@/assets/styles/config-editor.css'
import '@/assets/styles/job-detail.css'

type TabKey = 'resources' | 'metadata' | 'env' | 'events'

const CONDITION_TEXT: Record<string, string> = {
  PodScheduled: '已调度（PodScheduled）',
  Initialized: '初始化完成（Initialized）',
  ContainersReady: '容器就绪（ContainersReady）',
  Ready: '容器组就绪（Ready）',
  PodReadyToStartContainers: '网络就绪（PodReadyToStartContainers）',
  DisruptionTarget: '即将中断（DisruptionTarget）'
}

const route = useRoute()
const router = useRouter()
// 返回当前标签内的上一页
const goBack = useTabBack()
const clusterStore = useClusterStore()

// 页面被 keep-alive 按完整路径缓存, 参数在创建时取一次
const clusterId = ref((route.params.clusterId as string) || (route.query.clusterId as string) || clusterStore.currentClusterId || '')
const namespace = ref((route.query.namespace as string) || '')
const name = ref((route.query.name as string) || '')
const clusterName = computed(() => clusterStore.currentClusterName)

const loading = ref(false)
const pod = ref<PodDetailVO | null>(null)
const attrOpen = ref(true)
const activeTab = ref<TabKey>('resources')
const tabs = computed(() => [
  { key: 'resources' as TabKey, label: '资源状态', count: pod.value?.containers.length },
  { key: 'metadata' as TabKey, label: '元数据' },
  { key: 'env' as TabKey, label: '环境变量' },
  { key: 'events' as TabKey, label: '事件', count: pod.value?.events.length }
])

// 容器设置与运行状态按名称合并, 初始化容器在前
const containerRows = computed(() =>
  (pod.value?.containers || []).map(c => ({
    name: c.name,
    init: c.init,
    image: c.image,
    status: pod.value?.containerStatuses.find(s => s.name === c.name && s.init === c.init),
    usage: pod.value?.containerUsages?.find(u => u.name === c.name && u.init === c.init)
  }))
)
const logContainers = computed(() => (pod.value?.containers || []).map(c => c.name))

const ownerPath = computed(() => ownerDetailPath(pod.value?.ownerKind))

const target = () => ({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })

// 启动中、删除中或有容器未就绪时定时刷新; 页面切走后停止
const active = ref(true)
let timer: ReturnType<typeof setTimeout> | undefined
const needPoll = (p: PodDetailVO) =>
  !!p.deletionTime || p.status === 'Pending' || (p.status === 'Running' && p.containerStatuses.some(s => !s.init && !s.ready))
const schedulePoll = () => {
  clearTimeout(timer)
  if (active.value && pod.value && needPoll(pod.value)) timer = setTimeout(() => fetchDetail(true), 5000)
}

const fetchDetail = async (silent = false) => {
  if (!clusterId.value || !namespace.value || !name.value) return
  if (!silent) loading.value = true
  try {
    const resp = await podApi.getDetail(target())
    pod.value = resp.code === 200 ? resp.data : null
  } catch {
    // 删除完成后轮询会查不到, 保留最后一次的数据
    if (!silent) pod.value = null
  } finally {
    loading.value = false
    schedulePoll()
  }
}

// 终端
const terminalVisible = ref(false)
const terminalContainer = ref('')
const canTerminal = computed(() => !!pod.value && !pod.value.deletionTime && pod.value.containerStatuses.some(s => s.state === 'Running'))
const openTerminal = (container?: string) => {
  terminalContainer.value = container || ''
  terminalVisible.value = true
}

// 日志
const logVisible = ref(false)
const logContainer = ref('')
const openLogs = (container?: string) => {
  logContainer.value = container || pod.value?.containers.find(c => !c.init)?.name || ''
  logVisible.value = true
}

// 编辑信息
const infoVisible = ref(false)
const saveInfo = (alias: string, describe: string) => podApi.updateInfo({ ...target(), alias, describe })

// 更新镜像
const imageVisible = ref(false)
const imageSaving = ref(false)
const imageForm = reactive<Record<string, string>>({})
const openImageDialog = () => {
  Object.keys(imageForm).forEach(k => delete imageForm[k])
  pod.value?.containers.forEach(c => (imageForm[c.name] = c.image))
  imageVisible.value = true
}
const saveImages = async () => {
  const images: Record<string, string> = {}
  pod.value?.containers.forEach(c => {
    const v = (imageForm[c.name] || '').trim()
    if (v && v !== c.image) images[c.name] = v
  })
  if (!Object.keys(images).length) {
    ElMessage.info('镜像没有变化')
    return
  }
  imageSaving.value = true
  try {
    await podApi.updateImages({ ...target(), images })
    ElMessage.success('镜像已更新，容器正在重启')
    imageVisible.value = false
    fetchDetail()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    imageSaving.value = false
  }
}

// YAML
const yamlVisible = ref(false)
const yamlContent = ref('')
const openYaml = async () => {
  try {
    yamlContent.value = await fetchResourceYaml({ ...target(), kind: 'pod' })
    yamlVisible.value = true
  } catch {
    // 请求拦截器已提示错误
  }
}

// 文件下载 / 上传
const downloadVisible = ref(false)
const uploadVisible = ref(false)
const fileLoading = ref(false)
const fileForm = reactive({ containerName: '', filePath: '' })
const selectedFile = ref<File | null>(null)
const uploadRef = ref<UploadInstance>()

const openDownload = (container: string) => {
  Object.assign(fileForm, { containerName: container, filePath: '' })
  downloadVisible.value = true
}
const openUpload = (container: string) => {
  Object.assign(fileForm, { containerName: container, filePath: '/tmp/' })
  selectedFile.value = null
  uploadRef.value?.clearFiles()
  uploadVisible.value = true
}
const onFileChange = (file: UploadFile) => {
  selectedFile.value = file.raw || null
}

const confirmDownload = async () => {
  const filePath = fileForm.filePath.trim()
  if (!filePath) return
  fileLoading.value = true
  try {
    const blob = await podApi.downloadFile({ ...target(), containerName: fileForm.containerName, filePath })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${filePath.replace(/\/+$/, '').split('/').pop() || 'download'}.tar`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    ElMessage.success('文件下载成功')
    downloadVisible.value = false
  } catch {
    // 请求拦截器已提示错误
  } finally {
    fileLoading.value = false
  }
}

const confirmUpload = async () => {
  const filePath = fileForm.filePath.trim()
  if (!filePath || !selectedFile.value) return
  fileLoading.value = true
  try {
    await podApi.uploadFile({ ...target(), containerName: fileForm.containerName, filePath }, selectedFile.value)
    ElMessage.success('文件上传成功')
    uploadVisible.value = false
  } catch {
    // 请求拦截器已提示错误
  } finally {
    fileLoading.value = false
  }
}

const handleExpel = async () => {
  try {
    await ElMessageBox.confirm(
      `驱逐会遵守 PodDisruptionBudget，由所属的工作负载在其他节点重建容器组${pod.value?.ownerKind ? '' : '；该容器组没有所属工作负载，驱逐后不会重建'}。`,
      `驱逐 "${name.value}"`,
      { type: 'warning', confirmButtonText: '驱逐', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await podApi.expel(target())
    ElMessage.success('已发起驱逐')
    fetchDetail()
  } catch {
    // 请求拦截器已提示错误
  }
}

const handleDelete = async () => {
  if (!pod.value) return
  const force = !!pod.value.deletionTime
  try {
    await ElMessageBox.confirm(
      force
        ? '容器组一直处于删除中，强制删除会立即从集群中移除记录，不再等待容器退出（节点失联时容器可能仍在运行）。确定强制删除吗？'
        : `确定要删除容器组 "${name.value}" 吗？${pod.value.ownerKind ? `它由 ${pod.value.ownerKind} ${pod.value.ownerName} 管理，删除后会自动重建。` : '它没有所属工作负载，删除后不会重建。'}`,
      force ? '强制删除' : '确认删除',
      { type: 'warning', confirmButtonText: force ? '强制删除' : '确定删除', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await podApi.delete({ ...target(), force })
    ElMessage.success(force ? '已强制删除' : '已开始删除')
    goBack()
  } catch {
    // 请求拦截器已提示错误
  }
}

const base = computed(() => `/cluster/${clusterId.value}`)
const goNamespace = () => router.push({ path: `${base.value}/namespace/detail`, query: { namespace: namespace.value } })
const goNode = () => router.push({ path: `${base.value}/node/detail`, query: { hostname: pod.value?.nodeName } })
const goOwner = () => {
  if (!ownerPath.value || !pod.value) return
  router.push({ path: `${base.value}/${ownerPath.value}`, query: { namespace: namespace.value, name: pod.value.ownerName } })
}

onMounted(() => fetchDetail())
onActivated(() => {
  if (!active.value) {
    active.value = true
    fetchDetail(true)
  }
})
onDeactivated(() => {
  active.value = false
  clearTimeout(timer)
})
onBeforeUnmount(() => clearTimeout(timer))
</script>

<style scoped>
.pd-cname {
  margin-right: 6px;
  color: #1e293b;
  font-weight: 600;
}

.pd-warn {
  color: #f59e0b;
  font-weight: 600;
}

.pd-block {
  display: block;
  font-size: 12px;
}

.pd-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.pd-tol-title {
  margin-top: 18px;
}

.pd-alert {
  margin-bottom: 16px;
}

.pd-tip {
  margin: 0 0 14px;
  color: #64748b;
  font-size: 13px;
}

.mono-input :deep(input) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
