# HTTP GUI Tool

一个基于 Go 和 Fyne 框架的图形化 HTTP 请求工具。

## 中文字体支持

为了正确显示中文字符，请使用提供的启动脚本：

### 使用方法

```bash
# 使用启动脚本（推荐）
./run-with-font.sh

# 或者手动设置环境变量
export FYNE_FONT="/System/Library/Fonts/PingFang.ttc"
go run main.go
```

### 关键改进

1. **简化了代码**：移除了复杂的字符编码处理逻辑
2. **使用 FYNE_FONT 环境变量**：这是 Fyne 框架正确设置字体的标准方法
3. **使用系统字体**：PingFang SC 是 macOS 上的标准中文字体
4. **标准文件选择器**：使用 Fyne 的内置文件选择对话框

### 功能特性

- 批量 HTTP 请求处理
- CSV 文件数据导入
- QPS 限流控制
- 重试机制
- 实时日志显示
- 配置保存/加载

### 编译

```bash
go build -o http-tool main.go
```

### 运行

```bash
# 设置字体环境变量并运行
./run-with-font.sh
```

现在中文文件名和界面文字应该能够正确显示了！