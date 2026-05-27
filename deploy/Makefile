.PHONY: help dev prod up down restart logs status rebuild clean

# 默认目标
help: ## 显示帮助信息
	@echo ""
	@echo "武俱打卡 - Docker Compose 部署"
	@echo "=========================================="
	@echo ""
	@echo "用法: make <命令>"
	@echo ""
	@echo "开发环境:"
	@echo "  make dev          启动开发环境 (自动热重载)"
	@echo "  make dev-down     停止开发环境"
	@echo "  make dev-logs     查看开发环境日志"
	@echo ""
	@echo "生产环境:"
	@echo "  make prod         启动生产环境"
	@echo "  make prod-down    停止生产环境"
	@echo "  make prod-logs    查看生产环境日志"
	@echo ""
	@echo "通用命令:"
	@echo "  make up           启动服务 (默认开发模式)"
	@echo "  make down         停止服务"
	@echo "  make restart      重启服务"
	@echo "  make logs         查看日志"
	@echo "  make status       查看状态"
	@echo "  make rebuild      强制重建"
	@echo "  make clean        清理未使用的镜像和容器"
	@echo "  make shell-api    进入 API Server 容器"
	@echo "  make shell-mongo  进入 MongoDB shell"
	@echo ""

# ---------- 开发环境 ----------

dev: ## 启动开发环境
	@echo "启动开发环境..."
	@docker compose up -d --build
	@echo ""
	@echo "开发环境已启动:"
	@echo "  http://localhost       (Nginx)"
	@echo "  http://localhost:8080  (API Server)"
	@echo "  http://localhost:8081  (Media Server)"
	@echo "  http://localhost:9001  (MinIO Console)"

dev-down: ## 停止开发环境
	docker compose down

dev-logs: ## 查看开发环境日志
	docker compose logs -f --tail=100

# ---------- 生产环境 ----------

prod: ## 启动生产环境
	@echo "启动生产环境..."
	@docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
	@echo ""
	@echo "生产环境已启动:"
	@echo "  http://localhost (Nginx)"

prod-down: ## 停止生产环境
	docker compose -f docker-compose.yml -f docker-compose.prod.yml down

prod-logs: ## 查看生产环境日志
	docker compose -f docker-compose.yml -f docker-compose.prod.yml logs -f --tail=100

# ---------- 通用命令 ----------

up: dev ## 启动服务 (默认开发模式)

down: ## 停止服务
	docker compose down

restart: ## 重启服务
	docker compose down --remove-orphans 2>/dev/null || true
	docker compose up -d --build

logs: ## 查看日志
	docker compose logs -f --tail=100

status: ## 查看服务状态
	@docker compose ps
	@echo ""

rebuild: ## 强制重建并启动
	docker compose down --remove-orphans 2>/dev/null || true
	docker compose build --no-cache
	docker compose up -d

clean: ## 清理未使用的镜像和容器
	docker compose down --remove-orphans --rmi local 2>/dev/null || true
	docker system prune -f
	@echo "清理完成"

# ---------- 工具命令 ----------

shell-api: ## 进入 API Server 容器
	docker compose exec api-server sh

shell-mongo: ## 进入 MongoDB shell
	docker compose exec mongo mongosh -u root -p root --authenticationDatabase admin wuxie

shell-redis: ## 进入 Redis CLI
	docker compose exec redis redis-cli

shell-media: ## 进入 Media Server 容器
	docker compose exec media-server sh
