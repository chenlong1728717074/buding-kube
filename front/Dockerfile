# 使用nginx作为基础镜像
FROM nginx:alpine

# 复制本地已构建好的dist目录到nginx的默认服务目录
COPY dist/ /usr/share/nginx/html/

# 复制nginx配置文件
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露80端口
EXPOSE 80

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]