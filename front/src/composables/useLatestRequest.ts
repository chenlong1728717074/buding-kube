// 列表查询只采用最后一次请求的结果: 快速切换命名空间时, 先发出但后返回的请求(例如查询全部命名空间)不会覆盖新结果.
// 用法:
//   const latest = useLatestRequest()
//   const isLatest = latest()
//   const resp = await api.list(...)
//   if (!isLatest()) return
export function useLatestRequest() {
  let seq = 0
  return () => {
    const id = ++seq
    return () => id === seq
  }
}
