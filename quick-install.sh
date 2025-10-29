#!/bin/bash

# HTTP批量请求工具 - 智能安装脚本
# 自动检测系统架构并下载对应安装包
# 作者: vam1223
# 仓库: https://github.com/vam1223/http-gui-tool

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目信息
REPO_OWNER="vam1223"
REPO_NAME="http-gui-tool"
GITHUB_API_URL="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}"
APP_NAME="HTTP批量请求工具"

# 打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_header() {
    echo ""
    print_message $CYAN "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    print_message $PURPLE "🚀 ${APP_NAME} - 智能安装脚本"
    print_message $CYAN "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
}

print_step() {
    local step=$1
    local message=$2
    print_message $BLUE "[$step] $message"
}

print_success() {
    print_message $GREEN "✅ $1"
}

print_warning() {
    print_message $YELLOW "⚠️  $1"
}

print_error() {
    print_message $RED "❌ $1"
}

# 检测系统信息
detect_system() {
    print_step "1/6" "检测系统信息..."
    
    OS=$(uname -s)
    ARCH=$(uname -m)
    
    case $OS in
        Darwin)
            PLATFORM="macOS"
            ;;
        Linux)
            PLATFORM="Linux"
            ;;
        *)
            print_error "不支持的操作系统: $OS"
            exit 1
            ;;
    esac
    
    case $ARCH in
        x86_64)
            ARCHITECTURE="amd64"
            ;;
        arm64|aarch64)
            ARCHITECTURE="arm64"
            ;;
        *)
            print_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
    
    print_success "检测到系统: $PLATFORM $ARCHITECTURE"
}

# 检查依赖
check_dependencies() {
    print_step "2/6" "检查系统依赖..."
    
    # 检查curl
    if ! command -v curl &> /dev/null; then
        print_error "curl未安装，请先安装curl"
        exit 1
    fi
    
    # 检查unzip
    if ! command -v unzip &> /dev/null; then
        print_error "unzip未安装，请先安装unzip"
        exit 1
    fi
    
    print_success "系统依赖检查完成"
}

# 获取最新版本信息
get_latest_release() {
    print_step "3/6" "获取最新版本信息..."
    
    # 尝试从GitHub API获取最新release信息
    LATEST_RELEASE=$(curl -s "${GITHUB_API_URL}/releases/latest" 2>/dev/null || echo "")
    
    if [ -z "$LATEST_RELEASE" ] || echo "$LATEST_RELEASE" | grep -q "API rate limit exceeded"; then
        print_warning "无法从GitHub API获取版本信息，使用默认下载方式"
        VERSION="latest"
        DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/archive/refs/heads/main.zip"
    else
        VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        
        if [ -z "$VERSION" ]; then
            print_warning "无法解析版本信息，使用默认下载方式"
            VERSION="latest"
            DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/archive/refs/heads/main.zip"
        else
            # 构建下载URL - 根据系统选择合适的资源
            if [ "$PLATFORM" = "macOS" ]; then
                # 对于macOS，下载源码包含预编译的.app文件
                DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/archive/refs/heads/main.zip"
            else
                # 对于其他系统，尝试下载release资源
                DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/archive/refs/heads/main.zip"
            fi
        fi
    fi
    
    print_success "版本: $VERSION"
}

# 下载安装包
download_package() {
    print_step "4/6" "下载安装包..."
    
    TEMP_DIR=$(mktemp -d)
    DOWNLOAD_FILE="$TEMP_DIR/http-gui-tool.zip"
    
    print_message $CYAN "📥 下载地址: $DOWNLOAD_URL"
    
    if curl -L --progress-bar "$DOWNLOAD_URL" -o "$DOWNLOAD_FILE"; then
        print_success "下载完成"
    else
        print_error "下载失败"
        cleanup
        exit 1
    fi
    
    # 解压文件
    cd "$TEMP_DIR"
    if unzip -q "$DOWNLOAD_FILE"; then
        print_success "解压完成"
    else
        print_error "解压失败"
        cleanup
        exit 1
    fi
    
    # 查找解压后的目录
    EXTRACTED_DIR=$(find . -maxdepth 1 -type d -name "*http-gui-tool*" | head -1)
    if [ -z "$EXTRACTED_DIR" ]; then
        print_error "找不到解压后的项目目录"
        cleanup
        exit 1
    fi
    
    SOURCE_DIR="$TEMP_DIR/$EXTRACTED_DIR"
}

# 安装应用程序
install_application() {
    print_step "5/6" "安装应用程序..."
    
    case $PLATFORM in
        macOS)
            install_macos
            ;;
        Linux)
            install_linux
            ;;
        *)
            print_error "不支持的平台: $PLATFORM"
            exit 1
            ;;
    esac
}

# macOS安装
install_macos() {
    # 检查Applications目录
    if [ ! -d "/Applications" ]; then
        print_error "找不到Applications目录"
        exit 1
    fi
    
    # 检查是否存在预编译的.app文件
    APP_PATH="$SOURCE_DIR/installer/payload/Applications/${APP_NAME}.app"
    
    if [ -d "$APP_PATH" ]; then
        # 使用预编译的.app文件
        print_message $CYAN "📦 使用预编译应用程序..."
        
        # 删除旧版本
        if [ -d "/Applications/${APP_NAME}.app" ]; then
            print_message $YELLOW "🗑️  删除旧版本..."
            rm -rf "/Applications/${APP_NAME}.app"
        fi
        
        # 复制应用程序
        cp -R "$APP_PATH" "/Applications/"
        
        # 设置权限
        chmod +x "/Applications/${APP_NAME}.app/Contents/MacOS/"*
        
    else
        # 编译安装
        print_message $CYAN "🔨 编译并安装应用程序..."
        
        # 检查Go环境
        if ! command -v go &> /dev/null; then
            print_error "Go未安装，请先安装Go语言环境"
            print_message $CYAN "💡 安装Go: brew install go"
            exit 1
        fi
        
        # 进入源码目录
        cd "$SOURCE_DIR"
        
        # 编译应用程序
        if go build -o "http-gui-tool" main.go; then
            print_success "编译成功"
        else
            print_error "编译失败"
            exit 1
        fi
        
        # 创建.app结构（简化版本）
        mkdir -p "/Applications/${APP_NAME}.app/Contents/MacOS"
        cp "http-gui-tool" "/Applications/${APP_NAME}.app/Contents/MacOS/"
        chmod +x "/Applications/${APP_NAME}.app/Contents/MacOS/http-gui-tool"
    fi
    
    # 安装macOS依赖
    install_macos_dependencies
    
    print_success "macOS安装完成"
}

# 安装macOS依赖
install_macos_dependencies() {
    print_message $CYAN "🔍 检查macOS依赖..."
    
    # 检查Homebrew
    if ! command -v brew &> /dev/null; then
        print_warning "未检测到Homebrew，正在安装..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        
        # 添加到PATH
        if [[ $(uname -m) == "arm64" ]]; then
            echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
            eval "$(/opt/homebrew/bin/brew shellenv)"
        else
            echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zprofile
            eval "$(/usr/local/bin/brew shellenv)"
        fi
    fi
    
    # 安装必要的依赖库
    for dep in leptonica tesseract; do
        if ! brew list $dep &> /dev/null; then
            print_message $CYAN "📦 安装 $dep..."
            brew install $dep
        else
            print_success "$dep 已安装"
        fi
    done
}

# Linux安装
install_linux() {
    print_message $CYAN "🔨 编译并安装Linux版本..."
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        print_error "Go未安装，请先安装Go语言环境"
        exit 1
    fi
    
    # 进入源码目录并编译
    cd "$SOURCE_DIR"
    
    if go build -o "http-gui-tool" main.go; then
        print_success "编译成功"
    else
        print_error "编译失败"
        exit 1
    fi
    
    # 安装到系统路径
    INSTALL_DIR="/usr/local/bin"
    if [ -w "$INSTALL_DIR" ]; then
        cp "http-gui-tool" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/http-gui-tool"
    else
        sudo cp "http-gui-tool" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/http-gui-tool"
    fi
    
    print_success "Linux安装完成"
}

# 验证安装
verify_installation() {
    print_step "6/6" "验证安装..."
    
    case $PLATFORM in
        macOS)
            if [ -d "/Applications/${APP_NAME}.app" ]; then
                print_success "应用程序已安装到: /Applications/${APP_NAME}.app"
                
                # 测试启动
                print_message $CYAN "🧪 测试应用程序启动..."
                timeout 5 "/Applications/${APP_NAME}.app/Contents/MacOS/http-gui-tool" &> /dev/null &
                TEST_PID=$!
                sleep 2
                
                if kill -0 $TEST_PID 2>/dev/null; then
                    print_success "应用程序启动测试成功"
                    kill $TEST_PID 2>/dev/null
                else
                    print_warning "应用程序启动测试未完成，但安装已完成"
                fi
                
                # 打开Applications文件夹
                open "/Applications" 2>/dev/null || true
            else
                print_error "安装验证失败"
                exit 1
            fi
            ;;
        Linux)
            if command -v http-gui-tool &> /dev/null; then
                print_success "应用程序已安装，可通过命令 'http-gui-tool' 启动"
            else
                print_error "安装验证失败"
                exit 1
            fi
            ;;
    esac
}

# 清理临时文件
cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

# 显示完成信息
show_completion() {
    echo ""
    print_message $GREEN "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    print_message $GREEN "🎉 ${APP_NAME} 安装完成！"
    print_message $GREEN "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    case $PLATFORM in
        macOS)
            print_message $CYAN "📱 应用程序位置: /Applications/${APP_NAME}.app"
            print_message $CYAN "🔍 您可以在以下位置找到应用程序："
            print_message $CYAN "   • Applications文件夹"
            print_message $CYAN "   • Launchpad"
            print_message $CYAN "   • Spotlight搜索（搜索'HTTP'）"
            ;;
        Linux)
            print_message $CYAN "🖥️  启动命令: http-gui-tool"
            ;;
    esac
    
    echo ""
    print_message $PURPLE "💡 如果遇到问题，请访问: https://github.com/${REPO_OWNER}/${REPO_NAME}"
    print_message $PURPLE "📧 反馈问题: https://github.com/${REPO_OWNER}/${REPO_NAME}/issues"
    echo ""
}

# 主函数
main() {
    # 捕获退出信号，确保清理临时文件
    trap cleanup EXIT
    
    print_header
    
    detect_system
    check_dependencies
    get_latest_release
    download_package
    install_application
    verify_installation
    
    show_completion
}

# 运行主函数
main "$@"