#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
print_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
print_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo ""
echo "=========================================="
echo "  武俱打卡 - Docker Compose 部署"
echo "=========================================="
echo ""

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    print_error "未安装 Docker"
    exit 1
fi

if ! docker compose version &> /dev/null; then
    print_error "未安装 docker compose 插件"
    exit 1
fi

# 检查配置文件
if [ ! -f "$PROJECT_DIR/api-server/configs/config.yaml" ]; then
    print_error "api-server/configs/config.yaml 不存在"
    exit 1
fi

if [ ! -f "$PROJECT_DIR/media-server/configs/config.yaml" ]; then
    print_error "media-server/configs/config.yaml 不存在"
    exit 1
fi

# 创建 .env 文件（如果不存在）
if [ ! -f "$SCRIPT_DIR/.env" ]; then
    if [ -f "$SCRIPT_DIR/.env.example" ]; then
        cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"
        print_warn "已创建 .env 文件，请根据需要修改配置"
    fi
fi

# 解析参数
ACTION="${1:-up}"
MODE="${2:-dev}"

# Docker Compose 命令
DC="docker compose"

# 生产模式使用 prod 配置
if [ "$MODE" = "prod" ]; then
    DC="docker compose -f docker-compose.yml -f docker-compose.prod.yml"
fi

case $ACTION in
    up|start)
        print_info "停止旧容器..."
        cd "$SCRIPT_DIR"
        $DC down --remove-orphans 2>/dev/null || true

        print_info "构建并启动服务..."
        $DC up -d --build

        print_info "等待服务就绪..."
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
        echo "查看日志:  ./deploy.sh logs"
        echo "停止服务:  ./deploy.sh down"
        echo ""
        ;;

    down|stop)
        print_info "停止所有服务..."
        cd "$SCRIPT_DIR"
        $DC down
        print_ok "已停止"
        ;;

    restart)
        print_info "重启所有服务..."
        cd "$SCRIPT_DIR"
        $DC down --remove-orphans 2>/dev/null || true
        $DC up -d --build
        print_ok "已重启"
        ;;

    logs)
        cd "$SCRIPT_DIR"
        $DC logs -f --tail=100
        ;;

    status)
        cd "$SCRIPT_DIR"
        $DC ps
        echo ""
        ;;

    rebuild)
        print_info "强制重新构建..."
        cd "$SCRIPT_DIR"
        $DC down --remove-orphans 2>/dev/null || true
        $DC build --no-cache
        $DC up -d
        print_ok "已重建并启动"
        ;;

    clean)
        print_info "清理未使用的镜像和容器..."
        cd "$SCRIPT_DIR"
        $DC down --remove-orphans --rmi local 2>/dev/null || true
        docker system prune -f
        print_ok "清理完成"
        ;;

    *)
        echo "用法: $0 {up|down|restart|logs|status|rebuild|clean} [dev|prod]"
        echo ""
        echo "命令:"
        echo "  up/start   - 构建并启动所有服务"
        echo "  down/stop  - 停止所有服务"
        echo "  restart    - 重启所有服务"
        echo "  logs       - 查看服务日志"
        echo "  status     - 查看服务状态"
        echo "  rebuild    - 强制重新构建并启动"
        echo "  clean      - 清理未使用的镜像和容器"
        echo ""
        echo "模式:"
        echo "  dev        - 开发环境 (默认)"
        echo "  prod       - 生产环境 (含资源限制、日志轮转)"
        echo ""
        echo "示例:"
        echo "  ./deploy.sh up           # 启动开发环境"
        echo "  ./deploy.sh up prod      # 启动生产环境"
        echo "  ./deploy.sh down prod    # 停止生产环境"
        exit 1
        ;;
esac
