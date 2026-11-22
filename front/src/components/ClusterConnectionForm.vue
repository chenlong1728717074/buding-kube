<template>
  <el-form-item label="连接方式">
    <el-radio-group v-model="authType">
      <el-radio-button value="config">KubeConfig</el-radio-button>
      <el-radio-button value="token">URI + Token</el-radio-button>
    </el-radio-group>
  </el-form-item>
  <el-form-item v-if="authType === 'config'" label="Kubeconfig 配置" prop="config">
    <el-input
      v-model="config"
      type="textarea"
      :rows="18"
      :placeholder="configPlaceholder"
      style="font-family: 'SF Mono', Monaco, Consolas, monospace; font-size: 13px;"
    />
    <div v-if="!requiredConfig" style="margin-top: 8px; font-size: 12px; color: #6b7280;">
      提示：留空则不更新配置，仅更新别名和描述
    </div>
  </el-form-item>
  <template v-else>
    <el-row :gutter="20">
      <el-col :span="12">
        <el-form-item label="API Server URI" prop="uri">
          <el-input v-model="uri" placeholder="例如：https://10.0.0.1:6443" />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="Bearer Token" prop="token">
          <el-input v-model="token" type="password" show-password placeholder="请输入 token" />
        </el-form-item>
      </el-col>
    </el-row>
    <el-form-item label="CA 证书">
      <el-input
        v-model="caData"
        type="textarea"
        :rows="5"
        :disabled="insecure"
        placeholder="API Server 的 CA 证书（PEM 格式或 kubeconfig 中 certificate-authority-data 的 base64 值）。留空则使用系统证书校验"
        style="font-family: 'SF Mono', Monaco, Consolas, monospace; font-size: 12px;"
      />
    </el-form-item>
    <el-form-item label="证书校验">
      <el-checkbox v-model="insecure">跳过证书校验（不安全，仅用于测试环境）</el-checkbox>
      <div v-if="insecure" class="ccf-warn">跳过校验后无法确认连接的是真正的 API Server，Token 可能被中间人截获</div>
    </el-form-item>
  </template>
</template>

<script setup lang="ts">
// 集群连接方式表单(KubeConfig / URI+Token),供添加和编辑对话框复用
// 表单项使用 prop,校验规则由父级 el-form 的 :rules 提供;
// 隐藏(v-if)的字段不会参与校验,因此 authType 切换后只校验当前可见项
withDefaults(
  defineProps<{
    /** 编辑场景 kubeconfig 可留空,展示不同提示 */
    requiredConfig?: boolean
    configPlaceholder?: string
  }>(),
  {
    requiredConfig: true,
    configPlaceholder: '请粘贴完整的 kubeconfig 配置文件内容'
  }
)

const authType = defineModel<string>('authType', { default: 'config' })
const config = defineModel<string>('config', { default: '' })
const uri = defineModel<string>('uri', { default: '' })
const token = defineModel<string>('token', { default: '' })
const caData = defineModel<string>('caData', { default: '' })
const insecure = defineModel<boolean>('insecure', { default: false })
</script>

<style scoped>
.ccf-warn {
  width: 100%;
  margin-top: 4px;
  color: #d97706;
  font-size: 12px;
  line-height: 1.5;
}
</style>
