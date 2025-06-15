# 使用 Alpine Linux 作为基础镜像（最小精简的 Linux 发行版）
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制编译好的二进制文件
COPY kube /app/kube

# 设置可执行权限
RUN chmod +x /app/kube

# 暴露端口
EXPOSE 8080

# 创建非 root 用户（安全最佳实践）
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 切换到非 root 用户
USER appuser

# 启动应用
CMD ["/app/kube"]