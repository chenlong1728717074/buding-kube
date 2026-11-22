// 列表表格的表头图标与状态列样式, 配合 assets/styles/list-page.css 使用
// 用法: <el-table :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">

interface CellClassParams {
  column: { label?: string; type?: string }
}

// 按顺序匹配列名, 命中第一条规则即使用对应图标
const HEADER_ICON_RULES: Array<[RegExp, string]> = [
  [/操作/, 'settings'],
  [/名称|主机名|用户名/, 'box'],
  [/数量|个数/, 'hash'],
  [/重启|次数/, 'recycle'],
  [/镜像/, 'image'],
  [/变量/, 'key'],
  [/绑定/, 'link'],
  [/时间|日期|时长/, 'clock'],
  [/调度/, 'calendar'],
  [/状态|阶段|就绪/, 'activity'],
  [/命名空间/, 'layers'],
  [/别名|标签/, 'tag'],
  [/描述|备注|原因|消息/, 'message'],
  [/IP|地址|域名|端口/i, 'globe'],
  [/节点/, 'server'],
  [/CPU/i, 'cpu'],
  [/内存/, 'memory'],
  [/副本/, 'copy'],
  [/版本/, 'hash'],
  [/作用域/, 'target'],
  [/容量|存储|Volume|StorageClass|持久卷/i, 'database'],
  [/访问模式|权限|角色/, 'shield'],
  [/回收|策略/, 'recycle'],
  [/Token|密钥|Secret/i, 'key'],
  [/类型|Kind|类$/i, 'shapes'],
  [/组|供应者/, 'folder'],
  [/邮箱/, 'mail'],
  [/手机|电话/, 'phone'],
  [/部门/, 'building'],
  [/姓名|用户/, 'user'],
  [/克隆|扩展|允许|启用|自动/, 'toggle']
]

export const tableHeaderClass = ({ column }: CellClassParams): string => {
  const label = (column.label || '').trim()
  if (!label || column.type === 'selection' || column.type === 'index') return ''
  const rule = HEADER_ICON_RULES.find(([pattern]) => pattern.test(label))
  return `th-ico th-ico--${rule ? rule[1] : 'list'}`
}

// 状态列的标签显示为带圆点的胶囊
export const tableCellClass = ({ column }: CellClassParams): string => {
  return /状态|阶段|就绪/.test(column.label || '') ? 'cell-status' : ''
}
