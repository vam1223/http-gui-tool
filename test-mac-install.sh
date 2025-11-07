#!/bin/bash

# 测试macOS一键安装脚本的核心功能

echo "🚀 测试macOS一键安装脚本"
echo "=================================="
echo ""

# 检查系统
if [[ "$(uname)" != "Darwin" ]]; then
    echo "❌ 此脚本仅支持macOS系统"
    exit 1
fi

echo "✅ 系统检查通过: macOS $(uname -m)"

# 检查网络
echo "🔍 检查网络连接..."
if curl -s --max-time 5 "https://api.github.com" > /dev/null; then
    echo "✅ 网络连接正常"
else
    echo "❌ 无法连接到GitHub"
    exit 1
fi

# 测试版本获取
echo "📋 测试版本获取..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/vam1223/http-gui-tool/releases/latest" 2>/dev/null || echo "")

if [ -n "$LATEST_RELEASE" ] && ! echo "$LATEST_RELEASE" | grep -q "API rate limit exceeded"; then
    VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -n "$VERSION" ]; then
        echo "✅ 最新版本: $VERSION"
    else
        echo "⚠️  无法解析版本，使用main分支"
        VERSION="main"
    fi
else
    echo "⚠️  无法获取版本信息，使用main分支"
    VERSION="main"
fi

# 测试下载URL构建
echo "🔗 测试下载URL构建..."
if [ "$VERSION" = "main" ]; then
    DOWNLOAD_URL="https://github.com/vam1223/http-gui-tool/raw/main/installer/payload/Applications/HTTP批量请求工具.app/Contents/MacOS/http-gui-tool"
else
    DOWNLOAD_URL="https://github.com/vam1223/http-gui-tool/releases/download/${VERSION}/http-gui-tool-darwin-$(uname -m)"
fi

echo "📥 下载URL: $DOWNLOAD_URL"

# 测试下载（只下载前1KB来验证链接）
echo "🧪 测试下载链接..."
TEMP_FILE=$(mktemp)
if curl -L -f -r 0-1023 --max-time 10 "$DOWNLOAD_URL" -o "$TEMP_FILE" 2>/dev/null; then
    echo "✅ 下载链接有效"
    rm -f "$TEMP_FILE"
    
    echo ""
    echo "🎉 测试完成！"
    echo "=================================="
    echo ""
    echo "💡 安装脚本可以正常运行"
    echo "📥 实际安装将下载完整的二进制文件"
    echo "📱 安装位置: /Applications/HTTP批量请求工具.app"
    echo ""
    echo "🔧 运行完整安装: ./mac-install.sh"
    
else
    echo "❌ 下载链接测试失败"
    rm -f "$TEMP_FILE"
    exit 1
fi

echo ""
echo "✅ 所有测试通过！"