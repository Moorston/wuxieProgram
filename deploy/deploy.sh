#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=========================================="
echo "  武俱打卡 - Docker Compose 部署"
echo "=========================================="

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "错误: 未安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "错误: 未安装 docker-compose"
    exit 1
fi

# 检查配置文件
if [ ! -f "$PROJECT_DIR/api-server/configs/config.yaml" ]; then
    echo "错误: api-server/configs/config.yaml 不存在"
    exit 1
fi

if [ ! -f "$PROJECT_DIR/media-server/configs/config.yaml" ]; then
    echo "错误: media-server/configs/config.yaml 不存在"
    exit 1
fi

# 解析参数
ACTION="${1:-up}"

case $ACTION in
    up|start)
        echo ""
        echo "[1/3] 停止旧容器..."
        cd "$SCRIPT_DIR"
        docker compose down --remove-orphans 2>/dev/null || true

        echo "[2/3] 构建并启动服务..."
        docker compose up -d --build

        echo "[3/3] 等待服务就绪..."
        sleep 5

        echo ""
        echo "=========================================="
        echo "  部署完成！"
        echo "=========================================="
        echo ""
        echo "服务访问地址:"
        echo "  Nginx 统一入口:  http://localhost"
        echo "  API Server:      http://localhost:8080"
        echo "  Media Server:    http://localhost:8081"
        echo "  MinIO 控制台:    http://localhost:9001"
        echo "  MongoDB:         localhost:27017"
        echo "  Redis:           localhost:6379"
        echo ""
        echo "MinIO 账号: ${MINIO_USER:-minioadmin} / ${MINIO_PASSWORD:-minioadmin}"
        echo ""
        echo "查看日志:  docker compose logs -f"
        echo "停止服务:  ./deploy.sh down"
        echo ""
        ;;

    down|stop)
        echo "停止所有服务..."
        cd "$SCRIPT_DIR"
        docker compose down
        echo "已停止"
        ;;

    restart)
        echo "重启所有服务..."
        cd "$SCRIPT_DIR"
        docker compose down --remove-orphans 2>/dev/null || true
        docker compose up -d --build
        echo "已重启"
        ;;

    logs)
        cd "$SCRIPT_DIR"
        docker compose logs -f --tail=100
        ;;

    status)
        cd "$SCRIPT_DIR"
        docker compose ps
        ;;

    rebuild)
        echo "强制重新构建..."
        cd "$SCRIPT_DIR"
        docker compose down --remove-orphans 2>/dev/null || true
        docker compose build --no-cache
        docker compose up -d
        echo "已重建并启动"
        ;;

    *)
        echo "用法: $0 {up|down|restart|logs|status|rebuild}"
        echo ""
        echo "  up/start  - 构建并启动所有服务"
        echo "  down/stop - 停止所有服务"
        echo "  restart   - 重启所有服务"
        echo "  logs      - 查看服务日志"
        echo "  status    - 查看服务状态"
        echo "  rebuild   - 强制重新构建并启动"
        exit 1
        ;;
esac
