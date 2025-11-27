#!/bin/bash

# 设置变量
IMAGE_NAME="rate"
DOCKER_HUB_USER="yexk"  # 
TAG="2.0.0"

# 构建 Docker 镜像
echo "正在构建 Docker 镜像..."
docker build -t ${IMAGE_NAME}:${TAG} .

# 标记镜像用于推送到 Docker Hub
echo "标记镜像..."
docker tag ${IMAGE_NAME}:${TAG} ${DOCKER_HUB_USER}/${IMAGE_NAME}:${TAG}

# 推送到 Docker Hub
echo "推送到 Docker Hub..."
docker push ${DOCKER_HUB_USER}/${IMAGE_NAME}:${TAG}

echo "构建和推送完成！"
echo "镜像地址: ${DOCKER_HUB_USER}/${IMAGE_NAME}:${TAG}"