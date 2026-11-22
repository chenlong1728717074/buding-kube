import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { RouteLocationNormalized, Router } from 'vue-router'

// 顶部标签(窗口)
// 规则: 只有点击左侧菜单才会打开新标签; 其他跳转(进入详情、详情之间互相跳转、返回)都在当前标签内进行.
// 每个标签记录自己的浏览记录(stack), 详情页的「返回」回到这个标签的上一页, 不受其他标签影响
export interface TagItem {
  // 标签归属的菜单页路径, 例如 /cluster/host/job/job
  key: string
  // 当前显示的页面标题
  title: string
  // 当前显示的页面
  fullPath: string
  // 本标签内的浏览记录, 最后一项为当前页面
  stack: string[]
}

type Intent = { type: 'menu' | 'tag'; key: string } | null

const STORAGE_KEY = 'tagsView.v2'
const MAX_STACK = 20

// 去掉刷新用的 __refresh 参数
export const normalizeFullPath = (fp: string): string => {
  try {
    const url = new URL(fp, window.location.origin)
    url.searchParams.delete('__refresh')
    const qs = url.searchParams.toString()
    return url.pathname + (qs ? `?${qs}` : '')
  } catch {
    return fp
  }
}

const pathOf = (fp: string) => fp.split('?')[0]

// 标签归属: 菜单页是它自己; 详情页通过路由 meta.parent 归属到对应的菜单页
const rootOf = (route: RouteLocationNormalized) => {
  const parent = route.meta.parent as string | undefined
  const clusterId = route.params.clusterId as string | undefined
  return parent && clusterId ? `/cluster/${clusterId}/${parent}` : route.path
}

const load = (): { tags: TagItem[]; activeKey: string } => {
  try {
    localStorage.removeItem('tagsView') // 旧版本按完整地址保存, 不再使用
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const data = JSON.parse(raw)
      const tags = (Array.isArray(data.tags) ? data.tags : []).filter((t: TagItem) => t && t.key && t.fullPath)
      tags.forEach((t: TagItem) => {
        if (!Array.isArray(t.stack) || !t.stack.length) t.stack = [t.fullPath]
      })
      return { tags, activeKey: String(data.activeKey || '') }
    }
  } catch {
    // 数据损坏时重新开始
  }
  return { tags: [], activeKey: '' }
}

export const useTagsStore = defineStore('tags', () => {
  const saved = load()
  const tags = ref<TagItem[]>(saved.tags)
  const activeKey = ref(saved.activeKey)
  // 下一次跳转的来源; 没有时表示在当前标签内跳转
  let intent: Intent = null

  const activeTag = computed(() => tags.value.find(t => t.key === activeKey.value) || null)

  const persist = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ tags: tags.value, activeKey: activeKey.value }))
    } catch {
      // 忽略存储失败
    }
  }

  const pushStack = (tag: TagItem, fullPath: string) => {
    const s = tag.stack
    if (s[s.length - 1] === fullPath) return
    // 回到上一页(返回按钮、浏览器后退、删除后回到列表)
    if (s.length >= 2 && s[s.length - 2] === fullPath) {
      s.pop()
      return
    }
    s.push(fullPath)
    if (s.length > MAX_STACK) s.splice(0, s.length - MAX_STACK)
  }

  // 每次路由跳转完成后调用(router.afterEach)
  // initial: 打开或刷新浏览器后的第一次跳转
  const onRoute = (route: RouteLocationNormalized, initial = false) => {
    const current = intent
    intent = null
    if (!route.path.startsWith('/cluster/') || !route.params.clusterId) return

    const fullPath = normalizeFullPath(route.fullPath)
    const title = String(route.meta.title || route.name || route.path)
    let tag: TagItem | undefined

    if (current?.type === 'menu') {
      // 点击菜单: 打开(或切换到)该菜单的标签, 并回到菜单页本身
      tag = tags.value.find(t => t.key === current.key)
      if (!tag) {
        tag = { key: current.key, title, fullPath, stack: [] }
        tags.value.push(tag)
      }
      tag.stack = [fullPath]
    } else {
      if (current?.type === 'tag') tag = tags.value.find(t => t.key === current.key)
      // 刷新页面或直接打开地址时, 找显示该页面或所属菜单的标签
      if (!tag && initial) {
        tag = tags.value.find(t => t.fullPath === fullPath) || tags.value.find(t => t.key === rootOf(route))
      }
      // 当前标签内跳转
      if (!tag) tag = activeTag.value || undefined
      if (!tag) {
        tag = { key: rootOf(route), title, fullPath, stack: [] }
        tags.value.push(tag)
      }
      pushStack(tag, fullPath)
    }

    tag.fullPath = fullPath
    tag.title = title
    activeKey.value = tag.key
    persist()
  }

  const clearIntent = () => {
    intent = null
  }

  const navigate = (router: Router, to: string, next: Intent) => {
    const currentFull = normalizeFullPath(router.currentRoute.value.fullPath)
    if (normalizeFullPath(to) === currentFull) {
      // 地址没变不会触发跳转, 直接处理
      intent = next
      onRoute(router.currentRoute.value)
      return
    }
    intent = next
    router.push(to).then(failure => {
      if (failure) intent = null
    })
  }

  // 点击左侧菜单
  const openMenu = (router: Router, path: string) => navigate(router, path, { type: 'menu', key: pathOf(path) })

  // 点击标签: 恢复它停留的页面
  const activate = (router: Router, tag: TagItem) => navigate(router, tag.fullPath, { type: 'tag', key: tag.key })

  // 刷新标签: 加一个随机参数重新创建页面
  const refresh = (router: Router, tag: TagItem) => {
    const url = new URL(tag.fullPath, window.location.origin)
    url.searchParams.set('__refresh', String(Date.now()))
    intent = { type: 'tag', key: tag.key }
    router.push(url.pathname + url.search).then(failure => {
      if (failure) intent = null
    })
  }

  // 详情页的返回: 回到当前标签的上一页; 直接打开的详情页回到所属的列表页
  const back = (router: Router) => {
    const tag = activeTag.value
    const current = normalizeFullPath(router.currentRoute.value.fullPath)
    let prev = tag && tag.stack.length >= 2 ? tag.stack[tag.stack.length - 2] : ''
    if (!prev && tag && pathOf(current) !== tag.key) prev = tag.key
    if (!prev) {
      router.back()
      return
    }
    // 浏览器历史的上一页正好是它时用后退, 保持历史记录干净
    const historyBack = (window.history.state as { back?: string } | null)?.back
    if (historyBack && normalizeFullPath(historyBack) === prev) router.back()
    else router.push(prev)
  }

  const close = (router: Router, tag: TagItem) => {
    const idx = tags.value.findIndex(t => t.key === tag.key)
    if (idx < 0) return
    tags.value.splice(idx, 1)
    if (activeKey.value === tag.key) {
      const next = tags.value[idx - 1] || tags.value[idx]
      activeKey.value = ''
      if (next) activate(router, next)
    }
    persist()
  }

  const closeLeft = (tag: TagItem) => {
    const idx = tags.value.findIndex(t => t.key === tag.key)
    if (idx <= 0) return
    tags.value.splice(0, idx)
    persist()
  }

  const closeRight = (tag: TagItem) => {
    const idx = tags.value.findIndex(t => t.key === tag.key)
    if (idx < 0) return
    tags.value.splice(idx + 1)
    persist()
  }

  const closeAll = (router: Router) => {
    tags.value = []
    activeKey.value = ''
    persist()
    const matched = router.currentRoute.value.path.match(/^\/cluster\/([^/]+)/)
    if (matched?.[1]) openMenu(router, `/cluster/${matched[1]}/overview`)
    else router.push('/home')
  }

  return { tags, activeKey, activeTag, onRoute, clearIntent, openMenu, activate, refresh, back, close, closeLeft, closeRight, closeAll }
})
