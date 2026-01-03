#!/bin/bash

# Gin Web 应用部署脚本
# 使用方法: ./scripts/deploy.sh [环境]
# 环境: dev, prod (默认: prod)

set -e

ENV=${1:-prod}
APP_NAME="gin_web"
APP_DIR="/opt/$APP_NAME"
SERVICE_NAME="gin-web"
BUILD_DIR="./bin"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then 
        log_error "请使用 sudo 运行此脚本"
        exit 1
    fi
}

# 构建应用
build_app() {
    log_info "开始构建应用..."
    
    # 创建构建目录
    mkdir -p $BUILD_DIR
    
    # 构建
    if [ "$ENV" = "prod" ]; then
        log_info "构建生产版本..."
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o $BUILD_DIR/$APP_NAME .
    else
        log_info "构建开发版本..."
        go build -o $BUILD_DIR/$APP_NAME .
    fi
    
    if [ $? -eq 0 ]; then
        log_info "构建成功: $BUILD_DIR/$APP_NAME"
    else
        log_error "构建失败"
        exit 1
    fi
}

# 停止服务
stop_service() {
    log_info "停止服务..."
    if systemctl is-active --quiet $SERVICE_NAME; then
        systemctl stop $SERVICE_NAME
        log_info "服务已停止"
    else
        log_warn "服务未运行"
    fi
}

# 备份旧版本
backup_old_version() {
    if [ -f "$APP_DIR/$APP_NAME" ]; then
        log_info "备份旧版本..."
        BACKUP_FILE="$APP_DIR/${APP_NAME}.backup.$(date +%Y%m%d_%H%M%S)"
        cp $APP_DIR/$APP_NAME $BACKUP_FILE
        log_info "备份完成: $BACKUP_FILE"
    fi
}

# 部署文件
deploy_files() {
    log_info "部署文件..."
    
    # 创建应用目录
    mkdir -p $APP_DIR
    mkdir -p $APP_DIR/logs
    mkdir -p $APP_DIR/configs
    
    # 复制可执行文件
    cp $BUILD_DIR/$APP_NAME $APP_DIR/
    chmod +x $APP_DIR/$APP_NAME
    
    # 复制配置文件（如果不存在）
    if [ ! -f "$APP_DIR/configs/config.yaml" ]; then
        log_warn "配置文件不存在，复制默认配置..."
        cp configs/config.yaml $APP_DIR/configs/
        log_warn "请编辑配置文件: $APP_DIR/configs/config.yaml"
    fi
    
    log_info "文件部署完成"
}

# 启动服务
start_service() {
    log_info "启动服务..."
    
    # 检查 systemd 服务文件
    if [ ! -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
        log_warn "Systemd 服务文件不存在，创建中..."
        create_systemd_service
    fi
    
    # 重新加载 systemd
    systemctl daemon-reload
    
    # 启动服务
    systemctl start $SERVICE_NAME
    
    # 等待服务启动
    sleep 2
    
    # 检查服务状态
    if systemctl is-active --quiet $SERVICE_NAME; then
        log_info "服务启动成功"
        systemctl status $SERVICE_NAME --no-pager
    else
        log_error "服务启动失败"
        journalctl -u $SERVICE_NAME -n 50 --no-pager
        exit 1
    fi
}

# 创建 systemd 服务文件
create_systemd_service() {
    log_info "创建 systemd 服务文件..."
    
    cat > /etc/systemd/system/$SERVICE_NAME.service <<EOF
[Unit]
Description=Gin Web Application
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/$APP_NAME
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$APP_NAME

Environment="GIN_MODE=release"

LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

    log_info "Systemd 服务文件已创建"
}

# 主函数
main() {
    log_info "开始部署 $APP_NAME (环境: $ENV)"
    
    # 检查 root
    check_root
    
    # 构建
    build_app
    
    # 停止服务
    stop_service
    
    # 备份
    backup_old_version
    
    # 部署
    deploy_files
    
    # 启动服务
    start_service
    
    log_info "部署完成！"
    log_info "查看日志: sudo journalctl -u $SERVICE_NAME -f"
    log_info "查看状态: sudo systemctl status $SERVICE_NAME"
}

# 运行主函数
main

