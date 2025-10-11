#!/bin/bash

echo "🔍 HTTP批量请求工具 - 系统兼容性检查"
echo "=========================================="

# 获取系统信息
SYSTEM_VERSION=$(sw_vers -productVersion)
SYSTEM_BUILD=$(sw_vers -buildVersion)
HARDWARE=$(uname -m)

echo "📱 系统信息："
echo "   macOS版本: $SYSTEM_VERSION"
echo "   构建版本: $SYSTEM_BUILD"
echo "   硬件架构: $HARDWARE"
echo ""

# 检查CPU架构
echo "🔧 CPU架构检查："
if [ "$HARDWARE" = "arm64" ]; then
    echo "   ✅ Apple Silicon (M1/M2/M3) - 完全兼容"
    ARCH_COMPATIBLE=true
elif [ "$HARDWARE" = "x86_64" ]; then
    echo "   ⚠️  Intel Mac - 当前版本不兼容"
    echo "   💡 需要Intel版本的应用程序"
    ARCH_COMPATIBLE=false
else
    echo "   ❌ 未知架构: $HARDWARE"
    ARCH_COMPATIBLE=false
fi
echo ""

# 检查macOS版本
echo "🍎 macOS版本检查："
MAJOR_VERSION=$(echo $SYSTEM_VERSION | cut -d. -f1)
MINOR_VERSION=$(echo $SYSTEM_VERSION | cut -d. -f2)

if [ "$MAJOR_VERSION" -gt 10 ] || ([ "$MAJOR_VERSION" -eq 10 ] && [ "$MINOR_VERSION" -ge 11 ]); then
    echo "   ✅ macOS $SYSTEM_VERSION - 版本兼容"
    OS_COMPATIBLE=true
else
    echo "   ❌ macOS $SYSTEM_VERSION - 版本过低"
    echo "   💡 需要macOS 10.11或更高版本"
    OS_COMPATIBLE=false
fi
echo ""

# 检查Rosetta 2（如果是Intel Mac）
if [ "$HARDWARE" = "x86_64" ]; then
    echo "🔄 Rosetta 2检查："
    if /usr/bin/pgrep oahd >/dev/null 2>&1; then
        echo "   ✅ Rosetta 2已安装"
    else
        echo "   ⚠️  Rosetta 2未安装"
        echo "   💡 虽然已安装，但当前应用程序仍不兼容Intel Mac"
    fi
    echo ""
fi

# 检查安全设置
echo "🔒 安全设置检查："
GATEKEEPER_STATUS=$(spctl --status 2>/dev/null)
if echo "$GATEKEEPER_STATUS" | grep -q "assessments enabled"; then
    echo "   ⚠️  Gatekeeper已启用"
    echo "   💡 首次运行时可能需要手动允许应用程序"
else
    echo "   ℹ️  Gatekeeper已禁用"
fi
echo ""

# 检查网络连接
echo "🌐 网络连接检查："
if ping -c 1 www.apple.com >/dev/null 2>&1; then
    echo "   ✅ 网络连接正常"
    NETWORK_OK=true
else
    echo "   ⚠️  网络连接异常"
    echo "   💡 请检查网络设置"
    NETWORK_OK=false
fi
echo ""

# 检查Applications目录权限
echo "📁 安装权限检查："
if [ -w "/Applications" ]; then
    echo "   ✅ 具有Applications目录写入权限"
    INSTALL_OK=true
else
    echo "   ⚠️  缺少Applications目录写入权限"
    echo "   💡 安装时可能需要管理员权限"
    INSTALL_OK=false
fi
echo ""

# 综合兼容性评估
echo "📊 兼容性评估："
echo "=========================================="

if [ "$ARCH_COMPATIBLE" = true ] && [ "$OS_COMPATIBLE" = true ]; then
    echo "🎉 系统完全兼容！"
    echo ""
    echo "✅ 可以正常安装和运行HTTP批量请求工具"
    echo ""
    echo "📋 安装步骤："
    echo "   1. 运行: ./install.sh"
    echo "   2. 在Applications文件夹中找到应用程序"
    echo "   3. 首次运行时允许安全提示"
    OVERALL_COMPATIBLE=true
elif [ "$ARCH_COMPATIBLE" = false ] && [ "$OS_COMPATIBLE" = true ]; then
    echo "⚠️  部分兼容 - 需要特殊版本"
    echo ""
    echo "❌ 当前应用程序不支持Intel Mac"
    echo "✅ macOS版本符合要求"
    echo ""
    echo "💡 解决方案："
    echo "   • 请联系开发者获取Intel版本"
    echo "   • 或者使用支持Apple Silicon的Mac"
    OVERALL_COMPATIBLE=false
elif [ "$ARCH_COMPATIBLE" = true ] && [ "$OS_COMPATIBLE" = false ]; then
    echo "⚠️  部分兼容 - 需要系统升级"
    echo ""
    echo "✅ 硬件架构支持"
    echo "❌ macOS版本过低"
    echo ""
    echo "💡 解决方案："
    echo "   • 升级macOS到10.11或更高版本"
    echo "   • 建议升级到最新版本以获得最佳体验"
    OVERALL_COMPATIBLE=false
else
    echo "❌ 系统不兼容"
    echo ""
    echo "❌ 硬件架构不支持"
    echo "❌ macOS版本不符合要求"
    echo ""
    echo "💡 解决方案："
    echo "   • 使用Apple Silicon Mac"
    echo "   • 升级macOS到10.11或更高版本"
    echo "   • 或联系开发者获取兼容版本"
    OVERALL_COMPATIBLE=false
fi

echo ""
echo "=========================================="

# 提供额外建议
if [ "$OVERALL_COMPATIBLE" = true ]; then
    echo "🔧 使用建议："
    echo "   • 确保网络连接稳定"
    echo "   • 首次运行时注意安全提示"
    echo "   • 如遇问题，请查看分发说明.md"
    
    if [ "$NETWORK_OK" = false ]; then
        echo "   ⚠️  请先解决网络连接问题"
    fi
    
    if [ "$INSTALL_OK" = false ]; then
        echo "   ⚠️  安装时可能需要输入管理员密码"
    fi
else
    echo "🆘 获取帮助："
    echo "   • 查看完整的分发说明.md文档"
    echo "   • 联系技术支持获取兼容版本"
    echo "   • 考虑使用兼容的硬件和系统"
fi

echo ""
echo "检查完成！"

# 返回兼容性状态
if [ "$OVERALL_COMPATIBLE" = true ]; then
    exit 0
else
    exit 1
fi