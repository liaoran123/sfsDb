# sfsDb 预编译环境 Dockerfile
# 用于提供完整的编译环境，支持多架构交叉编译

FROM golang:1.20-bullseye

LABEL maintainer="sfsDb Team"
LABEL version="1.0.0"
LABEL description="sfsDb 预编译环境，支持多架构交叉编译"

# 安装依赖
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    g++ \
    make \
    cmake \
    git \
    curl \
    wget \
    && rm -rf /var/lib/apt/lists/*

# 安装交叉编译工具链（ARM64 和 MIPS64）
RUN apt-get update && apt-get install -y \
    gcc-aarch64-linux-gnu \
    g++-aarch64-linux-gnu \
    gcc-mips64-linux-gnuabi64 \
    g++-mips64-linux-gnuabi64 \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /app

# 复制项目文件
COPY . .

# 设置Go环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io,direct

# 安装依赖
RUN go mod tidy

# 构建脚本权限
RUN chmod +x ./build/build.sh

# 环境变量
ENV SFSDB_HOME=/app
ENV PATH=$PATH:/app/build

# 暴露端口（如果需要）
# EXPOSE 8080

# 默认命令
CMD ["bash"]

# 构建命令示例：
# docker build -t sfsdb-build-env .
# 
# 运行容器示例：
# docker run -it --rm -v $(pwd):/app sfsdb-build-env
# 
# 在容器中编译：
# ./build/build.sh
