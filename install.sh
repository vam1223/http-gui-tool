#!/bin/bash

echo "🚀 正在安装HTTP批量请求工具..."

# 检查Applications目录
if [ ! -d "/Applications" ]; then
    echo "❌ 错误：找不到Applications目录"
    exit 1
fi

# 检查并安装Homebrew
echo "🔍 检查Homebrew..."
if ! command -v brew &> /dev/null; then
    echo "⚠️  未检测到Homebrew，正在安装..."
    echo "📥 下载并安装Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    
    # 添加Homebrew到PATH（针对Apple Silicon Mac）
    if [[ $(uname -m) == "arm64" ]]; then
        echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
        eval "$(/opt/homebrew/bin/brew shellenv)"
    else
        echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zprofile
        eval "$(/usr/local/bin/brew shellenv)"
    fi
    
    if command -v brew &> /dev/null; then
        echo "✅ Homebrew安装成功！"
    else
        echo "❌ Homebrew安装失败，请手动安装后重试"
        echo "💡 访问 https://brew.sh 获取安装说明"
        exit 1
    fi
else
    echo "✅ Homebrew已安装"
fi

# 检查并安装依赖库
echo "🔍 检查应用程序依赖..."

# 检查leptonica
if ! brew list leptonica &> /dev/null; then
    echo "📦 安装leptonica依赖库..."
    brew install leptonica
    if [ $? -eq 0 ]; then
        echo "✅ leptonica安装成功"
    else
        echo "❌ leptonica安装失败"
        exit 1
    fi
else
    echo "✅ leptonica已安装"
fi

# 检查tesseract
if ! brew list tesseract &> /dev/null; then
    echo "📦 安装tesseract依赖库..."
    brew install tesseract
    if [ $? -eq 0 ]; then
        echo "✅ tesseract安装成功"
    else
        echo "❌ tesseract安装失败"
        exit 1
    fi
else
    echo "✅ tesseract已安装"
fi

echo "🎯 所有依赖已就绪，开始安装应用程序..."

# 删除已存在的应用程序（如果有）
if [ -d "/Applications/HTTP批量请求工具.app" ]; then
    echo "🗑️  删除旧版本..."
    rm -rf "/Applications/HTTP批量请求工具.app"
fi

# 复制新的应用程序
echo "📦 复制应用程序到Applications文件夹..."
cp -R "./installer/payload/Applications/HTTP批量请求工具.app" "/Applications/"

# 检查复制是否成功
if [ -d "/Applications/HTTP批量请求工具.app" ]; then
    echo "✅ 安装成功！"
    echo ""
    echo "📱 应用程序已安装到：/Applications/HTTP批量请求工具.app"
    echo "🔍 您现在可以在以下位置找到应用程序："
    echo "   • Applications文件夹"
    echo "   • Launchpad"
    echo "   • Spotlight搜索（搜索'HTTP'）"
    echo ""
    
    # 测试应用程序是否能正常启动
    echo "🧪 测试应用程序启动..."
    timeout 5 "/Applications/HTTP批量请求工具.app/Contents/MacOS/http-gui-tool" &> /dev/null &
    TEST_PID=$!
    sleep 2
    
    if kill -0 $TEST_PID 2>/dev/null; then
        echo "✅ 应用程序启动测试成功"
        kill $TEST_PID 2>/dev/null
    else
        echo "⚠️  应用程序启动测试未完成，但安装已完成"
    fi
    
    echo ""
    echo "🎉 安装完成！双击应用程序即可使用。"
    echo "💡 如果遇到启动问题，请确保已安装所有依赖库"
    
    # 尝试刷新Finder
    killall Finder 2>/dev/null || true
    
    # 打开Applications文件夹
    open "/Applications"
    
else
    echo "❌ 安装失败！请检查权限。"
    exit 1
fi