# 汇率通知服务

这是一个使用 Go 语言开发的汇率通知服务，通过 Wise API 获取实时汇率，并通过 Lark webhook 发送通知。

## 功能特性

- 每小时自动获取 USD-CNY、MYR-CNY、MYR-HKD 汇率
- 通过 Lark webhook 发送格式化通知
- 支持定时任务，每小时整点执行
- 包含时间戳显示
- 动态显示汇率涨跌趋势（↑↓→）
- Docker 支持，可部署到任何支持 Docker 的环境

## 环境变量

在使用前需要设置以下环境变量：

```bash
export LARK_WEBHOOK_URL="your_lark_webhook_url"
```

## 本地运行

1. 确保已安装 Go 1.21 或更高版本
2. 克隆项目并进入目录
3. 设置环境变量
4. 运行程序：

```bash
go run .
```

或者编译后运行：

```bash
go build -o rate-notifier
./rate-notifier
```

## Docker 部署

### 方法一：使用 docker-compose（推荐）

1. 复制并编辑环境变量：
```bash
cp .env.example .env
# 编辑 .env 文件，设置你的 LARK_WEBHOOK_URL
```

2. 运行服务：
```bash
docker-compose up -d
```

### 方法二：使用 Docker 命令

1. 构建镜像：
```bash
docker build -t rate-notifier .
```

2. 运行容器：
```bash
docker run -d \
  --name rate-notifier \
  --restart unless-stopped \
  -e LARK_WEBHOOK_URL="your_webhook_url" \
  rate-notifier
```

### 方法三：推送到 Docker Hub

1. 编辑 `build-and-push.sh` 文件，将 `your-docker-hub-username` 替换为你的 Docker Hub 用户名

2. 构建并推送：
```bash
chmod +x build-and-push.sh
./build-and-push.sh
```

3. 在其他服务器上拉取并运行：
```bash
docker pull your-docker-hub-user/rate-notifier:latest
docker run -d \
  --name rate-notifier \
  --restart unless-stopped \
  -e LARK_WEBHOOK_URL="your_webhook_url" \
  your-docker-hub-user/rate-notifier:latest
```

## 消息格式

发送到 Lark 的消息格式如下：

```
美金USD-CNY, 结汇: 7.169300 ↑
马币MYR-CNY, 结汇: 1.543411 ↓
马币MYR-HKD, 结汇: 1.685621 →
更新时间: 2024-01-01 10:00:00
```

箭头含义：
- ↑ 汇率上升
- ↓ 汇率下降
- → 汇率不变

## 部署建议

- 使用 systemd 或其他进程管理工具确保服务持续运行
- 可以部署到云服务器或本地服务器
- 建议设置日志轮转以避免日志文件过大
- Docker 部署推荐使用 docker-compose 进行管理

## 注意事项

- 程序会在每个整点执行（如 10:00, 11:00, 12:00）
- 第一次运行时所有箭头显示为 ↑
- 需要稳定的网络连接以访问 Wise API
- 确保 Lark webhook URL 正确且有效