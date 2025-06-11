# buding-kube

#### 介绍
buding-kube 是一款轻量级的 Kubernetes 多集群管理工具，支持跨集群资源统一查看、权限集中控制、应用便捷发布，帮助运维和开发团队高效管理多套 K8s 集群。

#### 软件架构
buding-kube 采用前后端分离的架构设计：
- 后端：基于 Golang 和 Gin 框架开发，使用 client-go 与 Kubernetes 集群交互
- 前端：尚未开始
- 数据存储：利用 Kubernetes Secret 资源存储集群配置信息

#### 已实现功能

1. **集群管理**
   - 多集群接入与管理
   - 集群状态监控
   - 集群详情查看
   - 集群添加/删除/更新

2. **资源管理**
   - 命名空间管理
   - Pod 资源查看与操作

3. **系统功能**
   - 跨域请求支持
   - 日志记录
   - 异常恢复机制

#### 开发中功能

以下功能正在开发中：
- 部署管理（Deployment）
- 服务管理（Service）
- 存储管理（PV/PVC）
- 配置管理（ConfigMap/Secret）
- 工作负载管理（StatefulSet/DaemonSet/Job/CronJob）
- 网络管理（Ingress）
- 节点管理（Node）
- 用户权限管理

#### 安装教程

> 项目正在积极开发中，安装文档将在稳定版本发布后提供。

#### 使用说明

> 项目正在积极开发中，详细使用文档将在稳定版本发布后提供。

#### 参与贡献

1. Fork 本仓库
2. 新建 Feat_xxx 分支
3. 提交代码
4. 新建 Pull Request

#### 项目状态

⚠️ **注意：本项目目前处于积极开发阶段，API 和功能可能会有变动。**

欢迎有兴趣的开发者参与贡献，共同完善这个轻量级的 Kubernetes 多集群管理工具！

#### 特技

1.  使用 Readme\_XXX.md 来支持不同的语言，例如 Readme\_en.md, Readme\_zh.md
2.  Gitee 官方博客 [blog.gitee.com](https://blog.gitee.com)
3.  你可以 [https://gitee.com/explore](https://gitee.com/explore) 这个地址来了解 Gitee 上的优秀开源项目
4.  [GVP](https://gitee.com/gvp) 全称是 Gitee 最有价值开源项目，是综合评定出的优秀开源项目
5.  Gitee 官方提供的使用手册 [https://gitee.com/help](https://gitee.com/help)
6.  Gitee 封面人物是一档用来展示 Gitee 会员风采的栏目 [https://gitee.com/gitee-stars/](https://gitee.com/gitee-stars/)
