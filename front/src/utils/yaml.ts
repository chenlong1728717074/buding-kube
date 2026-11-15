/**
 * 将接口返回的 JSON 对象转为 YAML 文本(仅用于详情展示, 不做锚点/折叠等高级特性)
 */
export function jsonToYaml(value: unknown, indent = 0): string {
  const pad = '  '.repeat(indent)

  if (value === null || value === undefined) return 'null'
  if (typeof value === 'string') return formatScalar(value)
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)

  if (Array.isArray(value)) {
    if (value.length === 0) return `${pad}[]`
    return value
      .map((item) => {
        if (item !== null && typeof item === 'object') {
          const body = jsonToYaml(item, indent + 1)
          // 首行与 "- " 同行, 其余键与首个键对齐
          return `${pad}- ${body.slice((indent + 1) * 2)}`
        }
        return `${pad}- ${formatScalar(String(item))}`
      })
      .join('\n')
  }

  const entries = Object.entries(value as Record<string, unknown>)
  if (entries.length === 0) return `${pad}{}`
  return entries
    .map(([key, val]) => {
      if (val !== null && typeof val === 'object') {
        return `${pad}${key}:\n${jsonToYaml(val, indent + 1)}`
      }
      return `${pad}${key}: ${formatScalar(String(val))}`
    })
    .join('\n')
}

function formatScalar(text: string): string {
  if (text === '') return '""'
  // 数字样式/布尔样式/空白/特殊字符的字符串需要加引号, 防止展示歧义
  if (
    /^-?\d+(\.\d+)?$/.test(text) ||
    text === 'true' ||
    text === 'false' ||
    text === 'null' ||
    text === '~' ||
    /^\s|\s$/.test(text) ||
    text.includes('\n') ||
    /[:#{}[\],&*?|>'"%@`]/.test(text)
  ) {
    return JSON.stringify(text)
  }
  return text
}
