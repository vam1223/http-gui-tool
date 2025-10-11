# HTTP 批量请求工具

一个功能强大的图形化 HTTP 批量请求工具，基于 Go 语言和 Fyne 框架开发，支持 CSV 数据驱动、QPS 限流、重试机制等高级功能。

## 🌟 功能特性

### 核心功能
- **批量 HTTP 请求**：支持基于 CSV 数据的批量请求发送
- **QPS 限流控制**：可配置每秒请求数量，避免服务器过载
- **智能重试机制**：支持失败请求自动重试，可配置重试次数
- **实时日志显示**：实时显示请求进度和结果
- **配置保存/加载**：支持配置文件的保存和加载

### 高级特性
- **参数映射系统**：灵活的 CSV 列到请求参数的映射配置
- **多种参数模式**：
  - **对象模式**：将 CSV 数据映射为 JSON 对象
  - **数组模式**：将 CSV 数据映射为 JSON 数组
- **多服务器支持**：支持配置多个目标服务器 IP 和端口
- **Cookie 管理**：支持自定义 Cookie 设置
- **请求模板**：支持自定义请求体模板

## 🚀 快速开始

### 环境要求
- Go 1.21 或更高版本
- macOS / Windows / Linux 操作系统

### 安装步骤

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd http-gui-tool
   ```

2. **安装依赖**
   ```bash
   go mod tidy
   ```

3. **运行应用**
   
   **macOS (推荐方式):**
   ```bash
   # 使用启动脚本（自动设置中文字体）
   ./run-with-font.sh
   ```
   
   **手动运行:**
   ```bash
   # 设置中文字体环境变量
   export FYNE_FONT="/System/Library/Fonts/PingFang.ttc"
   go run main.go
   ```

4. **编译可执行文件**
   ```bash
   # 编译
   go build -o http-tool main.go
   
   # 运行编译后的程序
   ./http-tool
   ```

## 📋 使用指南

### 1. 准备 CSV 数据文件
创建包含测试数据的 CSV 文件，例如：

```csv
userId,userName,email,age
1001,张三,zhangsan@example.com,25
1002,李四,lisi@example.com,30
1003,王五,wangwu@example.com,28
```

### 2. 配置参数映射
在工具界面中设置参数映射：

- **CSV 列**：CSV 文件中的列索引（从0开始）或列名
- **参数名**：请求参数名称（对象模式）或数组索引（数组模式）
- **参数类型**：支持 string、int、float、bool、string[]、int[] 等类型
- **默认值**：当 CSV 列为空时使用的默认值

### 3. 配置请求参数
- **URL**：目标 API 接口地址
- **Cookie**：身份认证 Cookie
- **请求体模板**：包含 `${jsonParam}` 占位符的 JSON 模板
- **IP列表**：目标服务器地址列表，每行一个
- **QPS**：每秒请求数量限制
- **并发数**：同时执行的请求数量
- **重试次数**：失败请求的重试次数

### 4. 启动批量请求
点击"开始执行"按钮，工具将自动：
1. 读取 CSV 文件
2. 根据映射规则生成请求参数
3. 向配置的 IP 地址发送请求
4. 实时显示执行进度和结果

## 🛠️ 配置文件格式

### 示例配置文件 (`config.json`)
```json
{
  "url": "https://api.example.com/v1/test",
  "cookie": "your-cookie-here",
  "bodyTemp": "{\"key\":\"your-api-key\",\"jsonParam\":\"${jsonParam}\"}",
  "ipList": [
    "127.0.0.1:8080",
    "192.168.1.100:8080"
  ],
  "qps": 10,
  "workers": 5,
  "maxRetries": 3,
  "paramMappings": [
    {
      "csvColumn": "0",
      "paramName": "userId",
      "paramType": "int",
      "defaultValue": "1001",
      "arrayIndex": 0
    },
    {
      "csvColumn": "1",
      "paramName": "userName",
      "paramType": "string",
      "defaultValue": "default_user",
      "arrayIndex": 1
    }
  ],
  "paramMode": "array"
}
```

## 📁 项目结构

```
http-gui-tool/
├── main.go                 # 主程序文件
├── go.mod                  # Go 模块定义
├── go.sum                  # 依赖版本锁定
├── README.md              # 项目文档
├── QUICKSTART.md          # 快速开始指南
├── example-config.json    # 配置文件示例
├── install.sh             # 安装脚本
├── check.sh               # 检查脚本
├── run-with-font.sh       # 字体设置脚本
└── installer/             # macOS 安装包配置
    ├── Distribution.xml
    └── payload/
        └── Applications/
            └── HTTP批量请求工具.app/
```

## 🔧 开发说明

### 技术栈
- **后端**: Go 1.21+
- **GUI框架**: Fyne v2.6.3
- **构建工具**: Go Modules

### 构建安装包 (macOS)
```bash
# 构建应用
go build -o http-gui-tool main.go

# 创建安装包
# 使用 installer 目录中的配置创建 .pkg 安装包
```

### 字体配置
为了正确显示中文界面，应用会自动设置以下字体：
- macOS: PingFang SC 或 Arial Unicode
- Windows: 微软雅黑
- Linux: 文泉驿微米黑

## 🐛 常见问题

### Q: 中文显示乱码怎么办？
A: 使用提供的启动脚本 `run-with-font.sh`，它会自动设置中文字体环境变量。

### Q: 如何配置多个服务器？
A: 在"IP列表"中每行输入一个服务器地址，格式为 `IP:端口`。

### Q: CSV文件格式有什么要求？
A: 支持标准CSV格式，可以使用Excel或文本编辑器创建。第一行可以是列名（可选）。

### Q: 如何查看详细的请求日志？
A: 应用界面会实时显示请求进度，包括成功/失败状态和响应信息。

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个项目！

## 📞 联系方式

如有问题或建议，请通过以下方式联系：
- 提交 GitHub Issue
- 发送邮件至项目维护者

---

**版本**: v1.0.0  
**最后更新**: 2025年10月11日