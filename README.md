# Process 进程管理库

Process 是一个用 Go 语言编写的轻量级进程管理库，提供了完整的进程生命周期管理功能，支持自动重启、日志管理、信号控制等特性。同时提供了 Web API 扩展，可以轻松集成到 HTTP 或 Gin 等 Web 框架中。

## 特性

- 完整的进程生命周期管理
- 支持自动重启策略
- 标准输出和错误输出日志管理
- 灵活的进程启动配置
- 多平台支持 (Linux、Darwin、Windows)
- Web API 扩展支持
- 优雅的进程重启机制
- 进程状态监控
- 支持用户权限控制
- 支持环境变量配置

## 安装

```bash
go get github.com/darkit/process
```

## 基础使用

### 创建进程管理器

```go
import "github.com/darkit/process"

// 创建进程管理器
manager := process.NewManager()
```

### 创建和管理进程

```go
// 方式1：使用选项创建进程
proc, err := manager.NewProcess(
    process.ProcName("myapp"),
    process.ProcCommand("./myapp"),
    process.ProcArgs("--config", "config.yaml"),
    process.ProcDirectory("/app"),
    process.ProcAutoStart(true),
    process.ProcAutoReStart(process.AutoReStartTrue),
    process.ProcStdoutLog("logs/stdout.log", "50MB", 10),
    process.ProcStderrLog("logs/stderr.log", "50MB", 10),
)

// 方式2：直接使用命令创建进程
proc, err := manager.NewProcessCmd("python app.py", nil)

// 启动进程
proc.Start(true)  // true 表示阻塞等待进程启动

// 停止进程
proc.Stop(true)   // true 表示阻塞等待进程停止

// 获取进程信息
info := proc.GetProcessInfo()
```

### 进程配置选项

- `ProcName(name string)` - 设置进程名称
- `ProcCommand(cmd string)` - 设置启动命令
- `ProcArgs(args ...string)` - 设置启动参数
- `ProcDirectory(dir string)` - 设置工作目录
- `ProcAutoStart(auto bool)` - 设置是否自动启动
- `ProcAutoReStart(restart AutoReStart)` - 设置自动重启策略
- `ProcUser(user string)` - 设置运行用户
- `ProcEnvironment(env map[string]string)` - 设置环境变量
- `ProcStdoutLog(file string, maxBytes string, backups int)` - 设置标准输出日志
- `ProcStderrLog(file string, maxBytes string, backups int)` - 设置错误输出日志
- `ProcStartRetries(retries int)` - 设置启动重试次数
- `ProcStartSecs(secs int)` - 设置启动超时时间
- `ProcStopWaitSecs(secs int)` - 设置停止等待时间
- `ProcPriority(priority int)` - 设置启动优先级

## Web API 扩展使用

Process 库提供了 Web API 扩展功能，支持通过 HTTP 接口管理进程。支持原生 HTTP 和 Gin 框架。

### HTTP Server 示例

```go
import (
    "github.com/darkit/process"
    "github.com/darkit/process/handlers"
)

func main() {
    // 创建进程管理器
    manager := process.NewManager()

    // 创建 HTTP 处理器
    handler := handlers.NewProcessHandler(manager, func(h http.HandlerFunc) http.HandlerFunc {
        return h
    })

    // 获取处理器接口
    httpHandlers := handler.GetHandlers()

    // 设置路由
    mux := http.NewServeMux()
    mux.HandleFunc("/processes", httpHandlers.ListProcesses())
    mux.HandleFunc("/process/create", httpHandlers.CreateProcess())
    mux.HandleFunc("/process/delete", httpHandlers.DeleteProcess())
    mux.HandleFunc("/process/start", httpHandlers.StartProcess())
    mux.HandleFunc("/process/stop", httpHandlers.StopProcess())
    mux.HandleFunc("/process/restart", httpHandlers.RestartProcess())
    mux.HandleFunc("/process/stdout", httpHandlers.GetStdoutLog())
    mux.HandleFunc("/process/stderr", httpHandlers.GetStderrLog())

    // 启动服务器
    http.ListenAndServe(":8080", mux)
}
```

### Gin 框架集成示例

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/darkit/process"
    "github.com/darkit/process/handlers"
)

func main() {
    // 创建进程管理器
    manager := process.NewManager()

    // 创建 Gin 处理器
    handler := handlers.NewProcessHandler(manager, func(h http.HandlerFunc) gin.HandlerFunc {
        return func(c *gin.Context) {
            h(c.Writer, c.Request)
        }
    })

    // 获取处理器接口
    ginHandlers := handler.GetHandlers()

    // 创建 Gin 路由
    r := gin.Default()
    
    // 设置路由
    r.GET("/processes", ginHandlers.ListProcesses())
    r.POST("/process/create", ginHandlers.CreateProcess())
    r.DELETE("/process/delete", ginHandlers.DeleteProcess())
    r.POST("/process/start", ginHandlers.StartProcess())
    r.POST("/process/stop", ginHandlers.StopProcess())
    r.POST("/process/restart", ginHandlers.RestartProcess())
    r.GET("/process/stdout", ginHandlers.GetStdoutLog())
    r.GET("/process/stderr", ginHandlers.GetStderrLog())

    // 启动服务器
    r.Run(":8080")
}
```

### Web API 接口说明

| 接口 | 方法 | 描述 |
|------|------|------|
| `/processes` | GET | 获取所有进程列表 |
| `/process/create` | POST | 创建新进程 |
| `/process/delete` | DELETE | 删除指定进程 |
| `/process/start` | POST | 启动指定进程 |
| `/process/stop` | POST | 停止指定进程 |
| `/process/restart` | POST | 重启指定进程 |
| `/process/stdout` | GET | 获取标准输出日志 |
| `/process/stderr` | GET | 获取错误输出日志 |

#### 创建进程 POST 请求示例

```json
{
    "name": "myapp",
    "command": "./myapp",
    "args": "--config config.yaml",
    "directory": "/app",
    "user": "appuser",
    "environment": "KEY1=value1\nKEY2=value2",
    "autoStart": true,
    "autoRestart": 1,
    "stdoutLogfile": "logs/stdout.log",
    "stderrLogfile": "logs/stderr.log"
}
```

## 贡献

欢迎提交问题和 Pull Request。在提交 PR 之前，请确保：

1. 更新测试
2. 更新文档
3. 遵循代码规范
4. 提供必要的测试用例

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 作者

- DarkIT Team 
- Email: <your-email@example.com>
- Website: https://your-website.com

## 更新日志

### v1.0.0 (2024-04-20)

- 初始版本发布
- 支持基本的进程管理功能
- 添加 Web API 扩展支持