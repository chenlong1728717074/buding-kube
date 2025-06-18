<template>
  <div class="node-list">
    <div class="page-header">
      <h1>节点列表</h1>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="集群">
          <el-select 
            v-model="searchForm.clusterId" 
            placeholder="请选择集群"
            style="width: 200px"
            @change="handleClusterChange"
          >
            <el-option
              v-for="cluster in clusterList"
              :key="cluster.id || cluster.name"
              :label="cluster.name"
              :value="cluster.id || cluster.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="节点名称">
          <el-input 
            v-model="searchForm.keyword" 
            placeholder="请输入节点名称" 
            style="width: 200px"
            clearable
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select 
            v-model="searchForm.status" 
            placeholder="请选择状态"
            style="width: 150px"
            clearable
          >
            <el-option label="Ready" value="Ready" />
            <el-option label="NotReady" value="NotReady" />
            <el-option label="Unknown" value="Unknown" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">
            搜索
          </el-button>
          <el-button :icon="Refresh" @click="handleRefresh">
            刷新
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 节点列表 -->
    <el-card class="table-card">
      <el-table 
        v-loading="loading" 
        :data="nodeList" 
        stripe
      >
        <el-table-column prop="hostname" label="主机名" min-width="200" show-overflow-tooltip />
        <el-table-column prop="role" label="角色" width="120" align="center">
          <template #default="{ row }">
            <el-tag 
              :type="row.role === 'master' ? 'danger' : 'info'"
              size="small"
            >
              {{ row.role }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120" align="center">
          <template #default="{ row }">
            <el-tag 
              :type="getStatusType(row.status)"
              size="small"
            >
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP地址" min-width="180" show-overflow-tooltip />
        <el-table-column prop="cpu" label="CPU" min-width="150" align="center" />
        <el-table-column prop="memory" label="内存" min-width="180" show-overflow-tooltip />
        <el-table-column prop="unSchedule" label="调度状态" min-width="200" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.unSchedule"
              :active-text="row.unSchedule ? '停止调度' : '启用调度'"
              :inactive-text="row.unSchedule ? '停止调度' : '启用调度'"
              @change="handleScheduleChange(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchNodeList"
          @current-change="fetchNodeList"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { clusterApi, type ClusterVO } from '@/api/cluster'

const route = useRoute()
const router = useRouter()

// 搜索表单
const searchForm = reactive({
  clusterId: '',
  keyword: '',
  status: ''
})

// 集群列表
const clusterList = ref<ClusterVO[]>([])

// 节点列表
const nodeList = ref<any[]>([])
const loading = ref(false)

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 获取集群列表
const fetchClusterList = async () => {
  try {
    const response = await clusterApi.getClusters({
      page: 1,
      pageSize: 100 // 获取所有集群
    })
    
    if (response.code === 200 && response.data) {
      clusterList.value = response.data.items || []
      
      // 如果没有选中集群且有集群数据，选择第一个
      if (!searchForm.clusterId && clusterList.value.length > 0) {
        searchForm.clusterId = clusterList.value[0].id || clusterList.value[0].name
      }
    }
  } catch (error) {
    console.error('获取集群列表失败:', error)
    ElMessage.error('获取集群列表失败')
  }
}

// 获取节点列表
const fetchNodeList = async () => {
  if (!searchForm.clusterId) {
    ElMessage.warning('请先选择集群')
    return
  }
  
  loading.value = true
  try {
    const response = await clusterApi.getNodes(searchForm.clusterId, {
      ...searchForm,
      page: pagination.page,
      pageSize: pagination.size
    })
    
    if (response.code === 200 && response.data) {
      nodeList.value = response.data.items || []
      pagination.total = response.data.total || 0
    } else {
      nodeList.value = []
      pagination.total = 0
      ElMessage.warning(response.msg || '暂无节点数据')
    }
  } catch (error) {
    console.error('加载节点列表失败:', error)
    nodeList.value = []
    pagination.total = 0
    ElMessage.error('加载节点列表失败')
  } finally {
    loading.value = false
  }
}

// 集群变更
const handleClusterChange = () => {
  pagination.page = 1
  fetchNodeList()
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchNodeList()
}

// 刷新
const handleRefresh = () => {
  fetchNodeList()
}

// 查看详情
const handleViewDetail = (row: any) => {
  router.push({
    path: '/node/detail',
    query: {
      clusterId: searchForm.clusterId,
      hostname: row.hostname || row.name
    }
  })
}

// 处理调度状态切换
const handleScheduleChange = async (row: any) => {
  try {
    const action = row.unSchedule ? '停止调度' : '启用调度'
    await clusterApi.toggleNodeSchedule(searchForm.clusterId, row.hostname, row.unSchedule)
    ElMessage.success(`${action}操作成功`)
  } catch (error) {
    console.error('调度状态切换失败:', error)
    ElMessage.error('调度状态切换失败')
    // 恢复原状态
    row.unSchedule = !row.unSchedule
  }
}

// 获取状态类型
const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    'Ready': 'success',
    'NotReady': 'danger',
    'Unknown': 'warning'
  }
  return statusMap[status] || 'info'
}

// 初始化
onMounted(async () => {
  // 检查是否从集群列表跳转过来
  const clusterId = route.query.clusterId as string
  if (clusterId) {
    searchForm.clusterId = clusterId
  }
  
  await fetchClusterList()
  
  // 如果有集群ID，直接获取节点列表
  if (searchForm.clusterId) {
    fetchNodeList()
  }
})
</script>

<style scoped>
.node-list {
  padding: 20px;
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

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.search-card {
  margin-bottom: 20px;
}

.table-card {
  min-height: 400px;
}

.pagination-wrapper {
  margin-top: 20px;
  text-align: right;
}
</style>