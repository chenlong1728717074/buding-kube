# buding-kube

![License](https://img.shields.io/github/license/chenlong1728717074/buding-kube)
![Go Version](https://img.shields.io/badge/Go-1.18+-blue)
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

### 集群管理
- 多集群接入与管理
- 集群状态监控
- 集群详情查看
- 集群添加/删除/更新

### 资源管理
- 命名空间管理
- Pod 资源查看与操作
  - Pod 详情查看
  - Pod 日志查看
  - Pod 终端操作
  - Pod 删除操作
  - Pod 驱逐
  - 文件上传/下载（运行态）
  - 运行态操作守卫（非运行态禁用日志/上传/下载/进入容器）

### 体验优化
- 标签式导航与页面状态保留
- 登录与管理页浅蓝灰主题、圆润风格
- 统一错误提示与 401 自动跳转登录

### 系统功能
- RESTful API
- 跨域请求支持
- 结构化日志记录
- 健壮的异常恢复机制

## 🚧 开发中功能

以下功能正在积极开发中：

- **工作负载管理**
  - Deployment
  - StatefulSet
  - DaemonSet
  - Job/CronJob

- **服务与网络**
  - Service
  - Ingress
  - NetworkPolicy

- **配置与存储**
  - ConfigMap/Secret
  - PV/PVC
  - StorageClass

- **集群管理**
  - Node 管理
  - 集群监控
  - 事件查看

- **安全性**
  - 用户认证
  - 权限管理
  - 资源隔离

## 📦 快速开始

### 前置条件
- Kubernetes 集群 (v1.20+)
- Go 1.18+ (仅开发需要)
- Node.js 16+ (仅前端开发需要)

### 部署方式

```bash
# 克隆仓库
git clone https://github.com/chenlong1728717074/buding-kube.git
cd buding-kube

# 部署到 Kubernetes（示例）
kubectl apply -f devlop.yaml
```

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

#### 特技

1.  使用 Readme\_XXX.md 来支持不同的语言，例如 Readme\_en.md, Readme\_zh.md
2.  Gitee 官方博客 [blog.gitee.com](https://blog.gitee.com)
3.  你可以 [https://gitee.com/explore](https://gitee.com/explore) 这个地址来了解 Gitee 上的优秀开源项目
4.  [GVP](https://gitee.com/gvp) 全称是 Gitee 最有价值开源项目，是综合评定出的优秀开源项目
5.  Gitee 官方提供的使用手册 [https://gitee.com/help](https://gitee.com/help)
6.  Gitee 封面人物是一档用来展示 Gitee 会员风采的栏目 [https://gitee.com/gitee-stars/](https://gitee.com/gitee-stars/)
