<template>
  <div class="kubectl-page list-page">
    <PageHeader title="Web kubectl" subtitle="在浏览器中使用 kubectl / helm 管理当前集群，拥有集群管理员权限" icon="k8s">
      <template v-if="started" #meta>
        <TerminalStatusTag :status="status" />
      </template>
      <template v-if="started">
        <TerminalToolbar :view="view" :closable="false" @reconnect="connect" />
        <el-button size="small" :disabled="status === 'closed'" @click="disconnect">
          <el-icon><SwitchButton /></el-icon>
          断开
        </el-button>
      </template>
    </PageHeader>

    <el-alert v-if="!allowed" type="warning" show-icon :closable="false" title="Web kubectl 拥有集群管理员权限，只允许管理员使用" />

    <!-- 不自动连接: 连接时才会在集群中创建资源 -->
    <el-card v-else-if="!started" class="kc-intro">
      <h3 class="kc-intro__title">连接到集群 {{ clusterStore.currentClusterName || clusterId }}</h3>
      <p class="kc-intro__desc">点击「连接」后，平台会在当前集群中准备 kubectl 环境：</p>
      <ul class="kc-intro__list">
        <li>创建命名空间 <code>buding-kubectl</code> 和绑定 cluster-admin 的服务账户（只在第一次创建）</li>
        <li>为你运行一个 kubectl 容器组，镜像 <code>alpine/k8s</code>，预留 50m CPU / 64Mi 内存，24 小时后自动回收</li>
        <li>镜像拉取失败或无法启动时会自动删除容器组；内网环境请让管理员设置后端环境变量 <code>KUBECTL_IMAGE</code> 指向内网镜像仓库</li>
        <li>所有会话都会记录审计日志</li>
      </ul>
      <el-button type="primary" size="large" @click="start">
        <el-icon><Connection /></el-icon>
        连接
      </el-button>
    </el-card>

    <div v-else class="kc-terminal">
      <TerminalView
        ref="view"
        :build-url="buildUrl"
        height="calc(100vh - 300px)"
        tip="常用命令：kubectl get pods -A、kubectl describe、kubectl logs -f、helm list -A；切换页面不会断开连接，超过 30 分钟没有操作会自动断开"
        @status="(s: TerminalStatus) => (status = s)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, ref, watch } from 'vue'
import { Connection, SwitchButton } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import TerminalView from '@/components/terminal/TerminalView.vue'
import TerminalStatusTag from '@/components/terminal/TerminalStatusTag.vue'
import TerminalToolbar from '@/components/terminal/TerminalToolbar.vue'
import { useClusterStore } from '@/stores/cluster'
import { useUserStore } from '@/stores/user'
import { buildWsUrl } from '@/utils/ws'
import type { TerminalStatus } from '@/utils/terminal'

const clusterStore = useClusterStore()
const userStore = useUserStore()
const allowed = computed(() => userStore.isAdmin)
const clusterId = computed(() => clusterStore.currentClusterId)

const view = ref<InstanceType<typeof TerminalView> | null>(null)
const status = ref<TerminalStatus>('closed')
// 点击「连接」后才显示终端
const started = ref(false)

const buildUrl = () => (clusterId.value ? buildWsUrl('/terminal/kubectl', { clusterId: clusterId.value }) : '')
const connect = () => view.value?.connect(`正在连接集群 ${clusterStore.currentClusterName || clusterId.value} 的 kubectl 环境...`)
const disconnect = () => view.value?.disconnect()

const start = async () => {
  started.value = true
  // 等终端挂载后再连接
  await nextTick()
  connect()
}

// 页面被缓存, 切回来时重新计算终端尺寸
onActivated(() => {
  view.value?.fit()
  view.value?.focus()
})

// 切换集群后断开, 回到连接说明, 由用户决定是否连接新集群
watch(clusterId, (id, old) => {
  if (id !== old && started.value) {
    disconnect()
    started.value = false
    status.value = 'closed'
  }
})
</script>

<style scoped>
.kc-intro {
  max-width: 880px;
}

.kc-intro__title {
  margin: 0 0 10px;
  color: #1e293b;
  font-size: 18px;
}

.kc-intro__desc {
  margin: 0 0 8px;
  color: #475569;
}

.kc-intro__list {
  margin: 0 0 20px;
  padding-left: 20px;
  color: #475569;
  line-height: 2;
}

.kc-intro code {
  padding: 0 4px;
  border-radius: 4px;
  background: rgba(59, 116, 242, 0.1);
}

.kc-terminal {
  overflow: hidden;
  border: 1px solid #1e293b;
  border-radius: 12px;
  background: #0f172a;
}
</style>
