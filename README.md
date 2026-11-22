# buding-kube

![License](https://img.shields.io/github/license/chenlong1728717074/buding-kube)
![Go Version](https://img.shields.io/badge/Go-1.25+-blue)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.20+-blue)

## 📝 介绍

buding-kube（buding k8s多集群管理平台）是一款轻量级的 Kubernetes 多集群管理工具，支持跨集群资源统一查看、权限集中控制、应用便捷发布，帮助运维和开发团队高效管理多套 K8s 集群。

### 🔗 项目地址

GitHub: https://github.com/chenlong1728717074/buding-kube

## 🏗️ 软件架构

buding-kube 采用前后端分离的架构设计：

- **后端**：基于 Golang 和 Gin 框架开发，使用 client-go 与 Kubernetes 集群交互
- **前端**：Vue 3 + TypeScript + Vite，提供现代化的用户界面
- **数据存储**：利用 Kubernetes Secret 资源存储集群配置信息，无需额外的数据库
- **部署方式**：支持容器化部署，可轻松集成到现有 Kubernetes 集群

## ✅ 已实现功能

### 集群
- 多集群接入、切换与管理，集群概览（资源用量、最近告警事件）
- 节点：资源分配与实时用量、标签 / 污点编辑、停止调度、排空节点（cordon + drain，遵守 PDB）
- 事件中心：按集群 / 命名空间 / 类型 / 资源查看全部事件
- Web kubectl：集群级 kubectl / helm 终端（仅管理员）

### 工作负载
- Deployment / StatefulSet / DaemonSet：分步创建、扩缩容、更新镜像、重新创建、版本回退
- 弹性伸缩（HPA）：按 CPU / 内存使用率自动调整副本数，工作负载详情中可直接设置
- ReplicaSet、Job / CronJob（重新执行、立即执行、暂停）
- Pod：详情、日志、容器终端、文件上传下载、CPU / 内存用量

### 服务与网络
- Service、Ingress（表单创建）、Endpoint / EndpointSlice、IngressClass
- 网络策略（NetworkPolicy）：表单 + 常用模板

### 配置、存储与策略
- ConfigMap / Secret / ServiceAccount
- PV / PVC 分步创建，StorageClass（存储系统模板）
- 资源配额（ResourceQuota）、默认资源限制（LimitRange），命名空间详情中可直接设置
- 中断预算（PDB）、优先级（PriorityClass）
- CRD 及其实例（例如 cert-manager 的 Certificate）的查看与 YAML 编辑

### 通用
- 所有资源支持查看 / 编辑 YAML、YAML 创建
- 标签式导航（一个菜单一个窗口）与页面状态保留

## 📦 快速开始

### 前置条件
- Kubernetes 集群 (v1.20+)
- Go 1.25+ (仅开发需要)
- Node.js 16+ (仅前端开发需要)

### 部署方式

```bash
# 克隆仓库
git clone https://github.com/chenlong1728717074/buding-kube.git
cd buding-kube

# 部署到 Kubernetes（示例）
kubectl apply -f devlop.yaml
```

### 环境变量

| 变量 | 说明 |
| --- | --- |
| `KUBE_RUNTIME_PORT` | 后端监听端口 |
| `JWT_SECRET_KEY` / `KUBE_SECRET_KEY` | 登录令牌签名密钥 / 集群凭证加密密钥，生产环境必须修改 |
| `APP_ENV=demo` | 演示模式：禁止写操作和终端 |
| `KUBECTL_IMAGE` | Web kubectl 使用的镜像，默认 `alpine/k8s:1.35.10`，内网环境改为可访问的镜像 |

> Web kubectl 首次使用时会在目标集群创建命名空间 `buding-kubectl`、绑定 cluster-admin 的服务账户，并为每个用户运行一个容器组（24 小时后回收）；平台接入集群所用的凭证需要有这些权限。
> 生产环境通过 Nginx 代理时，`/api/terminal/` 需要转发 WebSocket（`Upgrade` / `Connection` 请求头）。

## 🔧 开发说明

```bash
# 后端开发
go mod tidy
go run main.go

# 前端开发
cd front
pnpm install
pnpm dev

# 构建与预览
pnpm build
pnpm preview
```

## 🤝 参与贡献

1. Fork 本仓库
2. 新建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## ⚠️ 项目状态

**注意：本项目目前处于积极开发阶段，API 和功能可能会有变动。**

欢迎有兴趣的开发者参与贡献，共同完善这个轻量级的 Kubernetes 多集群管理工具！
