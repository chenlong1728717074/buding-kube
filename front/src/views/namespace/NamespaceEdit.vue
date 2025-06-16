<template>
  <div class="namespace-edit">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack" type="text" size="large">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="header-info">
          <h1>编辑命名空间</h1>
          <p class="header-description">修改命名空间的基本信息</p>
        </div>
      </div>
      <div class="header-actions">
        <el-button @click="handleViewDetail">查看详情</el-button>
      </div>
    </div>

    <el-card class="form-card">
      <el-form
        ref="formRef"
        :model="namespaceForm"
        :rules="formRules"
        label-width="120px"
        class="namespace-form"
        v-loading="loading"
      >
        <el-form-item label="集群">
          <el-input v-model="clusterName" disabled />
        </el-form-item>
        
        <el-form-item label="命名空间">
          <el-input v-model="namespaceForm.namespace" disabled />
        </el-form-item>
        
        <el-form-item label="别名" prop="alias">
          <el-input 
            v-model="namespaceForm.alias" 
            placeholder="请输入命名空间别名"
            maxlength="50"
            show-word-limit
          />
        </el-form-item>
        
        <el-form-item label="描述" prop="describe">
          <el-input 
            v-model="namespaceForm.describe" 
            type="textarea"
            :rows="4"
            placeholder="请输入命名空间描述"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
        
        <div class="form-actions">
          <el-button @click="handleBack">取消</el-button>
          <el-button 
            type="primary" 
            @click="handleSubmit"
            :loading="submitLoading"
          >
            保存
          </el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { namespaceApi } from '@/api/namespace'
import { clusterApi } from '@/api/cluster'
import type { NamespaceVO, NamespaceBaseDTO, UpdateNamespaceDTO } from '@/api/namespace'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const submitLoading = ref(false)
const formRef = ref()
const namespaceInfo = ref<NamespaceVO | null>(null)
const clusterName = ref('')

// 表单数据
const namespaceForm = reactive({
  clusterId: '',
  namespace: '',
  alias: '',
  describe: ''
})

// 表单验证规则
const formRules = {
  alias: [
    { max: 50, message: '别名长度不能超过50个字符', trigger: 'blur' }
  ],
  describe: [
    { max: 200, message: '描述长度不能超过200个字符', trigger: 'blur' }
  ]
}

// 获取命名空间详情
const fetchNamespaceDetail = async () => {
  try {
    loading.value = true
    const params: NamespaceBaseDTO = {
      clusterId: route.query.clusterId as string,
      namespace: route.query.namespace as string
    }
    
    const response = await namespaceApi.getInfo(params)
    console.log('命名空间详情API响应:', response)
    
    if (response.code === 200 && response.data) {
      namespaceInfo.value = response.data
      
      // 填充表单
      Object.assign(namespaceForm, {
        clusterId: route.query.clusterId as string,
        namespace: response.data.name,
        alias: response.data.alias || '',
        describe: response.data.describe || ''
      })
    }
    
    // 获取集群名称
    await fetchClusterName(route.query.clusterId as string)
  } catch (error) {
    console.error('获取命名空间详情失败:', error)
    ElMessage.error('获取命名空间详情失败')
  } finally {
    loading.value = false
  }
}

// 获取集群名称
const fetchClusterName = async (clusterId: string) => {
  try {
    const response = await clusterApi.getClusters()
    console.log('集群列表API响应:', response)
    
    if (response.code === 200 && response.data && response.data.items) {
      const cluster = response.data.items.find((item: any) => item.id === clusterId)
      if (cluster) {
        clusterName.value = cluster.name
      }
    }
  } catch (error) {
    console.error('获取集群信息失败:', error)
  }
}

// 返回
const handleBack = () => {
  router.back()
}

// 查看详情
const handleViewDetail = () => {
  router.push({
    path: '/namespace/detail',
    query: {
      clusterId: route.query.clusterId,
      namespace: route.query.namespace
    }
  })
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
    
    await ElMessageBox.confirm(
      '确定要保存修改吗？',
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    submitLoading.value = true
    
    const updateData: UpdateNamespaceDTO = {
      clusterId: namespaceForm.clusterId,
      namespace: namespaceForm.namespace,
      alias: namespaceForm.alias,
      describe: namespaceForm.describe
    }
    
    await namespaceApi.update(updateData)
    
    ElMessage.success('修改成功')
    handleBack()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('修改命名空间失败:', error)
      ElMessage.error(error.message || '修改失败')
    }
  } finally {
    submitLoading.value = false
  }
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
.namespace-edit {
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

.form-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.namespace-form {
  padding: 20px;
  max-width: 800px;
}

.form-actions {
  display: flex;
  gap: 12px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

:deep(.el-card__body) {
  padding: 0;
}

:deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

:deep(.el-input.is-disabled .el-input__inner) {
  background-color: #f5f7fa;
  border-color: #e4e7ed;
  color: #c0c4cc;
}

:deep(.el-button + .el-button) {
  margin-left: 0;
}
</style>