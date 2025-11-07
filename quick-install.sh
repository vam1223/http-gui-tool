#!/bin/bash

# HTTP批量请求工具 - 一键安装脚本
# 支持自动下载和本地安装
# 作者: vam1223
# 仓库: https://github.com/vam1223/http-gui-tool

set -e

APP_NAME="HTTP批量请求工具"
APP_BUNDLE="HTTP批量请求工具.app"
INSTALL_DIR="/Applications"
GITHUB_REPO="vam1223/http-gui-tool"
ZIP_NAME="http-gui-tool-installer.zip"
TEMP_DIR="/tmp/http-gui-install-$$"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为 macOS 系统
check_macos() {
    if [[ "$OSTYPE" != "darwin"* ]]; then
        print_error "此安装脚本仅适用于 macOS 系统"
        exit 1
    fi
}

# 检查必要工具
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        print_error "需要 curl 工具来下载文件"
        exit 1
    fi
    
    if ! command -v unzip &> /dev/null; then
        print_error "需要 unzip 工具来解压文件"
        exit 1
    fi
}

# 创建临时目录
create_temp_dir() {
    mkdir -p "$TEMP_DIR"
    print_info "创建临时目录: $TEMP_DIR"
}

# 清理临时文件
cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
        print_info "清理临时文件"
    fi
}

# 设置清理陷阱
trap cleanup EXIT

# 下载安装包
download_package() {
    print_info "正在从 GitHub 下载安装包..."
    
    # 尝试从 GitHub 主分支下载预构建包
    DOWNLOAD_URL="https://raw.githubusercontent.com/${GITHUB_REPO}/main/${ZIP_NAME}"
    if curl -fL "$DOWNLOAD_URL" -o "${TEMP_DIR}/${ZIP_NAME}" 2>/dev/null; then
        print_success "从 GitHub 主分支下载预构建包成功"
        cd "$TEMP_DIR"
        # 设置正确的编码环境变量
        export LC_ALL=en_US.UTF-8
        export LANG=en_US.UTF-8
        if unzip -q "${ZIP_NAME}"; then
            print_success "解压成功"
            return 0
        else
            print_error "解压失败"
            exit 1
        fi
    fi
    
    print_warning "GitHub 主分支中未找到预构建包"
    print_error "请联系开发者上传预构建的安装包到 GitHub"
    print_info "或者手动下载项目并运行本地安装脚本"
    print_info "GitHub 项目地址: https://github.com/${GITHUB_REPO}"
    exit 1
}

# 检查本地是否有zip包
check_local_package() {
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    
    # 检查当前目录是否有zip包
    if [ -f "${SCRIPT_DIR}/${ZIP_NAME}" ]; then
        print_info "发现本地安装包: ${ZIP_NAME}"
        cp "${SCRIPT_DIR}/${ZIP_NAME}" "${TEMP_DIR}/"
        cd "$TEMP_DIR"
        # 设置正确的编码环境变量
        export LC_ALL=en_US.UTF-8
        export LANG=en_US.UTF-8
        if unzip -q "${ZIP_NAME}"; then
            print_success "本地安装包解压成功"
            return 0
        else
            print_warning "本地安装包解压失败，尝试其他方式"
        fi
    fi
    
    # 检查是否已经在项目目录中
    if [ -d "${SCRIPT_DIR}/${APP_BUNDLE}" ]; then
        print_info "发现本地应用程序包"
        cp -R "${SCRIPT_DIR}/installer" "${TEMP_DIR}/"
        return 0
    fi
    
    return 1
}

# 检查权限
check_permissions() {
    if [ ! -w "$INSTALL_DIR" ]; then
        print_warning "需要管理员权限来安装到 ${INSTALL_DIR}"
        print_info "请输入管理员密码..."
        sudo -v
        if [ $? -ne 0 ]; then
            print_error "需要管理员权限才能继续安装"
            exit 1
        fi
    fi
}

# 卸载旧版本
uninstall_old_version() {
    if [ -d "${INSTALL_DIR}/${APP_BUNDLE}" ]; then
        print_warning "检测到已安装的旧版本"
        read -p "是否要替换现有版本？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "正在移除旧版本..."
            if [ -w "$INSTALL_DIR" ]; then
                rm -rf "${INSTALL_DIR}/${APP_BUNDLE}"
            else
                sudo rm -rf "${INSTALL_DIR}/${APP_BUNDLE}"
            fi
            print_success "旧版本已移除"
        else
            print_info "安装已取消"
            exit 0
        fi
    fi
}

# 安装应用程序
install_app() {
    print_info "正在安装 ${APP_NAME}..."
    
    # 查找应用程序包 - 使用find命令支持中文路径
    APP_SOURCE=""
    
    # 方法1: 在installer/payload/Applications目录下查找.app包
    if [ -d "${TEMP_DIR}/installer/payload/Applications" ]; then
        APP_SOURCE=$(find "${TEMP_DIR}/installer/payload/Applications" -name "*.app" -type d | head -1)
    fi
    
    # 方法2: 如果方法1没找到，在整个临时目录中查找.app包
    if [ -z "$APP_SOURCE" ]; then
        APP_SOURCE=$(find "${TEMP_DIR}" -name "*.app" -type d | head -1)
    fi
    
    if [ -z "$APP_SOURCE" ]; then
        print_error "找不到应用程序包"
        print_info "解压后的目录结构:"
        find "${TEMP_DIR}" -type d -maxdepth 3 | head -20
        exit 1
    fi
    
    print_info "找到应用程序包: $APP_SOURCE"
    
    # 复制应用程序
    if [ -w "$INSTALL_DIR" ]; then
        cp -R "$APP_SOURCE" "${INSTALL_DIR}/"
    else
        sudo cp -R "$APP_SOURCE" "${INSTALL_DIR}/"
    fi
    
    # 设置正确的权限
    if [ ! -w "$INSTALL_DIR" ]; then
        sudo chown -R root:admin "${INSTALL_DIR}/${APP_BUNDLE}"
        sudo chmod -R 755 "${INSTALL_DIR}/${APP_BUNDLE}"
    fi
}

# 验证安装
verify_installation() {
    if [ -d "${INSTALL_DIR}/${APP_BUNDLE}" ] && [ -x "${INSTALL_DIR}/${APP_BUNDLE}/Contents/MacOS"/* ]; then
        print_success "安装成功！"
        return 0
    else
        print_error "安装验证失败"
        return 1
    fi
}

# 主安装流程
main() {
    echo "=================================="
    echo "    ${APP_NAME} 一键安装程序"
    echo "=================================="
    echo
    
    print_info "开始安装 ${APP_NAME}..."
    
    # 执行检查
    check_macos
    check_dependencies
    check_permissions
    create_temp_dir
    
    # 卸载旧版本
    uninstall_old_version
    
    # 获取安装包
    if ! check_local_package; then
        download_package
    fi
    
    # 安装应用程序
    install_app
    
    # 验证安装
    if verify_installation; then
        print_success "${APP_NAME} 已成功安装到 ${INSTALL_DIR}"
        print_info "您可以在 Launchpad 或 Applications 文件夹中找到应用程序"
        
        echo
        print_info "应用程序功能："
        echo "  • HTTP批量请求：支持高并发请求"
        echo "  • CSV数据驱动：从CSV文件读取参数"
        echo "  • 参数映射：灵活配置请求参数"
        echo "  • 进度显示：实时显示处理进度"
        echo "  • 结果统计：显示成功和错误统计"
        echo "  • 重试机制：自动重试失败的请求"
        echo
        print_success "安装完成！享受使用 ${APP_NAME}！"
    else
        print_error "安装失败，请检查错误信息并重试"
        exit 1
    fi
}

# 运行主程序
main "$@"