# HTTP批量请求工具 - 一键安装指南

## 🚀 快速安装

### 方式一：一行命令安装（推荐）

```bash
# 智能安装脚本 - 自动检测系统并下载安装包
curl -fsSL https://raw.githubusercontent.com/vam1223/http-gui-tool/main/quick-install.sh | bash
```

### 方式二：手动下载安装

```bash
# 1. 下载安装脚本
wget https://raw.githubusercontent.com/vam1223/http-gui-tool/main/quick-install.sh

# 2. 设置执行权限
chmod +x quick-install.sh

# 3. 运行安装脚本
./quick-install.sh
```

## 📋 系统要求

### macOS
- **系统版本**: macOS 10.15 或更高版本
- **架构支持**: Intel (x86_64) 和 Apple Silicon (arm64)
- **依赖工具**: curl, unzip (系统自带)
- **可选依赖**: Homebrew (脚本会自动安装)

### Linux
- **系统版本**: Ubuntu 18.04+, CentOS 7+, 或其他主流发行版
- **架构支持**: x86_64, arm64
- **必需工具**: curl, unzip, Go语言环境
- **依赖安装**:
  ```bash
  # Ubuntu/Debian
  sudo apt update && sudo apt install -y curl unzip golang-go
  
  # CentOS/RHEL
  sudo yum install -y curl unzip golang
  
  # Fedora
  sudo dnf install -y curl unzip golang
  ```

## 🎯 安装过程说明

智能安装脚本会自动执行以下步骤：

1. **🔍 系统检测**: 自动识别操作系统和CPU架构
2. **📦 依赖检查**: 验证必需工具是否已安装
3. **🌐 版本获取**: 从GitHub获取最新版本信息
4. **📥 下载安装包**: 智能选择合适的安装包
5. **🔨 安装应用**: 根据系统类型执行相应安装流程
6. **✅ 验证安装**: 测试应用程序是否正常工作

## 📱 安装后使用

### macOS
安装完成后，您可以在以下位置找到应用程序：
- **Applications文件夹**: `/Applications/HTTP批量请求工具.app`
- **Launchpad**: 搜索"HTTP"
- **Spotlight**: 按 `Cmd + Space`，搜索"HTTP"

### Linux
安装完成后，可以通过命令行启动：
```bash
http-gui-tool
```

## 🛠️ 功能特性

- ✅ **批量HTTP请求**: 支持大量并发请求处理
- ✅ **CSV数据导入**: 从CSV文件批量读取请求参数
- ✅ **参数映射配置**: 灵活的参数映射和数据转换
- ✅ **实时进度监控**: 实时显示请求进度和结果
- ✅ **多IP轮询**: 支持多个服务器IP轮询请求
- ✅ **性能优化**: 异步日志处理，流畅的用户界面
- ✅ **错误重试**: 智能错误重试机制
- ✅ **结果导出**: 请求结果实时显示和导出

## 🔧 高级配置

### 自定义安装路径（macOS）
```bash
# 设置自定义安装路径
export INSTALL_PATH="/your/custom/path"
curl -fsSL https://raw.githubusercontent.com/vam1223/http-gui-tool/main/quick-install.sh | bash
```

### 离线安装
```bash
# 1. 下载项目源码
git clone https://github.com/vam1223/http-gui-tool.git
cd http-gui-tool

# 2. 运行本地安装脚本
./install.sh
```

## 🐛 故障排除

### 常见问题

**Q: 安装时提示权限不足**
```bash
# 解决方案：使用sudo运行（仅Linux）
curl -fsSL https://raw.githubusercontent.com/vam1223/http-gui-tool/main/quick-install.sh | sudo bash
```

**Q: macOS提示"无法验证开发者"**
```bash
# 解决方案：在系统偏好设置中允许应用运行
# 系统偏好设置 > 安全性与隐私 > 通用 > 允许从以下位置下载的应用
```

**Q: Go语言环境未安装（Linux）**
```bash
# Ubuntu/Debian
sudo apt install golang-go

# CentOS/RHEL
sudo yum install golang

# 或者从官网下载：https://golang.org/dl/
```

**Q: 网络连接问题**
```bash
# 使用代理
export https_proxy=http://your-proxy:port
curl -fsSL https://raw.githubusercontent.com/vam1223/http-gui-tool/main/quick-install.sh | bash
```

### 手动卸载

**macOS:**
```bash
rm -rf "/Applications/HTTP批量请求工具.app"
```

**Linux:**
```bash
sudo rm -f /usr/local/bin/http-gui-tool
```

## 📞 获取帮助

- **GitHub仓库**: https://github.com/vam1223/http-gui-tool
- **问题反馈**: https://github.com/vam1223/http-gui-tool/issues
- **功能建议**: https://github.com/vam1223/http-gui-tool/discussions

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 贡献

欢迎提交Pull Request和Issue！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解贡献指南。

---

**享受高效的HTTP批量请求体验！** 🚀