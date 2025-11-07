#!/bin/bash

# HTTP批量请求工具 - macOS一键安装脚本
# 直接从GitHub下载预编译二进制文件并安装
# 作者: vam1223
# 仓库: https://github.com/vam1223/http-gui-tool

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
REPO_OWNER="vam1223"
REPO_NAME="http-gui-tool"
APP_NAME="HTTP批量请求工具"
APP_BUNDLE_NAME="HTTP批量请求工具.app"

# 简洁的消息输出
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 检查是否为macOS
check_macos() {
    if [[ "$(uname)" != "Darwin" ]]; then
        log_error "此脚本仅支持macOS系统"
        exit 1
    fi
    
    # 检查架构
    ARCH=$(uname -m)
    if [[ "$ARCH" != "x86_64" && "$ARCH" != "arm64" ]]; then
        log_error "不支持的架构: $ARCH，仅支持x86_64和arm64"
        exit 1
    fi
    
    log_info "系统: macOS $ARCH"
}

# 检查网络连接
check_network() {
    log_info "检查网络连接..."
    if ! curl -s --max-time 5 "https://api.github.com" > /dev/null; then
        log_error "无法连接到GitHub，请检查网络连接"
        exit 1
    fi
    log_success "网络连接正常"
}

# 获取最新版本
get_latest_version() {
    log_info "获取最新版本信息..."
    
    # 尝试从GitHub API获取最新版本
    LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" 2>/dev/null || echo "")
    
    if [ -n "$LATEST_RELEASE" ] && ! echo "$LATEST_RELEASE" | grep -q "API rate limit exceeded"; then
        VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -n "$VERSION" ]; then
            log_success "最新版本: $VERSION"
            return 0
        fi
    fi
    
    # 如果API失败，使用main分支
    log_warning "无法获取最新版本，使用main分支"
    VERSION="main"
    return 0
}

# 下载预编译二进制文件
download_binary() {
    log_info "下载预编译二进制文件..."
    
    # 构建下载URL
    if [ "$VERSION" = "main" ]; then
        # 使用main分支的最新构建
        DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/raw/main/installer/payload/Applications/${APP_BUNDLE_NAME}/Contents/MacOS/http-gui-tool"
    else
        # 使用特定版本的构建（如果有releases）
        DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/http-gui-tool-darwin-${ARCH}"
    fi
    
    TEMP_DIR=$(mktemp -d)
    BINARY_PATH="$TEMP_DIR/http-gui-tool"
    
    log_info "下载地址: $DOWNLOAD_URL"
    
    # 下载二进制文件
    if curl -L -f --progress-bar --max-time 60 "$DOWNLOAD_URL" -o "$BINARY_PATH"; then
        chmod +x "$BINARY_PATH"
        log_success "二进制文件下载成功"
        return 0
    else
        log_error "二进制文件下载失败"
        return 1
    fi
}

# 创建macOS应用程序包
create_app_bundle() {
    log_info "创建macOS应用程序包..."
    
    APP_DIR="/Applications/${APP_BUNDLE_NAME}"
    CONTENTS_DIR="$APP_DIR/Contents"
    MACOS_DIR="$CONTENTS_DIR/MacOS"
    RESOURCES_DIR="$CONTENTS_DIR/Resources"
    
    # 删除旧版本
    if [ -d "$APP_DIR" ]; then
        log_info "删除旧版本..."
        rm -rf "$APP_DIR"
    fi
    
    # 创建目录结构
    mkdir -p "$MACOS_DIR" "$RESOURCES_DIR"
    
    # 复制可执行文件
    cp "$BINARY_PATH" "$MACOS_DIR/http-gui-tool"
    chmod +x "$MACOS_DIR/http-gui-tool"
    
    # 创建Info.plist
    cat > "$CONTENTS_DIR/Info.plist" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>http-gui-tool</string>
    <key>CFBundleIdentifier</key>
    <string>com.vam1223.http-gui-tool</string>
    <key>CFBundleName</key>
    <string>HTTP批量请求工具</string>
    <key>CFBundleDisplayName</key>
    <string>HTTP批量请求工具</string>
    <key>CFBundleVersion</key>
    <string>1.1.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.1.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleSignature</key>
    <string>????</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.12</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF
    
    log_success "应用程序包创建完成"
}

# 验证安装
verify_installation() {
    log_info "验证安装..."
    
    if [ -d "/Applications/${APP_BUNDLE_NAME}" ] && [ -x "/Applications/${APP_BUNDLE_NAME}/Contents/MacOS/http-gui-tool" ]; then
        log_success "安装验证成功"
        return 0
    else
        log_error "安装验证失败"
        return 1
    fi
}

# 显示完成信息
show_completion() {
    echo ""
    echo "🎉 安装完成！"
    echo "=================================="
    echo ""
    echo "📱 应用程序位置: /Applications/${APP_BUNDLE_NAME}"
    echo "🔍 打开方式:"
    echo "   • Applications文件夹"
    echo "   • Launchpad"
    echo "   • Spotlight搜索（搜索'HTTP'）"
    echo ""
    echo "💡 首次使用建议:"
    echo "   • 设置QPS为10-25，避免过高频率"
    echo "   • 设置Workers为50-100，根据机器性能"
    echo "   • 使用测试数据进行功能验证"
    echo ""
    echo "📧 问题反馈: https://github.com/${REPO_OWNER}/${REPO_NAME}/issues"
    echo ""
}

# 清理临时文件
cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

# 错误处理
handle_error() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        log_error "安装过程中出现错误（退出码: $exit_code）"
        log_info "请检查网络连接和系统权限"
        log_info "如需帮助，请访问: https://github.com/${REPO_OWNER}/${REPO_NAME}/issues"
    fi
    cleanup
    exit $exit_code
}

# 主函数
main() {
    # 设置错误处理
    trap handle_error EXIT
    
    echo ""
    echo "🚀 ${APP_NAME} - macOS一键安装"
    echo "=================================="
    echo ""
    
    # 检查系统
    check_macos
    
    # 检查网络
    check_network
    
    # 获取版本
    get_latest_version
    
    # 下载二进制文件
    if download_binary; then
        # 创建应用程序包
        create_app_bundle
        
        # 验证安装
        if verify_installation; then
            show_completion
        else
            log_error "安装验证失败"
            exit 1
        fi
    else
        log_error "安装失败，无法下载二进制文件"
        log_info "可能的原因:"
        log_info "• GitHub Releases中暂无预编译文件"
        log_info "• 网络连接问题"
        log_info "• 权限不足"
        exit 1
    fi
}

# 运行主函数
main "$@"