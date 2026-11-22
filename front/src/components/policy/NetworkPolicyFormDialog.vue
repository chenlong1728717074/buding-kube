<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="1060px"
    top="5vh"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--wide"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="reset"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">{{ title }}</h3>
        <div class="dialog-actions">
          <el-switch v-model="yamlMode" active-text="YAML 预览" :disabled="!!unsupported" />
        </div>
      </div>
    </template>
    <div v-loading="loading" class="config-editor">
      <div class="config-content">
        <div v-if="yamlMode" class="yaml-view">
          <YamlEditor :model-value="previewYaml" :readonly="true" height="100%" title="NetworkPolicy YAML 预览" :filename="`${name || 'networkpolicy'}.yaml`" />
        </div>
        <div v-else class="basic-form">
          <el-alert v-if="unsupported" type="warning" :closable="false" show-icon class="np-alert">
            这条网络策略{{ unsupported }}，表单无法完整表示，请使用「编辑YAML」修改。
          </el-alert>
          <template v-else>
            <NetworkPolicySupportAlert :cluster-id="clusterId" />
            <p class="np-tip">
              命名空间中只要有一条策略选中了某个容器组，该容器组就只允许策略中放行的流量。
            </p>
            <template v-if="!isEdit">
              <p class="fd-subtitle">常用模板</p>
              <div class="fd-choice-grid np-presets">
                <button v-for="p in NETWORK_POLICY_PRESETS" :key="p.key" type="button" class="fd-choice" @click="applyPreset(p.key)">
                  <span class="fd-choice__body">
                    <span class="fd-choice__title"><span>{{ p.label }}</span></span>
                    <span class="fd-choice__desc">{{ p.desc }}</span>
                  </span>
                </button>
              </div>
            </template>
            <el-form label-width="110px">
              <div class="np-row">
                <el-form-item label="命名空间" required>
                  <InfiniteSelect v-model="namespace" :fetch-data="namespaceFetcher" v-bind="namespaceSelectConfig" :clearable="false" :disabled="isEdit" placeholder="请选择命名空间" style="width: 100%" />
                </el-form-item>
                <el-form-item label="名称" required>
                  <el-input v-model="name" :disabled="isEdit" placeholder="例如 allow-frontend" maxlength="253" />
                </el-form-item>
              </div>
              <el-form-item label="应用于">
                <el-radio-group v-model="form.podScope">
                  <el-radio-button value="all">命名空间内所有容器组</el-radio-button>
                  <el-radio-button value="labels">指定标签的容器组</el-radio-button>
                </el-radio-group>
              </el-form-item>
            </el-form>
            <KVRowsTable v-if="form.podScope === 'labels'" v-model="form.podLabels" title="容器组标签" add-text="添加标签" empty-text="请添加标签，例如 app = web" />

            <div v-for="dir in DIRECTIONS" :key="dir.key" class="np-section">
              <div class="np-section__head">
                <el-switch v-model="form[dir.enabledKey]" />
                <span class="np-section__title">{{ dir.title }}</span>
                <span class="np-section__desc">{{ form[dir.enabledKey] ? (form[dir.key].length ? dir.hintRules : dir.hintEmpty) : dir.hintOff }}</span>
              </div>
              <template v-if="form[dir.enabledKey]">
                <div v-for="(rule, ri) in form[dir.key]" :key="ri" class="np-rule">
                  <div class="np-rule__head">
                    <span>规则 {{ ri + 1 }}</span>
                    <el-checkbox v-model="rule.any" @change="(v: unknown) => !v && !rule.peers.length && rule.peers.push(createPeer())">{{ dir.anyText }}</el-checkbox>
                    <el-button link type="danger" class="np-del" @click="form[dir.key].splice(ri, 1)"><el-icon><Delete /></el-icon>删除规则</el-button>
                  </div>
                  <template v-if="!rule.any">
                    <div v-for="(peer, pi) in rule.peers" :key="pi" class="np-peer">
                      <el-select v-model="peer.type" style="width: 230px">
                        <el-option v-for="t in PEER_TYPES" :key="t.value" :label="t.label" :value="t.value" />
                      </el-select>
                      <div class="np-peer__body">
                        <KVRowsTable v-if="peer.type === 'pods'" v-model="peer.labels" title="容器组标签" add-text="添加标签" empty-text="请添加标签" />
                        <InfiniteSelect v-else-if="peer.type === 'namespaces'" v-model="peer.namespaces" :fetch-data="namespaceFetcher" v-bind="namespaceSelectConfig" multiple placeholder="选择命名空间" style="width: 100%" />
                        <div v-else-if="peer.type === 'ipBlock'" class="np-ip">
                          <el-input v-model="peer.cidr" placeholder="例如 10.0.0.0/8" class="mono-input" />
                          <el-input v-model="peer.except" placeholder="排除的网段，多个用逗号分隔（可选）" class="mono-input" />
                        </div>
                        <span v-else class="np-muted">同一命名空间中的所有容器组</span>
                      </div>
                      <el-button link type="danger" @click="rule.peers.splice(pi, 1)"><el-icon><Delete /></el-icon></el-button>
                    </div>
                    <el-button size="small" @click="rule.peers.push(createPeer())"><el-icon><Plus /></el-icon>{{ dir.addPeerText }}</el-button>
                  </template>
                  <div class="np-ports">
                    <span class="np-ports__label">端口</span>
                    <div v-for="(port, pi) in rule.ports" :key="pi" class="np-port">
                      <el-select v-model="port.protocol" style="width: 100px">
                        <el-option label="TCP" value="TCP" />
                        <el-option label="UDP" value="UDP" />
                        <el-option label="SCTP" value="SCTP" />
                      </el-select>
                      <el-input v-model="port.port" placeholder="端口号或名称，留空为全部" style="width: 220px" />
                      <el-button link type="danger" @click="rule.ports.splice(pi, 1)"><el-icon><Delete /></el-icon></el-button>
                    </div>
                    <el-button size="small" @click="rule.ports.push({ protocol: 'TCP', port: '' })"><el-icon><Plus /></el-icon>添加端口</el-button>
                    <span v-if="!rule.ports.length" class="np-muted">不限制端口</span>
                  </div>
                </div>
                <el-button class="np-add" @click="form[dir.key].push(createRule())"><el-icon><Plus /></el-icon>添加{{ dir.short }}规则</el-button>
              </template>
            </div>
            <el-alert v-if="form.egressEnabled && !egressAllowsDns" type="info" :closable="false" show-icon class="np-alert">
              限制出站后容器组可能无法解析域名，通常需要放行 UDP / TCP 53 端口。
              <el-button link type="primary" @click="addDnsRule">添加 DNS 规则</el-button>
            </el-alert>
          </template>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="saving" :disabled="loading || !!unsupported" @click="save">{{ isEdit ? '保存' : '创建' }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import YamlEditor from '@/components/YamlEditor.vue'
import KVRowsTable from '@/components/KVRowsTable.vue'
import NetworkPolicySupportAlert from '@/components/policy/NetworkPolicySupportAlert.vue'
import { editableCopy, resourceApi, saveResourceObject, type KubeObject } from '@/api/resource'
import { jsonToYaml } from '@/utils/yaml'
import {
  NETWORK_POLICY_PRESETS,
  PEER_TYPES,
  buildNetworkPolicySpec,
  createNetworkPolicyForm,
  createPeer,
  createRule,
  parseNetworkPolicySpec,
  type NetworkPolicyForm,
  type RuleForm
} from '@/utils/networkPolicy'

const props = withDefaults(
  defineProps<{ modelValue: boolean; clusterId: string; namespace?: string; editName?: string }>(),
  { namespace: '', editName: '' }
)
// 命名空间下拉框分页加载
const namespaceFetcher = computed(() => useNamespaceFetcher(props.clusterId))
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; success: [] }>()

const NP_REF = { kind: 'networkpolicy' as const }
const DIRECTIONS = [
  {
    key: 'ingress' as const,
    enabledKey: 'ingressEnabled' as const,
    title: '限制入站流量',
    short: '入站',
    anyText: '任意来源',
    addPeerText: '添加来源',
    hintOff: '不限制入站流量',
    hintEmpty: '没有规则，拒绝所有入站流量',
    hintRules: '只放行满足任意一条规则的入站流量'
  },
  {
    key: 'egress' as const,
    enabledKey: 'egressEnabled' as const,
    title: '限制出站流量',
    short: '出站',
    anyText: '任意目标',
    addPeerText: '添加目标',
    hintOff: '不限制出站流量',
    hintEmpty: '没有规则，拒绝所有出站流量（包括 DNS）',
    hintRules: '只放行满足任意一条规则的出站流量'
  }
]

const isEdit = computed(() => !!props.editName)
const title = computed(() => (isEdit.value ? '编辑网络策略' : '创建网络策略'))
const loading = ref(false)
const saving = ref(false)
const yamlMode = ref(false)
const namespace = ref('')
const name = ref('')
const original = ref<KubeObject | null>(null)
const unsupported = ref('')
const form = reactive<NetworkPolicyForm>(createNetworkPolicyForm())

const egressAllowsDns = computed(() =>
  form.egress.some(r => r.ports.some(p => p.port.trim() === '53') || (!r.ports.length && r.any))
)

const applyPreset = (key: string) => {
  const preset = NETWORK_POLICY_PRESETS.find(p => p.key === key)
  if (preset) Object.assign(form, preset.apply())
  if (!name.value) name.value = key
}

const addDnsRule = () => {
  form.egress.push({ any: true, peers: [], ports: [{ protocol: 'UDP', port: '53' }, { protocol: 'TCP', port: '53' }] } as RuleForm)
}

const reset = async () => {
  Object.assign(form, createNetworkPolicyForm())
  original.value = null
  unsupported.value = ''
  yamlMode.value = false
  namespace.value = props.namespace
  name.value = props.editName
  if (!isEdit.value) return
  loading.value = true
  try {
    const { data } = await resourceApi.object({ ...NP_REF, clusterId: props.clusterId, namespace: namespace.value, name: props.editName })
    original.value = data
    const parsed = parseNetworkPolicySpec(data.spec)
    if ('reason' in parsed) unsupported.value = parsed.reason
    else Object.assign(form, parsed.form)
  } catch {
    emit('update:modelValue', false)
  } finally {
    loading.value = false
  }
}

const buildObject = (): KubeObject => {
  const base: KubeObject = original.value
    ? editableCopy(original.value)
    : { apiVersion: 'networking.k8s.io/v1', kind: 'NetworkPolicy', metadata: { name: name.value.trim(), namespace: namespace.value } }
  base.spec = buildNetworkPolicySpec(form)
  return base
}

const previewYaml = computed(() => jsonToYaml(JSON.parse(JSON.stringify(buildObject()))))

const CIDR = /^[0-9a-fA-F:.]+\/\d{1,3}$/

const validate = (): string => {
  if (!namespace.value) return '请选择命名空间'
  if (!/^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$/.test(name.value.trim())) return '名称只能包含小写字母、数字、"-" 和 "."'
  if (!form.ingressEnabled && !form.egressEnabled) return '请至少限制入站或出站流量'
  if (form.podScope === 'labels' && !form.podLabels.some(r => r.key.trim())) return '请填写容器组标签'
  for (const dir of DIRECTIONS) {
    if (!form[dir.enabledKey]) continue
    for (const [i, rule] of form[dir.key].entries()) {
      const label = `${dir.short}规则 ${i + 1}`
      if (!rule.any && !rule.peers.length) return `${label}请添加${dir.short === '入站' ? '来源' : '目标'}，或勾选「${dir.anyText}」`
      for (const p of rule.any ? [] : rule.peers) {
        if (p.type === 'pods' && !p.labels.some(r => r.key.trim())) return `${label}的容器组标签不能为空`
        if (p.type === 'namespaces' && !p.namespaces.length) return `${label}请选择命名空间`
        if (p.type === 'ipBlock' && !CIDR.test(p.cidr.trim())) return `${label}的 IP 段格式不正确，例如 10.0.0.0/8`
      }
      for (const port of rule.ports) {
        const v = port.port.trim()
        if (v && !/^\d+$/.test(v) && !/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/.test(v)) return `${label}的端口 ${v} 不正确`
        if (/^\d+$/.test(v) && (Number(v) < 1 || Number(v) > 65535)) return `${label}的端口 ${v} 不正确`
      }
    }
  }
  return ''
}

const save = async () => {
  const message = validate()
  if (message) {
    ElMessage.warning(message)
    return
  }
  saving.value = true
  try {
    await saveResourceObject(props.clusterId, NP_REF, buildObject(), !isEdit.value)
    ElMessage.success(isEdit.value ? '保存成功' : '网络策略已创建')
    emit('update:modelValue', false)
    emit('success')
  } catch {
    // 请求拦截器已提示错误
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.np-tip {
  margin: 0 0 14px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.6;
}

.np-alert {
  margin-bottom: 14px;
}

.np-presets {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.np-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.np-section {
  margin-top: 8px;
  padding: 14px 16px;
  border: 1px solid #e3eaf4;
  border-radius: 12px;
  background: #fbfcfe;
}

.np-section + .np-section {
  margin-top: 14px;
}

.np-section__head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.np-section__title {
  color: #1e293b;
  font-weight: 600;
}

.np-section__desc {
  color: #64748b;
  font-size: 12px;
}

.np-rule {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 10px;
  padding: 12px 14px;
  border: 1px solid #e6ecf5;
  border-radius: 10px;
  background: #fff;
}

.np-rule__head {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  color: #334155;
  font-weight: 600;
}

.np-del {
  margin-left: auto;
}

.np-peer {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  width: 100%;
}

.np-peer__body {
  flex: 1;
  min-width: 0;
}

.np-ip {
  display: grid;
  grid-template-columns: 1fr 1.4fr;
  gap: 8px;
}

.np-ports {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.np-ports__label {
  color: #64748b;
  font-size: 13px;
}

.np-port {
  display: flex;
  align-items: center;
  gap: 6px;
}

.np-muted {
  color: #94a3b8;
  font-size: 12px;
}

.np-add {
  width: 100%;
  border-style: dashed;
}

.mono-input :deep(input) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
