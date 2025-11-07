package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	
	"fyne.io/fyne/v2/widget"
)

// 自定义HTTP客户端，带连接池和超时设置
var httpClient = &http.Client{
	Timeout: 30 * time.Second, // 请求总超时
	Transport: &http.Transport{
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxIdleConnsPerHost: 20,               // 每个主机的最大空闲连接数
		MaxConnsPerHost:     50,               // 每个主机的最大连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时
		TLSHandshakeTimeout: 10 * time.Second, // TLS握手超时
		DisableCompression:  false,            // 启用压缩
		ForceAttemptHTTP2:   true,             // 尝试使用HTTP/2
		DisableKeepAlives:   false,            // 启用连接复用
	},
}

// 字符串构建器池，减少内存分配
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// 获取字符串构建器
func getStringBuilder() *strings.Builder {
	return stringBuilderPool.Get().(*strings.Builder)
}

// 归还字符串构建器
func putStringBuilder(sb *strings.Builder) {
	sb.Reset()
	stringBuilderPool.Put(sb)
}

// 参数映射配置
type ParamMapping struct {
	CSVColumn    string `json:"csvColumn"`    // CSV列名或索引
	ParamName    string `json:"paramName"`    // 请求参数名（对象模式）或数组索引（数组模式）
	ParamType    string `json:"paramType"`    // 参数类型：string, int, float, bool
	DefaultValue string `json:"defaultValue"` // 默认值
	ArrayIndex   int    `json:"arrayIndex"`   // 数组模式下的参数位置索引
}

// 参数映射行UI组件
type ParamMappingRow struct {
	CSVColumnEntry    *widget.Entry
	ParamNameEntry    *widget.Entry
	ParamTypeSelect   *widget.Select
	DefaultValueEntry *widget.Entry
	ArrayIndexEntry   *widget.Entry
	DeleteButton      *widget.Button
	Container         *fyne.Container
}

// Config 配置结构
type Config struct {
	URL           string         `json:"url"`
	Cookie        string         `json:"cookie"`
	BodyTemp      string         `json:"bodyTemp"`
	IPList        []string       `json:"ipList"`
	QPS           int            `json:"qps"`
	Workers       int            `json:"workers"`
	MaxRetries    int            `json:"maxRetries"`
	ParamMappings []ParamMapping `json:"paramMappings"` // 参数映射配置
	ParamMode     string         `json:"paramMode"`     // 参数生成模式：object(对象) 或 array(数组)
}

// RequestTask 请求任务结构
type RequestTask struct {
	ParamsJSON []byte
	RowIndex   int
}

// HTTPTool GUI应用结构
type HTTPTool struct {
	app    fyne.App
	window fyne.Window
	config *Config

	// UI组件
	urlEntry      *widget.Entry
	cookieEntry   *widget.Entry
	bodyEntry     *widget.Entry
	ipListEntry   *widget.Entry
	qpsEntry      *widget.Entry
	workersEntry  *widget.Entry
	retriesEntry  *widget.Entry
	csvPathEntry  *widget.Entry
	outputText    *widget.Entry
	
	// 控制组件
	startBtn   *widget.Button
	stopBtn    *widget.Button
	clearBtn   *widget.Button
	saveBtn    *widget.Button
	loadBtn    *widget.Button
	
	// 状态和进度组件
	progressBar   *widget.ProgressBar
	statusLabel   *widget.Label
	
	// 参数映射相关组件
	paramModeSelect       *widget.Select
	paramMappingContainer *fyne.Container
	paramMappingScroll    *container.Scroll
	paramMappingList      []*ParamMappingRow
	
	// 运行状态
	isRunning   bool
	cancelFunc  context.CancelFunc
	mutex       sync.RWMutex
	
	// 日志缓冲 - 优化版本
	logBuffer     []string
	logMutex      sync.Mutex
	logTicker     *time.Ticker
	logChannel    chan string
	maxLogLines   int
	lastUIUpdate  time.Time
	uiUpdateMutex sync.Mutex
	
	// 进度更新优化
	lastProgressUpdate time.Time
	progressMutex      sync.Mutex
	progressChannel    chan progressUpdate
}

// 进度更新结构
type progressUpdate struct {
	processed int
	total     int
	success   int
	errors    int
}


// 创建文件选择器
func (h *HTTPTool) showCustomFileDialog() {
	// 创建自定义大小的文件选择对话框
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		
		// 获取文件路径
		filePath := reader.URI().Path()
		h.csvPathEntry.SetText(filePath)
		h.appendLog(fmt.Sprintf("Selected file: %s", filePath))
	}, h.window)
	
	// 设置对话框大小
	fileDialog.Resize(fyne.NewSize(1200, 800))
	fileDialog.Show()
}


func main() {
	// 设置中文字体支持
	os.Setenv("FYNE_FONT", "/Library/Fonts/Arial Unicode.ttf")
	
	myApp := app.NewWithID("com.example.httptool")

	tool := &HTTPTool{
		app: myApp,
		config: &Config{
			URL:        "http://intest-manager.jd.com/api/v1/invokeJsfByCallType",
			QPS:        25,
			Workers:    100,
			MaxRetries: 3,
			ParamMode:  "object", // 默认使用对象模式
		},
	}

	tool.setupUI()
	tool.loadConfig()
	tool.window.ShowAndRun()
}

func (h *HTTPTool) setupUI() {
	h.window = h.app.NewWindow("HTTP 批量请求工具 - v1.0")
	h.window.Resize(fyne.NewSize(1200, 1000))
	h.window.CenterOnScreen()

	// 创建输入字段
	h.urlEntry = widget.NewEntry()
	h.urlEntry.SetText(h.config.URL)
	h.urlEntry.MultiLine = false

	h.cookieEntry = widget.NewEntry()
	h.cookieEntry.MultiLine = true
	h.cookieEntry.Wrapping = fyne.TextWrapWord

	h.bodyEntry = widget.NewEntry()
	h.bodyEntry.MultiLine = true
	h.bodyEntry.Wrapping = fyne.TextWrapWord
	h.bodyEntry.SetText(`{"key":"6053218425615928","name":"新建JSF","type":"InterfacePage","interfaceType":"JSF","active":true,"inputParamType":"java.lang.String,java.lang.Long","interfaceName":"com.jd.o2o.settlement.BackDoorInnerService","alias":"o2o-settlement-gray","method":"recalculateSettleOrderAmount","ipPort":"%s","token":"","callType":0,"jsonParam":"${jsonParam}","serialization":"msgpack","serializerFeature":"JSON","clientType":"generic","isForceBot":0,"eoneEnv":"","overtime":"","traced":false,"mockType":0,"matchType":"0","compareRuleInfo":{"presets":[],"compareScript":[]},"lineId":23572}`)

	h.ipListEntry = widget.NewEntry()
	h.ipListEntry.MultiLine = true
	h.ipListEntry.SetText("6.19.96.149:22000\n6.19.134.55:22000\n6.40.32.10:22000\n11.63.86.240:22000\n11.134.9.63:22000")

	h.qpsEntry = widget.NewEntry()
	h.qpsEntry.SetText("25")

	h.workersEntry = widget.NewEntry()
	h.workersEntry.SetText("100")

	h.retriesEntry = widget.NewEntry()
	h.retriesEntry.SetText("3")

	h.csvPathEntry = widget.NewEntry()
	h.csvPathEntry.SetPlaceHolder("Select CSV file path...")
	
	// 初始化参数映射容器
	h.paramMappingContainer = container.NewVBox()
	h.paramMappingList = make([]*ParamMappingRow, 0)
	
	// 初始化参数模式选择器
	h.paramModeSelect = widget.NewSelect(
		[]string{"object", "array"},
		func(selected string) {
			h.config.ParamMode = selected
			h.refreshParamMappingContainer() // 刷新界面以显示不同模式的字段
		},
	)
	h.paramModeSelect.SetSelected(h.config.ParamMode)
	
	// 添加一个示例参数映射行
	exampleRow := h.createParamMappingRow()
	exampleRow.CSVColumnEntry.SetText("0")
	exampleRow.ParamNameEntry.SetText("settleOrderId")
	exampleRow.ParamTypeSelect.SetSelected("string")
	exampleRow.DefaultValueEntry.SetText("")
	if h.config.ParamMode == "array" {
		exampleRow.ArrayIndexEntry.SetText("0")
	}
	h.paramMappingList = append(h.paramMappingList, exampleRow)
	
	h.refreshParamMappingContainer()
	
	// 创建参数映射滚动容器
	h.paramMappingScroll = container.NewScroll(h.paramMappingContainer)
	h.paramMappingScroll.Resize(fyne.NewSize(0, 400))

	h.outputText = widget.NewEntry()
	h.outputText.MultiLine = true
	h.outputText.Wrapping = fyne.TextWrapWord
	
	// 初始化优化的日志缓冲系统
	h.logBuffer = make([]string, 0, 100)     // 预分配容量
	h.logChannel = make(chan string, 200)    // 异步日志通道
	h.maxLogLines = 200                      // 最大日志行数
	h.progressChannel = make(chan progressUpdate, 50) // 进度更新通道
	h.setupLogBuffer()
	h.setupProgressUpdater()

	// 创建进度条和状态标签
	h.progressBar = widget.NewProgressBar()
	h.progressBar.Hide() // 初始隐藏
	h.statusLabel = widget.NewLabel("就绪")

	// 创建按钮
	h.startBtn = widget.NewButton("▶ 开始执行", h.startExecution)
	h.startBtn.Importance = widget.HighImportance
	
	h.stopBtn = widget.NewButton("⏹ 停止执行", h.stopExecution)
	h.stopBtn.Importance = widget.DangerImportance
	h.stopBtn.Disable()
	
	h.clearBtn = widget.NewButton("🗑 清除日志", func() {
		h.outputText.SetText("")
		h.statusLabel.SetText("日志已清除")
	})
	
	h.saveBtn = widget.NewButton("💾 保存配置", h.saveConfig)
	h.loadBtn = widget.NewButton("📁 加载配置", h.loadConfigFromFile)

	// 文件选择按钮
	csvSelectBtn := widget.NewButton("📂 选择文件", func() {
		h.showCustomFileDialog()
	})

	// 设置输入字段高度
	h.cookieEntry.Resize(fyne.NewSize(0, 80))
	h.bodyEntry.Resize(fyne.NewSize(0, 80))
	h.ipListEntry.Resize(fyne.NewSize(0, 80))

	// 设置文本框固定高度
	h.cookieEntry.Resize(fyne.NewSize(0, 150))
	h.bodyEntry.Resize(fyne.NewSize(0, 150))
	h.ipListEntry.Resize(fyne.NewSize(0, 150))

	// 布局
	configForm := container.NewVBox(
		widget.NewCard("🌐 基础配置", "", container.NewVBox(
			widget.NewLabel("请求地址:"),
			h.urlEntry,
			widget.NewLabel("Cookie:"),
			h.cookieEntry,
		)),
		
		widget.NewCard("📝 请求模板", "", container.NewVBox(
			widget.NewLabel("请求体模板 (payLoad):"),
			h.bodyEntry,
		)),
		
		widget.NewCard("🖥 服务器配置", "", container.NewVBox(
			widget.NewLabel("IP地址列表 (每行一个):"),
			h.ipListEntry,
		)),
		
		widget.NewCard("⚡ 性能参数", "", container.NewGridWithColumns(3,
			widget.NewLabel("QPS:"), h.qpsEntry, widget.NewLabel("请求/秒"),
			widget.NewLabel("并发数:"), h.workersEntry, widget.NewLabel("线程"),
			widget.NewLabel("重试次数:"), h.retriesEntry, widget.NewLabel("次"),
		)),
		
		widget.NewCard("📊 数据文件", "", container.NewVBox(
			widget.NewLabel("CSV 数据文件路径:"),
			container.NewBorder(nil, nil, nil, csvSelectBtn, h.csvPathEntry),
		)),
		
		widget.NewCard("🔗 参数映射配置", "",
			container.NewVBox(
				widget.NewLabel("配置CSV列与请求参数的映射关系:"),
				container.NewGridWithColumns(2,
					widget.NewLabel("参数生成模式:"),
					h.paramModeSelect,
				),
				widget.NewSeparator(),
				func() *fyne.Container {
					// 设置参数映射容器固定高度
					h.paramMappingContainer.Resize(fyne.NewSize(0, 500))
					return h.paramMappingContainer
				}(),
			),
		),
	)

	// 主要控制按钮
	mainControlPanel := container.NewHBox(
		h.startBtn,
		h.stopBtn,
		widget.NewSeparator(),
		h.clearBtn,
	)

	// 配置管理按钮
	configPanel := container.NewHBox(
		h.saveBtn,
		h.loadBtn,
	)

	// 状态栏
	statusBar := container.NewBorder(
		nil, nil,
		widget.NewLabel("状态:"), nil,
		h.statusLabel,
	)

	// 进度区域
	progressSection := container.NewVBox(
		h.progressBar,
		statusBar,
	)

	// 创建一个更大的日志显示区域
	logCard := widget.NewCard("📋 执行日志", "", container.NewScroll(h.outputText))
	
	// 上方控制区域
	topControlsPanel := container.NewVBox(
		widget.NewCard("🎮 控制面板", "", mainControlPanel),
		widget.NewCard("⚙️ 配置管理", "", configPanel),
		widget.NewCard("📈 执行状态", "", progressSection),
	)
	
	// 右侧面板布局优化 - 使用边框布局让日志区域占用更多空间
	rightPanel := container.NewBorder(
		topControlsPanel, // 顶部：控制面板
		nil,              // 底部：无
		nil, nil,         // 左右：无
		logCard,          // 中心：日志区域（占用剩余所有空间）
	)

	leftPanel := container.NewScroll(configForm)
	
	content := container.NewHSplit(leftPanel, rightPanel)
	content.SetOffset(0.45) // 左侧占45%

	h.window.SetContent(content)
}

func (h *HTTPTool) startExecution() {
	// 验证输入
	if err := h.validateInputs(); err != nil {
		dialog.ShowError(err, h.window)
		h.statusLabel.SetText("配置错误")
		return
	}

	h.mutex.Lock()
	h.isRunning = true
	h.mutex.Unlock()

	h.startBtn.Disable()
	h.stopBtn.Enable()
	h.outputText.SetText("")
	
	// 显示进度条和更新状态
	h.progressBar.Show()
	h.progressBar.SetValue(0)
	h.statusLabel.SetText("正在准备执行...")

	// 创建取消上下文
	ctx, cancel := context.WithCancel(context.Background())
	h.cancelFunc = cancel

	go h.executeRequests(ctx)
}

func (h *HTTPTool) stopExecution() {
	h.mutex.Lock()
	
	if h.cancelFunc != nil {
		h.cancelFunc()
	}
	h.isRunning = false
	h.mutex.Unlock()
	
	// 在UI线程中更新界面
	fyne.Do(func() {
		h.startBtn.Enable()
		h.stopBtn.Disable()
		h.statusLabel.SetText("执行已停止")
		h.progressBar.Hide()
	})
	
	h.appendLog("⏹ 用户手动停止执行")
}

func (h *HTTPTool) validateInputs() error {
	if strings.TrimSpace(h.urlEntry.Text) == "" {
		return fmt.Errorf("请求URL不能为空")
	}
	if strings.TrimSpace(h.csvPathEntry.Text) == "" {
		return fmt.Errorf("请选择CSV文件")
	}
	if _, err := strconv.Atoi(h.qpsEntry.Text); err != nil {
		return fmt.Errorf("QPS必须是数字")
	}
	if _, err := strconv.Atoi(h.workersEntry.Text); err != nil {
		return fmt.Errorf("并发数必须是数字")
	}
	if _, err := strconv.Atoi(h.retriesEntry.Text); err != nil {
		return fmt.Errorf("Retries must be a number")
	}
	return nil
}

func (h *HTTPTool) executeRequests(ctx context.Context) {
	defer func() {
		h.mutex.Lock()
		h.isRunning = false
		h.mutex.Unlock()
		
		// 在UI线程中更新按钮状态
		fyne.Do(func() {
			h.startBtn.Enable()
			h.stopBtn.Disable()
		})
		
		// 确保最后的日志都被刷新
		h.flushLogBuffer()
	}()

	// 解析配置
	qps, _ := strconv.Atoi(h.qpsEntry.Text)
	workers, _ := strconv.Atoi(h.workersEntry.Text)
	maxRetries, _ := strconv.Atoi(h.retriesEntry.Text)
	
	ipList := strings.Split(strings.TrimSpace(h.ipListEntry.Text), "\n")
	for i := range ipList {
		ipList[i] = strings.TrimSpace(ipList[i])
	}

	h.appendLog(fmt.Sprintf("Starting execution - QPS: %d, Workers: %d, Retries: %d", qps, workers, maxRetries))

	// 打开CSV文件
	file, err := os.Open(h.csvPathEntry.Text)
	if err != nil {
		h.appendLog(fmt.Sprintf("Failed to open CSV file: %v", err))
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 创建请求队列和限流器 - 优化队列大小
	requestQueue := make(chan RequestTask, workers*2) // 根据worker数量调整队列大小
	rateLimiter := time.NewTicker(time.Second / time.Duration(qps))
	defer rateLimiter.Stop()

	// 创建错误通道用于收集错误信息
	errorChan := make(chan error, workers)
	var wg sync.WaitGroup

	// Start worker goroutines - 优化worker管理，增强停止响应
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					select {
					case errorChan <- fmt.Errorf("worker %d panic: %v", workerID, r):
					case <-ctx.Done():
					}
				}
			}()
			
			for {
				select {
				case <-ctx.Done():
					return // 优先检查取消信号
				case task, ok := <-requestQueue:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return // 在限流前再次检查
					case <-rateLimiter.C:
						h.sendRequest(ctx, task, ipList, maxRetries)
					}
				}
			}
		}(i)
	}

	// 先读取所有行以计算总数 - 优化内存使用
	allRows := make([][]string, 0, 1000) // 预分配容量
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}
		allRows = append(allRows, row)
	}
	
	totalRows := len(allRows) - 1 // 减去标题行
	if totalRows <= 0 {
		h.appendLog("CSV文件没有数据行")
		close(requestQueue)
		wg.Wait()
		return
	}
	
	h.appendLog(fmt.Sprintf("Found %d data rows to process", totalRows))
	
	// 处理CSV数据 - 优化处理逻辑
	rowIndex := 0
	successCount := 0
	errorCount := 0
	processedCount := 0
	batchSize := 100 // 增加批量处理大小，减少UI更新频率
	
	// 启动错误监控goroutine
	go func() {
		for err := range errorChan {
			select {
			case <-ctx.Done():
				return
			default:
				h.appendLog(fmt.Sprintf("Worker error: %v", err))
			}
		}
	}()

	// 使用更高效的循环，定期检查停止信号
	checkInterval := 10 // 每10行检查一次停止信号
	for i, row := range allRows {
		// 定期检查停止信号，避免处理过多数据
		if i%checkInterval == 0 {
			select {
			case <-ctx.Done():
				h.appendLog("Execution cancelled during processing")
				close(requestQueue)
				wg.Wait()
				close(errorChan)
				return
			default:
			}
		}

		rowIndex++
		if rowIndex <= 1 { // Skip header row
			continue
		}

		paramsJSON, err := h.genParams(row)
		if err != nil {
			h.appendLog(fmt.Sprintf("Row %d param generation failed: %v", rowIndex, err))
			errorCount++
			processedCount++
			// 错误时也检查停止信号
			select {
			case <-ctx.Done():
				h.appendLog("Execution cancelled during error handling")
				close(requestQueue)
				wg.Wait()
				close(errorChan)
				return
			default:
				if processedCount%batchSize == 0 || processedCount == totalRows {
					h.updateProgress(processedCount, totalRows, successCount, errorCount)
				}
			}
			continue
		}

		// 批量发送任务，减少channel操作开销
		select {
		case <-ctx.Done():
			h.appendLog("Execution cancelled before sending task")
			close(requestQueue)
			wg.Wait()
			close(errorChan)
			return
		case requestQueue <- RequestTask{
			ParamsJSON: paramsJSON,
			RowIndex:   rowIndex,
		}:
			successCount++
			processedCount++
		}
		
		// 批量更新进度，减少UI更新频率
		if processedCount%batchSize == 0 || processedCount == totalRows {
			h.updateProgress(processedCount, totalRows, successCount, errorCount)
		}
	}

	close(requestQueue)
	wg.Wait()
	close(errorChan)
	
	// 最终状态更新
	h.updateProgress(totalRows, totalRows, successCount, errorCount)
	
	fyne.Do(func() {
		h.progressBar.SetValue(1.0)
		h.statusLabel.SetText(fmt.Sprintf("执行完成 - 成功: %d, 错误: %d", successCount, errorCount))
	})
	h.appendLog(fmt.Sprintf("Execution completed - Success: %d, Error: %d", successCount, errorCount))
}

// 新的参数生成函数，支持配置化映射
func (h *HTTPTool) genParams(rows []string) ([]byte, error) {
	if len(rows) == 1 {
		rows = strings.Split(rows[0], "\t")
	}
	
	// 获取参数映射配置
	mappings := h.getParamMappings()
	if len(mappings) == 0 {
		// 如果没有配置映射，使用原来的逻辑作为兼容
		return h.genParamsLegacy(rows)
	}
	
	// 根据参数模式生成不同格式的参数
	if h.config.ParamMode == "array" {
		return h.genParamsArray(rows, mappings)
	} else {
		return h.genParamsObject(rows, mappings)
	}
}

// 生成对象格式参数
func (h *HTTPTool) genParamsObject(rows []string, mappings []ParamMapping) ([]byte, error) {
	params := make(map[string]interface{})
	
	for _, mapping := range mappings {
		value, err := h.extractValueFromCSV(rows, mapping)
		if err != nil {
			h.appendLog(fmt.Sprintf("参数映射错误 [%s]: %v", mapping.ParamName, err))
			continue
		}
		params[mapping.ParamName] = value
	}
	
	return json.Marshal(params)
}

// 生成数组格式参数
func (h *HTTPTool) genParamsArray(rows []string, mappings []ParamMapping) ([]byte, error) {
	// 按数组索引排序映射
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].ArrayIndex < mappings[j].ArrayIndex
	})
	
	// 创建紧凑的数组，按顺序填充参数
	var params []interface{}
	
	// 填充数组参数
	for _, mapping := range mappings {
		value, err := h.extractValueFromCSV(rows, mapping)
		if err != nil {
			h.appendLog(fmt.Sprintf("参数映射错误 [索引%d]: %v", mapping.ArrayIndex, err))
			continue
		}
		params = append(params, value)
	}
	
	return json.Marshal(params)
}

// 兼容原有逻辑的函数
func (h *HTTPTool) genParamsLegacy(rows []string) ([]byte, error) {
	if len(rows) < 5 {
		return nil, fmt.Errorf("CSV行数据不足")
	}
	
	req := []interface{}{
		rows[4], // settleOrderId
		0,       // orderId - 根据需要调整
	}
	
	if len(rows) > 2 {
		if orderId, err := strconv.Atoi(rows[2]); err == nil {
			req[1] = orderId
		}
	}

	return json.Marshal(req)
}

// 从CSV行中提取值并转换类型
func (h *HTTPTool) extractValueFromCSV(rows []string, mapping ParamMapping) (interface{}, error) {
	var rawValue string
	
	// 尝试按索引获取值
	if index, err := strconv.Atoi(mapping.CSVColumn); err == nil {
		if index >= 0 && index < len(rows) {
			rawValue = rows[index]
		} else {
			rawValue = mapping.DefaultValue
		}
	} else {
		// 按列名获取值（需要CSV头部支持）
		rawValue = mapping.DefaultValue // 暂时使用默认值，后续可扩展支持列名
	}
	
	// 如果值为空，使用默认值
	if strings.TrimSpace(rawValue) == "" {
		rawValue = mapping.DefaultValue
	}
	
	// 根据类型转换值
	return h.convertValueByType(rawValue, mapping.ParamType)
}

// 根据类型转换值
func (h *HTTPTool) convertValueByType(value, paramType string) (interface{}, error) {
	value = strings.TrimSpace(value)
	
	switch paramType {
	case "string":
		return value, nil
	case "int":
		if value == "" {
			return 0, nil
		}
		return strconv.Atoi(value)
	case "float":
		if value == "" {
			return 0.0, nil
		}
		return strconv.ParseFloat(value, 64)
	case "bool":
		if value == "" {
			return false, nil
		}
		return strconv.ParseBool(value)
	case "string[]":
		if value == "" {
			return []string{}, nil
		}
		// 支持多种分隔符：逗号、分号、竖线、空格
		separators := []string{",", ";", "|", " "}
		var parts []string
		for _, sep := range separators {
			if strings.Contains(value, sep) {
				parts = strings.Split(value, sep)
				break
			}
		}
		if len(parts) == 0 {
			parts = []string{value}
		}
		// 清理空白字符
		var result []string
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result, nil
	case "int[]":
		if value == "" {
			return []int{}, nil
		}
		// 支持多种分隔符：逗号、分号、竖线、空格
		separators := []string{",", ";", "|", " "}
		var parts []string
		for _, sep := range separators {
			if strings.Contains(value, sep) {
				parts = strings.Split(value, sep)
				break
			}
		}
		if len(parts) == 0 {
			parts = []string{value}
		}
		// 转换为整数
		var result []int
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				if num, err := strconv.Atoi(trimmed); err == nil {
					result = append(result, num)
				}
			}
		}
		return result, nil
	default:
		return value, nil
	}
}

func (h *HTTPTool) sendRequest(ctx context.Context, task RequestTask, ipList []string, maxRetries int) {
	retryCount := 0
	startTime := time.Now()
	
	// 预编译body模板，避免重复解析
	var bodyTemplate map[string]interface{}
	if err := json.Unmarshal([]byte(h.bodyEntry.Text), &bodyTemplate); err != nil {
		h.appendLog(fmt.Sprintf("Row %d body template parse failed: %v", task.RowIndex, err))
		return
	}
	
	for retryCount < maxRetries {
		select {
		case <-ctx.Done():
			return
		default:
		} 
		// 为每次重试添加指数退避延迟，避免惊群效应
		if retryCount > 0 {
			backoffDelay := time.Duration(retryCount*retryCount*100) * time.Millisecond
			if backoffDelay > 5*time.Second {
				backoffDelay = 5 * time.Second
			}
			time.Sleep(backoffDelay)
		}

		// 选择IP
		randomIP := ipList[task.RowIndex%len(ipList)]

		// 构造请求体 - 使用模板副本避免并发问题
		data := make(map[string]interface{})
		for k, v := range bodyTemplate {
			data[k] = v
		}
		data["ipPort"] = randomIP
		data["jsonParam"] = string(task.ParamsJSON)

		body, err := json.Marshal(data)
		if err != nil {
			h.appendLog(fmt.Sprintf("Row %d JSON marshal failed: %v", task.RowIndex, err))
			return
		}

		// 创建带超时的子上下文
		reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		
		// 创建请求 - 使用bytes.NewBuffer避免重复分配
		req, err := http.NewRequestWithContext(reqCtx, "POST", h.urlEntry.Text, bytes.NewBuffer(body))
		if err != nil {
			cancel()
			h.appendLog(fmt.Sprintf("Row %d request creation failed: %v", task.RowIndex, err))
			return
		}

		// 设置请求头
		h.setHeaders(req)

		// 发送请求 - 使用优化的HTTP客户端
		requestStart := time.Now()
		resp, err := httpClient.Do(req)
		requestDuration := time.Since(requestStart)
		cancel() // 立即取消上下文，释放资源
		
		if err != nil {
			retryCount++
			h.appendLog(fmt.Sprintf("Row %d request failed (retry %d/%d, duration: %v): %v",
				task.RowIndex, retryCount, maxRetries, requestDuration, err))
			continue
		}

		// 优化响应读取 - 限制响应大小避免内存问题
		const maxResponseSize = 5 * 1024 * 1024 // 5MB限制，减少内存使用
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		resp.Body.Close()
		
		if err != nil {
			retryCount++
			h.appendLog(fmt.Sprintf("Row %d response read failed (retry %d/%d): %v", task.RowIndex, retryCount, maxRetries, err))
			continue
		}

		// 检查响应状态码
		if resp.StatusCode >= 500 {
			retryCount++
			h.appendLog(fmt.Sprintf("Row %d server error %d (retry %d/%d)", task.RowIndex, resp.StatusCode, retryCount, maxRetries))
			continue
		}

		if resp.StatusCode >= 400 {
			// 4xx错误不重试，直接记录为失败
			h.appendLog(fmt.Sprintf("Row %d client error %d: %s", task.RowIndex, resp.StatusCode, string(respBody)))
			return
		}

		if strings.Contains(string(respBody), "call failed") {
			retryCount++
			h.appendLog(fmt.Sprintf("第%d行call failed(重试%d/%d): %s", task.RowIndex, retryCount, maxRetries, string(respBody)))
			continue
		}

		// 记录成功响应和耗时
		totalDuration := time.Since(startTime)
		h.appendLog(fmt.Sprintf("Row %d success in %v (request: %v): %s",
			task.RowIndex, totalDuration, requestDuration, string(respBody)))
		return
	}
	
	h.appendLog(fmt.Sprintf("Row %d final failure after %d retries, total time: %v", task.RowIndex, maxRetries, time.Since(startTime)))
}

func (h *HTTPTool) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "intest-manager.jd.com")
	req.Header.Set("Origin", "http://xingyun.jd.com")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "http://xingyun.jd.com/deeptest/quicktest/list?env=master&parentId=21254&Id=137790")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")
	
	if cookie := strings.TrimSpace(h.cookieEntry.Text); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
}

// 设置优化的日志缓冲系统
func (h *HTTPTool) setupLogBuffer() {
	// 使用更长的间隔减少UI更新频率 - 增加到1秒
	h.logTicker = time.NewTicker(1000 * time.Millisecond)
	
	// 启动异步日志处理器
	go func() {
		for {
			select {
			case logMsg := <-h.logChannel:
				h.logMutex.Lock()
				h.logBuffer = append(h.logBuffer, logMsg)
				
				// 限制缓冲区大小，避免内存无限增长
				if len(h.logBuffer) > h.maxLogLines {
					// 保留最新的日志
					copy(h.logBuffer, h.logBuffer[len(h.logBuffer)-h.maxLogLines+20:])
					h.logBuffer = h.logBuffer[:h.maxLogLines-20]
				}
				h.logMutex.Unlock()
				
			case <-h.logTicker.C:
				h.flushLogBuffer()
			}
		}
	}()
}

// 设置进度更新处理器
func (h *HTTPTool) setupProgressUpdater() {
	go func() {
		for update := range h.progressChannel {
			// 进一步限制进度更新频率，避免UI过载 - 增加到500ms
			h.progressMutex.Lock()
			now := time.Now()
			if now.Sub(h.lastProgressUpdate) < 500*time.Millisecond {
				h.progressMutex.Unlock()
				continue
			}
			h.lastProgressUpdate = now
			h.progressMutex.Unlock()
			
			// 异步更新进度UI
			go func(u progressUpdate) {
				fyne.Do(func() {
					progress := float64(u.processed) / float64(u.total)
					h.progressBar.SetValue(progress)
					// 简化状态文本，减少UI计算
					h.statusLabel.SetText(fmt.Sprintf("%d/%d (%.0f%%) 成功:%d 错误:%d",
						u.processed, u.total, progress*100, u.success, u.errors))
				})
			}(update)
		}
	}()
}

// 优化的进度更新函数
func (h *HTTPTool) updateProgress(processed, total, success, errors int) {
	select {
	case h.progressChannel <- progressUpdate{
		processed: processed,
		total:     total,
		success:   success,
		errors:    errors,
	}:
		// 成功发送进度更新
	default:
		// 通道满了，跳过这次更新，避免阻塞
	}
}

// 优化的日志刷新函数 - 严格减少UI阻塞
func (h *HTTPTool) flushLogBuffer() {
	// 防止过于频繁的UI更新 - 增加到500ms
	h.uiUpdateMutex.Lock()
	now := time.Now()
	if now.Sub(h.lastUIUpdate) < 500*time.Millisecond {
		h.uiUpdateMutex.Unlock()
		return
	}
	h.lastUIUpdate = now
	h.uiUpdateMutex.Unlock()
	
	h.logMutex.Lock()
	if len(h.logBuffer) == 0 {
		h.logMutex.Unlock()
		return
	}
	
	// 只取最新的日志进行更新，减少处理量
	var logsToUpdate []string
	bufferLen := len(h.logBuffer)
	if bufferLen > 20 { // 减少到20条，避免UI处理过多数据
		// 只取最新20条日志更新UI
		logsToUpdate = make([]string, 20)
		copy(logsToUpdate, h.logBuffer[bufferLen-20:])
	} else {
		logsToUpdate = make([]string, bufferLen)
		copy(logsToUpdate, h.logBuffer)
	}
	
	// 清空缓冲区
	h.logBuffer = h.logBuffer[:0]
	h.logMutex.Unlock()
	
	// 异步更新UI，避免阻塞
	go func() {
		fyne.Do(func() {
			// 使用字符串构建器池减少内存分配
			builder := getStringBuilder()
			defer putStringBuilder(builder)
			
			currentText := h.outputText.Text
			
			// 进一步限制显示的总行数，避免文本过长导致性能问题
			lines := strings.Split(currentText, "\n")
			if len(lines) > 100 { // 减少到100行
				// 只保留最新80行
				lines = lines[len(lines)-80:]
				for i, line := range lines {
					if i > 0 {
						builder.WriteString("\n")
					}
					builder.WriteString(line)
				}
				builder.WriteString("\n")
			} else {
				builder.WriteString(currentText)
			}
			
			// 批量添加新日志，减少SetText调用次数
			for _, log := range logsToUpdate {
				builder.WriteString(log)
				builder.WriteString("\n")
			}
			
			// 只更新文本，不自动滚动，减少UI操作
			h.outputText.SetText(builder.String())
		})
	}()
}

// 高性能异步日志添加函数
func (h *HTTPTool) appendLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)
	
	// 使用非阻塞发送，避免goroutine阻塞
	select {
	case h.logChannel <- logMessage:
		// 成功发送到通道
	default:
		// 通道满了，丢弃旧日志，避免阻塞
		// 这种情况下优先保证程序响应性
	}
}

// 停止日志缓冲系统
func (h *HTTPTool) stopLogBuffer() {
	if h.logTicker != nil {
		h.logTicker.Stop()
		h.flushLogBuffer() // 最后刷新一次
	}
	
	// 关闭通道，避免goroutine泄漏
	if h.logChannel != nil {
		close(h.logChannel)
	}
	if h.progressChannel != nil {
		close(h.progressChannel)
	}
}

func (h *HTTPTool) saveConfig() {
	config := &Config{
		URL:           h.urlEntry.Text,
		Cookie:        h.cookieEntry.Text,
		BodyTemp:      h.bodyEntry.Text,
		IPList:        strings.Split(h.ipListEntry.Text, "\n"),
		QPS:           h.parseIntOrDefault(h.qpsEntry.Text, 25),
		Workers:       h.parseIntOrDefault(h.workersEntry.Text, 100),
		MaxRetries:    h.parseIntOrDefault(h.retriesEntry.Text, 3),
		ParamMappings: h.getParamMappings(),
		ParamMode:     h.paramModeSelect.Selected,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		dialog.ShowError(fmt.Errorf("Config serialization failed: %v", err), h.window)
		return
	}

	fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()
		
		if _, err := writer.Write(data); err != nil {
			dialog.ShowError(fmt.Errorf("Config save failed: %v", err), h.window)
			return
		}
		
		dialog.ShowInformation("Success", "Config saved", h.window)
	}, h.window)
	
	// 设置对话框大小
	fileSaveDialog.Resize(fyne.NewSize(800, 600))
	fileSaveDialog.Show()
}

func (h *HTTPTool) loadConfig() {
	// 尝试从应用数据目录加载默认配置
	h.loadDefaultConfig()
}

// 手动加载配置文件（用户点击加载按钮时调用）
func (h *HTTPTool) loadConfigFromFile() {
	// 显示文件选择对话框
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Config file read failed: %v", err), h.window)
			return
		}

		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			dialog.ShowError(fmt.Errorf("Config file parse failed: %v", err), h.window)
			return
		}

		h.applyConfig(&config)
		dialog.ShowInformation("Success", "Config loaded", h.window)
	}, h.window)
	
	// 设置对话框大小
	fileDialog.Resize(fyne.NewSize(800, 600))
	fileDialog.Show()
}

func (h *HTTPTool) loadDefaultConfig() bool {
	configPath := filepath.Join(h.getConfigDir(), "config.json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return false
	}

	h.applyConfig(&config)
	return true
}

func (h *HTTPTool) applyConfig(config *Config) {
	h.urlEntry.SetText(config.URL)
	h.cookieEntry.SetText(config.Cookie)
	h.bodyEntry.SetText(config.BodyTemp)
	h.ipListEntry.SetText(strings.Join(config.IPList, "\n"))
	h.qpsEntry.SetText(strconv.Itoa(config.QPS))
	h.workersEntry.SetText(strconv.Itoa(config.Workers))
	h.retriesEntry.SetText(strconv.Itoa(config.MaxRetries))
	
	// 应用参数模式配置
	if config.ParamMode != "" {
		h.config.ParamMode = config.ParamMode
		h.paramModeSelect.SetSelected(config.ParamMode)
	}
	
	// 应用参数映射配置
	if len(config.ParamMappings) > 0 {
		h.setParamMappings(config.ParamMappings)
	}
}

func (h *HTTPTool) getConfigDir() string {
	configDir := h.app.Storage().RootURI().Path()
	return configDir
}

func (h *HTTPTool) parseIntOrDefault(s string, defaultValue int) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultValue
}
// 创建参数映射行
func (h *HTTPTool) createParamMappingRow() *ParamMappingRow {
	csvColumnEntry := widget.NewEntry()
	csvColumnEntry.SetPlaceHolder("列名或索引(如: 0, 1, name)。列表类型用逗号/分号分隔")
	
	paramNameEntry := widget.NewEntry()
	if h.config.ParamMode == "object" {
		paramNameEntry.SetPlaceHolder("参数名")
	} else {
		paramNameEntry.SetPlaceHolder("参数描述(可选)")
	}
	
	paramTypeSelect := widget.NewSelect(
		[]string{"string", "int", "float", "bool", "string[]", "int[]"},
		nil,
	)
	paramTypeSelect.SetSelected("string")
	
	defaultValueEntry := widget.NewEntry()
	defaultValueEntry.SetPlaceHolder("默认值(可选)")
	
	arrayIndexEntry := widget.NewEntry()
	arrayIndexEntry.SetPlaceHolder("数组索引")
	
	row := &ParamMappingRow{
		CSVColumnEntry:    csvColumnEntry,
		ParamNameEntry:    paramNameEntry,
		ParamTypeSelect:   paramTypeSelect,
		DefaultValueEntry: defaultValueEntry,
		ArrayIndexEntry:   arrayIndexEntry,
	}
	
	deleteButton := widget.NewButton("🗑", func() {
		h.removeParamMappingRow(row)
	})
	deleteButton.Importance = widget.LowImportance
	
	row.DeleteButton = deleteButton
	
	// 根据模式创建不同的行容器
	if h.config.ParamMode == "array" {
		row.Container = container.NewGridWithColumns(6,
			csvColumnEntry,
			arrayIndexEntry,
			paramNameEntry,
			paramTypeSelect,
			defaultValueEntry,
			deleteButton,
		)
	} else {
		row.Container = container.NewGridWithColumns(5,
			csvColumnEntry,
			paramNameEntry,
			paramTypeSelect,
			defaultValueEntry,
			deleteButton,
		)
	}
	
	return row
}

// 添加参数映射行
func (h *HTTPTool) addParamMappingRow() {
	row := h.createParamMappingRow()
	h.paramMappingList = append(h.paramMappingList, row)
	h.refreshParamMappingContainer()
}

// 移除参数映射行
func (h *HTTPTool) removeParamMappingRow(targetRow *ParamMappingRow) {
	for i, row := range h.paramMappingList {
		if row == targetRow {
			h.paramMappingList = append(h.paramMappingList[:i], h.paramMappingList[i+1:]...)
			break
		}
	}
	h.refreshParamMappingContainer()
}

// 刷新参数映射容器
func (h *HTTPTool) refreshParamMappingContainer() {
	h.paramMappingContainer.RemoveAll()
	
	// 根据模式添加不同的标题行
	var headerContainer *fyne.Container
	if h.config.ParamMode == "array" {
		headerContainer = container.NewGridWithColumns(6,
			widget.NewLabel("CSV列"),
			widget.NewLabel("数组索引"),
			widget.NewLabel("参数描述"),
			widget.NewLabel("类型"),
			widget.NewLabel("默认值"),
			widget.NewLabel("操作"),
		)
	} else {
		headerContainer = container.NewGridWithColumns(5,
			widget.NewLabel("CSV列"),
			widget.NewLabel("参数名"),
			widget.NewLabel("类型"),
			widget.NewLabel("默认值"),
			widget.NewLabel("操作"),
		)
	}
	h.paramMappingContainer.Add(headerContainer)
	
	// 添加分隔线
	h.paramMappingContainer.Add(widget.NewSeparator())
	
	// 重新创建所有映射行以适应新模式
	for _, row := range h.paramMappingList {
		// 保存现有数据
		csvColumn := row.CSVColumnEntry.Text
		paramName := row.ParamNameEntry.Text
		paramType := row.ParamTypeSelect.Selected
		defaultValue := row.DefaultValueEntry.Text
		arrayIndex := ""
		if row.ArrayIndexEntry != nil {
			arrayIndex = row.ArrayIndexEntry.Text
		}
		
		// 重新创建行
		newRow := h.createParamMappingRow()
		newRow.CSVColumnEntry.SetText(csvColumn)
		newRow.ParamNameEntry.SetText(paramName)
		newRow.ParamTypeSelect.SetSelected(paramType)
		newRow.DefaultValueEntry.SetText(defaultValue)
		if h.config.ParamMode == "array" && newRow.ArrayIndexEntry != nil {
			newRow.ArrayIndexEntry.SetText(arrayIndex)
		}
		
		// 更新引用
		for i, oldRow := range h.paramMappingList {
			if oldRow == row {
				h.paramMappingList[i] = newRow
				break
			}
		}
	}
	
	// 添加所有映射行
	for _, row := range h.paramMappingList {
		h.paramMappingContainer.Add(row.Container)
	}
	
	// 如果没有映射行，显示提示信息
	if len(h.paramMappingList) == 0 {
		if h.config.ParamMode == "array" {
			h.paramMappingContainer.Add(widget.NewLabel("暂无参数映射配置，数组模式将按索引顺序生成参数列表"))
		} else {
			h.paramMappingContainer.Add(widget.NewLabel("暂无参数映射配置，对象模式将生成键值对参数"))
		}
	}
	
	// 添加分隔线
	h.paramMappingContainer.Add(widget.NewSeparator())
	
	// 添加"添加行"按钮
	addButton := widget.NewButton("➕ 添加参数映射", func() {
		h.addParamMappingRow()
	})
	addButton.Importance = widget.MediumImportance
	h.paramMappingContainer.Add(addButton)
	
	h.paramMappingContainer.Refresh()
}

// 获取参数映射配置
func (h *HTTPTool) getParamMappings() []ParamMapping {
	var mappings []ParamMapping
	for _, row := range h.paramMappingList {
		if strings.TrimSpace(row.CSVColumnEntry.Text) != "" {
			mapping := ParamMapping{
				CSVColumn:    strings.TrimSpace(row.CSVColumnEntry.Text),
				ParamName:    strings.TrimSpace(row.ParamNameEntry.Text),
				ParamType:    row.ParamTypeSelect.Selected,
				DefaultValue: strings.TrimSpace(row.DefaultValueEntry.Text),
			}
			
			// 如果是数组模式，获取数组索引
			if h.config.ParamMode == "array" && row.ArrayIndexEntry != nil {
				if arrayIndex, err := strconv.Atoi(strings.TrimSpace(row.ArrayIndexEntry.Text)); err == nil {
					mapping.ArrayIndex = arrayIndex
				}
			}
			
			mappings = append(mappings, mapping)
		}
	}
	return mappings
}

// 设置参数映射配置
func (h *HTTPTool) setParamMappings(mappings []ParamMapping) {
	h.paramMappingList = nil
	for _, mapping := range mappings {
		row := h.createParamMappingRow()
		row.CSVColumnEntry.SetText(mapping.CSVColumn)
		row.ParamNameEntry.SetText(mapping.ParamName)
		row.ParamTypeSelect.SetSelected(mapping.ParamType)
		row.DefaultValueEntry.SetText(mapping.DefaultValue)
		if h.config.ParamMode == "array" && row.ArrayIndexEntry != nil {
			row.ArrayIndexEntry.SetText(strconv.Itoa(mapping.ArrayIndex))
		}
		h.paramMappingList = append(h.paramMappingList, row)
	}
	h.refreshParamMappingContainer()
}
