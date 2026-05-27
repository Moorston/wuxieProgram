# 武俱打卡 - 部署文档

## 一、服务器要求

### 最低配置（开发/测试）

| 资源 | 配置 |
|------|------|
| CPU | 2 核 |
| 内存 | 4 GB |
| 硬盘 | 50 GB SSD |
| 带宽 | 5 Mbps |
| 系统 | Ubuntu 22.04 LTS |

### 推荐配置（生产环境）

| 资源 | 配置 |
|------|------|
| CPU | 4 核 |
| 内存 | 8 GB |
| 硬盘 | 200 GB SSD |
| 带宽 | 20 Mbps |
| 系统 | Ubuntu 22.04 LTS |

### 资源消耗估算

| 服务 | CPU | 内存 | 说明 |
|------|-----|------|------|
| MongoDB | 0.5 核 | 512MB-1GB | 文档数据库 |
| Redis | 0.1 核 | 256MB | 缓存+队列 |
| MinIO | 0.5 核 | 512MB | 对象存储 |
| API Server | 0.5 核 | 256MB | 业务逻辑 |
| Media Server | 1-2 核 | 512MB-1GB | FFmpeg转码(最耗资源) |
| Nginx | 0.1 核 | 64MB | 反向代理 |
| **总计** | **2.5-4 核** | **2-3 GB** | - |

---

## 二、环境准备

### 1. 安装 Docker

```bash
# Ubuntu
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# 验证
docker --version
docker compose version
```

### 2. 安装 Git

```bash
sudo apt update && sudo apt install -y git
```

### 3. 克隆项目

```bash
git clone https://github.com/Moorston/wuxieProgram.git
cd wuxieProgram
```

---

## 三、配置文件

### 1. 创建环境变量文件

```bash
cd deploy
cp .env.example .env
```

### 2. 修改 .env

```bash
vim .env
```

关键配置项：

```bash
# MinIO 账号密码（生产环境务必修改）
MINIO_USER=minioadmin
MINIO_PASSWORD=your-strong-password

# JWT 密钥（生产环境务必修改）
JWT_SECRET=your-random-secret-key

# 微信小程序配置（必须填写真实值）
WX_APP_ID=your-wx-app-id
WX_SECRET=your-wx-app-secret
WX_TEMPLATE_ID=your-subscribe-template-id
```

### 3. 修改后端配置

```bash
# API Server 配置
vim api-server/configs/config.yaml

# Media Server 配置
vim media-server/configs/config.yaml
```

确保配置中的服务地址与 Docker Compose 服务名一致：

```yaml
# api-server/configs/config.yaml
mongo:
  uri: "mongodb://mongo:27017"
redis:
  addr: "redis:6379"
media_url: "http://media-server:8081"

# media-server/configs/config.yaml
minio:
  endpoint: "minio:9000"
redis:
  addr: "redis:6379"
api_server: "http://api-server:8080"
```

---

## 四、部署方式

### 方式一：deploy.sh 脚本（推荐）

```bash
cd deploy

# 启动开发环境
./deploy.sh up

# 启动生产环境
./deploy.sh up prod

# 其他命令
./deploy.sh down        # 停止
./deploy.sh restart     # 重启
./deploy.sh logs        # 查看日志
./deploy.sh status      # 查看状态
./deploy.sh rebuild     # 强制重建
./deploy.sh clean       # 清理
```

### 方式二：Makefile

```bash
cd deploy

make dev          # 启动开发环境
make prod         # 启动生产环境
make down         # 停止
make logs         # 查看日志
make status       # 查看状态
make rebuild      # 强制重建
make clean        # 清理
make shell-api    # 进入 API Server 容器
make shell-mongo  # 进入 MongoDB shell
```

### 方式三：Docker Compose 直接使用

```bash
cd deploy

# 开发环境（自动合并 override 配置）
docker compose up -d

# 生产环境（显式指定 prod 配置）
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f

# 停止
docker compose down
```

---

## 五、服务访问

部署成功后，通过以下地址访问：

| 服务 | 地址 | 说明 |
|------|------|------|
| **应用入口** | `http://服务器IP` | Nginx 统一入口 |
| API Server | `http://服务器IP:8080` | 业务接口 |
| Media Server | `http://服务器IP:8081` | 视频服务 |
| MinIO 控制台 | `http://服务器IP:9001` | 存储管理 |
| MongoDB | `服务器IP:27017` | 数据库 |
| Redis | `服务器IP:6379` | 缓存 |

MinIO 控制台登录账号：`${MINIO_USER}` / `${MINIO_PASSWORD}`

---

## 六、验证部署

### 1. 检查服务状态

```bash
docker compose ps
```

所有服务应显示 `Up` 状态。

### 2. 检查健康检查

```bash
docker compose ps --format "table {{.Name}}\t{{.Status}}"
```

### 3. 测试 API

```bash
# 健康检查
curl http://localhost/health

# API 测试
curl http://localhost:8080/api/auth/login
```

### 4. 查看日志

```bash
# 所有服务日志
docker compose logs -f

# 单个服务日志
docker compose logs -f api-server
docker compose logs -f media-server
```

---

## 七、常用运维操作

### 查看日志

```bash
# 实时日志
docker compose logs -f

# 最近 100 行
docker compose logs --tail=100

# 指定服务
docker compose logs -f api-server
```

### 重启服务

```bash
# 重启所有
docker compose restart

# 重启单个服务
docker compose restart api-server
```

### 进入容器调试

```bash
# 进入 API Server
docker compose exec api-server sh

# 进入 MongoDB
docker compose exec mongo mongosh -u root -p root wuxie

# 进入 Redis
docker compose exec redis redis-cli
```

### 更新代码

```bash
cd wuxieProgram
git pull

cd deploy
./deploy.sh rebuild
```

### 备份数据

```bash
# 备份 MongoDB
docker compose exec mongo mongodump -u root -p root --authenticationDatabase admin -d wuxie -o /tmp/backup
docker cp wuxie-mongo:/tmp/backup ./backup

# 备份 MinIO 数据
docker cp wuxie-minio:/data ./minio-backup
```

### 恢复数据

```bash
# 恢复 MongoDB
docker cp ./backup wuxie-mongo:/tmp/backup
docker compose exec mongo mongorestore -u root -p root --authenticationDatabase admin -d wuxie /tmp/backup/wuxie
```

---

## 八、SSL 证书配置

### 1. 获取证书

```bash
# 使用 Let's Encrypt
sudo apt install certbot
sudo certbot certonly --standalone -d your-domain.com
```

### 2. 配置 Nginx

修改 `deploy/nginx.conf`：

```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /etc/nginx/ssl/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;

    # ... 其余配置不变
}
```

### 3. 挂载证书

修改 `docker-compose.yml` 的 nginx 服务：

```yaml
nginx:
  volumes:
    - ./nginx.conf:/etc/nginx/conf.d/default.conf
    - /etc/letsencrypt:/etc/nginx/ssl:ro
```

---

## 九、常见问题

### Q1: 服务启动失败

```bash
# 查看具体错误
docker compose logs api-server

# 常见原因：
# 1. 配置文件路径错误
# 2. MongoDB/Redis 未就绪
# 3. 端口被占用
```

### Q2: 视频转码失败

```bash
# 检查 Media Server 日志
docker compose logs -f media-server

# 常见原因：
# 1. FFmpeg 未安装（Docker 镜像内已包含）
# 2. 磁盘空间不足
# 3. MinIO 连接失败
```

### Q3: 上传失败

```bash
# 检查 MinIO 状态
docker compose logs minio

# 检查网络连接
docker compose exec api-server curl http://minio:9000/minio/health/live
```

### Q4: 微信登录失败

```bash
# 检查配置
grep -A3 "wx:" api-server/configs/config.yaml

# 常见原因：
# 1. AppID/Secret 错误
# 2. 服务器未配置微信回调域名
# 3. 微信接口限制（测试阶段需配置白名单）
```

### Q5: 磁盘空间不足

```bash
# 查看磁盘使用
docker system df

# 清理未使用的资源
docker system prune -a

# 清理旧的 Docker 卷
docker volume prune
```

---

## 十、生产环境 Checklist

- [ ] 修改 `.env` 中的所有密码和密钥
- [ ] 修改 `api-server/configs/config.yaml` 中的 JWT 密钥
- [ ] 配置微信小程序真实 AppID/Secret
- [ ] 配置 SSL 证书（HTTPS）
- [ ] 配置域名解析
- [ ] 配置防火墙规则（仅开放 80/443/22）
- [ ] 配置自动备份（MongoDB + MinIO）
- [ ] 配置日志轮转（已在 prod 配置中启用）
- [ ] 配置监控告警（可选：Prometheus + Grafana）
- [ ] 测试所有功能正常

---

## 十一、端口清单

| 端口 | 服务 | 生产环境建议 |
|------|------|-------------|
| 80 | Nginx HTTP | 开放 |
| 443 | Nginx HTTPS | 开放 |
| 22 | SSH | 开放（限制IP） |
| 8080 | API Server | 仅内网 |
| 8081 | Media Server | 仅内网 |
| 9000 | MinIO API | 仅内网 |
| 9001 | MinIO Console | 仅内网或关闭 |
| 27017 | MongoDB | 仅内网 |
| 6379 | Redis | 仅内网 |
