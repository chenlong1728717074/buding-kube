<template>
  <div class="namespace-detail">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack" circle>
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="header-info">
          <h1>{{ namespaceInfo?.name || '命名空间详情' }}</h1>
          <p class="header-description">查看命名空间的详细信息</p>
        </div>
      </div>
      <div class="header-actions">
        <el-button @click="handleEdit">
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
        <el-button @click="handleViewYaml">
          <el-icon><Document /></el-icon>
          查看YAML
        </el-button>
        <el-button type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </div>
    </div>

    <div class="detail-content" v-loading="loading">
      <el-row :gutter="20">
        <!-- 基本信息 -->
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>基本信息</span>
              </div>
            </template>
            <div class="info-list">
              <div class="info-item">
                <label>命名空间名称：</label>
                <span>{{ namespaceInfo?.name || '-' }}</span>
              </div>
              <div class="info-item">
                <label>别名：</label>
                <span>{{ namespaceInfo?.alias || '-' }}</span>
              </div>
              <div class="info-item">
                <label>描述：</label>
                <span>{{ namespaceInfo?.describe || '-' }}</span>
              </div>
              <div class="info-item">
                <label>状态：</label>
                <el-tag 
                  :type="getStatusType(namespaceInfo?.active)"
                  size="small"
                >
                  {{ getStatusText(namespaceInfo?.active) }}
                </el-tag>
              </div>
              <div class="info-item">
                <label>UID：</label>
                <span class="uid-text">{{ namespaceInfo?.uid || '-' }}</span>
              </div>
              <div class="info-item">
                <label>资源版本：</label>
                <span>{{ namespaceInfo?.resourceVersion || '-' }}</span>
              </div>
              <div class="info-item">
                <label>API版本：</label>
                <span>{{ namespaceInfo?.version || '-' }}</span>
              </div>
              <div class="info-item">
                <label>创建时间：</label>
                <span>{{ formatDate(namespaceInfo?.creationTimestamp) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>

        <!-- 状态信息 -->
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>状态信息</span>
              </div>
            </template>
            <div class="info-list">
              <div class="info-item">
                <label>阶段：</label>
                <el-tag 
                  :type="getStatusType(namespaceInfo?.active)"
                  size="small"
                >
                  {{ getStatusText(namespaceInfo?.active) }}
                </el-tag>
              </div>
              <div class="info-item">
                <label>Generation：</label>
                <span>{{ namespaceInfo?.generation || '-' }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px;">
        <!-- Labels -->
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>Labels</span>
                <el-tag size="small" type="info">{{ Object.keys(namespaceInfo?.labels || {}).length }}</el-tag>
              </div>
            </template>
            <div class="labels-content">
              <div v-if="!namespaceInfo?.labels || Object.keys(namespaceInfo.labels).length === 0" class="empty-state">
                <el-empty description="暂无Labels" :image-size="60" />
              </div>
              <div v-else class="labels-list">
                <div 
                  v-for="(value, key) in namespaceInfo.labels" 
                  :key="key" 
                  class="label-item"
                >
                  <div class="label-key">{{ key }}</div>
                  <div class="label-value">{{ value }}</div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>

        <!-- Annotations -->
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>Annotations</span>
                <el-tag size="small" type="info">{{ Object.keys(namespaceInfo?.annotations || {}).length }}</el-tag>
              </div>
            </template>
            <div class="annotations-content">
              <div v-if="!namespaceInfo?.annotations || Object.keys(namespaceInfo.annotations).length === 0" class="empty-state">
                <el-empty description="暂无Annotations" :image-size="60" />
              </div>
              <div v-else class="annotations-list">
                <div 
                  v-for="(value, key) in namespaceInfo.annotations" 
                  :key="key" 
                  class="annotation-item"
                >
                  <div class="annotation-key">{{ key }}</div>
                  <div class="annotation-value">{{ value }}</div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- Spec信息 -->
      <el-row style="margin-top: 20px;">
        <el-col :span="24">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>Spec配置</span>
              </div>
            </template>
            <div class="spec-content">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="Finalizers">
                  <div v-if="namespaceInfo?.spec?.finalizers && namespaceInfo.spec.finalizers.length > 0">
                    <el-tag 
                      v-for="finalizer in namespaceInfo.spec.finalizers" 
                      :key="finalizer" 
                      size="small" 
                      style="margin-right: 8px; margin-bottom: 4px;"
                    >
                      {{ finalizer }}
                    </el-tag>
                  </div>
                  <span v-else>-</span>
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- YAML查看对话框 -->
    <el-dialog 
      v-model="yamlDialogVisible" 
      title="YAML配置" 
      width="80%"
      destroy-on-close
    >
      <el-input 
        v-model="yamlContent" 
        type="textarea" 
        :rows="25" 
        readonly
        style="font-family: 'Courier New', monospace;"
      />
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlDialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="handleCopyYaml">
            <el-icon><CopyDocument /></el-icon>
            复制
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  ArrowLeft,
  Edit,
  Document,
  Delete,
  CopyDocument
} from '@element-plus/icons-vue'
import { 
  namespaceApi, 
  type NamespaceVO,
  type NamespaceBaseDTO
} from '@/api/namespace'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const namespaceInfo = ref<NamespaceVO | null>(null)
const yamlDialogVisible = ref(false)
const yamlContent = ref('')

// 获取命名空间详情
const fetchNamespaceDetail = async () => {
  try {
    loading.value = true
    const params: NamespaceBaseDTO = {
      clusterId: route.query.clusterId as string,
      namespace: route.query.namespace as string
    }
    
    const response = await namespaceApi.getInfo(params)
    namespaceInfo.value = response.data
  } catch (error: any) {
    ElMessage.error('获取命名空间详情失败')
  } finally {
    loading.value = false
  }
}

// 返回列表
const handleBack = () => {
  router.back()
}

// 编辑
const handleEdit = () => {
  // 跳转到编辑页面或打开编辑对话框
  ElMessage.info('编辑功能待实现')
}

// 查看YAML
const handleViewYaml = () => {
  if (namespaceInfo.value?.yaml) {
    yamlContent.value = namespaceInfo.value.yaml
    yamlDialogVisible.value = true
  } else {
    ElMessage.warning('暂无YAML配置')
  }
}

// 复制YAML
const handleCopyYaml = async () => {
  try {
    await navigator.clipboard.writeText(yamlContent.value)
    ElMessage.success('复制成功')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 删除
const handleDelete = () => {
  if (!namespaceInfo.value) return
  
  ElMessageBox.confirm(
    `确定要删除命名空间 "${namespaceInfo.value.name}" 吗？此操作不可恢复！`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      const params: NamespaceBaseDTO = {
        clusterId: route.query.clusterId as string,
        namespace: namespaceInfo.value!.name
      }
      await namespaceApi.delete(params)
      ElMessage.success('删除成功')
      router.back()
    } catch (error: any) {
      ElMessage.error('删除失败')
    }
  })
}

// 状态相关
const getStatusType = (status?: string) => {
  switch (status) {
    case 'Active':
      return 'success'
    case 'Terminating':
      return 'warning'
    default:
      return 'danger'
  }
}

const getStatusText = (status?: string) => {
  switch (status) {
    case 'Active':
      return '活跃'
    case 'Terminating':
      return '终止中'
    default:
      return status || '-'
  }
}

// 格式化日期
const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 页面加载时获取详情
onMounted(() => {
  if (route.query.clusterId && route.query.namespace) {
    fetchNamespaceDetail()
  } else {
    ElMessage.error('缺少必要参数')
    router.back()
  }
})
</script>

<style scoped>
.namespace-detail {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-info h1 {
  margin: 0 0 4px 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.header-description {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.detail-content {
  min-height: 400px;
}

.info-card {
  margin-bottom: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  color: #2c3e50;
}

.info-list {
  padding: 0;
}

.info-item {
  display: flex;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.info-item:last-child {
  border-bottom: none;
}

.info-item label {
  min-width: 120px;
  font-weight: 500;
  color: #606266;
  margin-right: 12px;
}

.info-item span {
  color: #303133;
  word-break: break-all;
}

.uid-text {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  background-color: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
}

.labels-content,
.annotations-content {
  max-height: 300px;
  overflow-y: auto;
}

.empty-state {
  text-align: center;
  padding: 20px;
}

.labels-list,
.annotations-list {
  padding: 0;
}

.label-item,
.annotation-item {
  display: flex;
  flex-direction: column;
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  margin-bottom: 8px;
  background-color: #fafafa;
}

.label-key,
.annotation-key {
  font-weight: 500;
  color: #409eff;
  font-size: 12px;
  margin-bottom: 4px;
  word-break: break-all;
}

.label-value,
.annotation-value {
  color: #606266;
  font-size: 13px;
  word-break: break-all;
  font-family: 'Courier New', monospace;
}

.spec-content {
  padding: 0;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-card__header) {
  background-color: #f8f9fa;
  border-bottom: 1px solid #ebeef5;
}

:deep(.el-card__body) {
  padding: 20px;
}

:deep(.el-descriptions__label) {
  font-weight: 500;
}

:deep(.el-button + .el-button) {
  margin-left: 8px;
}

:deep(.el-empty) {
  padding: 20px 0;
}
</style>