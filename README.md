# Process 进程管理库

[![Go Reference](https://pkg.go.dev/badge/github.com/darkit/process.svg)](https://pkg.go.dev/github.com/darkit/process)
[![Go Report Card](https://goreportcard.com/badge/github.com/darkit/process)](https://goreportcard.com/report/github.com/darkit/process)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/darkit/process/blob/master/LICENSE)


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
    process.WithName("myapp"),
    process.WithCommand("./myapp"),
    process.WithArgs("--config", "config.yaml"),
    process.WithDirectory("/app"),
    process.WithAutoStart(true),
    process.WithAutoReStart(process.AutoReStartTrue),
    process.WithStdoutLog("logs/stdout.log", "50MB", 10),
    process.WithStderrLog("logs/stderr.log", "50MB", 10),
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

- `WithName(name string)` - 设置进程名称
- `WithCommand(cmd string)` - 设置启动命令
- `WithArgs(args ...string)` - 设置启动参数
- `WithDirectory(dir string)` - 设置工作目录
- `WithAutoStart(auto bool)` - 设置是否自动启动
- `WithAutoReStart(restart AutoReStart)` - 设置自动重启策略
- `WithUser(user string)` - 设置运行用户
- `WithEnvironment(env map[string]string)` - 设置环境变量
- `WithStdoutLog(file string, maxBytes string, backups int)` - 设置标准输出日志
- `WithStderrLog(file string, maxBytes string, backups int)` - 设置错误输出日志
- `WithStartRetries(retries int)` - 设置启动重试次数
- `WithStartSecs(secs int)` - 设置启动超时时间
- `WithStopWaitSecs(secs int)` - 设置停止等待时间
- `WithPriority(priority int)` - 设置启动优先级

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

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。