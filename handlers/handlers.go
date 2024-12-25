package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/darkit/process"
)

// Handler 是一个泛型接口，定义了所有处理方法
type Handler[T any] interface {
	ListProcesses() T
	CreateProcess() T
	DeleteProcess() T
	StartProcess() T
	StopProcess() T
	RestartProcess() T
	GetStdoutLog() T
	GetStderrLog() T
}

// ProcessHandler 是一个泛型结构体，实现了 Handler 接口
type ProcessHandler[T any] struct {
	manager *process.Manager
	warp    func(http.HandlerFunc) T
}

// NewProcessHandler 创建一个新的 ProcessHandler 实例
func NewProcessHandler[T any](manager *process.Manager, warp func(http.HandlerFunc) T) *ProcessHandler[T] {
	return &ProcessHandler[T]{
		manager: manager,
		warp:    warp,
	}
}

// GetHandlers 返回实现了 Handler 接口的 ProcessHandler
func (h *ProcessHandler[T]) GetHandlers() Handler[T] {
	return h
}

// ListProcesses 获取进程列表
func (h *ProcessHandler[T]) ListProcesses() T {
	return h.warp(func(w http.ResponseWriter, r *http.Request) {
		processes, err := h.manager.GetAllProcessInfo()
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"data": processes,
		})
	})
}

// CreateProcess 创建新进程
func (h *ProcessHandler[T]) CreateProcess() T {
	return h.warp(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name          string `json:"name"`
			Command       string `json:"command"`
			Args          string `json:"args"`
			Directory     string `json:"directory"`
			User          string `json:"user"`
			Environment   string `json:"environment"`
			AutoStart     bool   `json:"autoStart"`
			AutoRestart   int    `json:"autoRestart"`
			StdoutLogfile string `json:"stdoutLogfile"`
			StderrLogfile string `json:"stderrLogfile"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "参数错误")
			return
		}

		env := make(map[string]string)
		if req.Environment != "" {
			for _, line := range strings.Split(req.Environment, "\n") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					env[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}

		args := []string{}
		if req.Args != "" {
			args = strings.Fields(req.Args)
		}

		proc := process.NewProcess(
			process.WithName(req.Name),
			process.WithCommand(req.Command),
			process.WithArgs(args...),
			process.WithDirectory(req.Directory),
			process.WithUser(req.User),
			process.WithEnvironment(env),
			process.WithAutoStart(req.AutoStart),
			process.WithAutoReStart(process.AutoReStart(req.AutoRestart)),
			process.WithStdoutLog(req.StdoutLogfile, "50MB", 10),
			process.WithStderrLog(req.StderrLogfile, "50MB", 10),
		)

		proc, err := h.manager.NewProcessByProcess(proc)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if req.AutoStart {
			proc.Start(true)
		}

		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "创建成功",
		})
	})
}

// DeleteProcess 删除进程
func (h *ProcessHandler[T]) DeleteProcess() T {
	return h.warp(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		proc := h.manager.Find(name)
		if proc == nil {
			errorResponse(w, http.StatusNotFound, "进程不存在")
			return
		}

		proc.Stop(true)
		h.manager.Remove(name)

		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "删除成功",
		})
	})
}

// StartProcess 启动进程
func (h *ProcessHandler[T]) StartProcess() T {
	return h.warp(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		ok, err := h.manager.StartProcess(name, true)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "启动成功",
			"data": ok,
		})
	})
}

// StopProcess 停止进程
func (h *ProcessHandler[T]) StopProcess() T {
	return h.warp(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		ok, err := h.manager.StopProcess(name, true)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "停止成功",
			"data": ok,
		})
	})
}

// RestartProcess 重启进程
func (h *ProcessHandler[T]) RestartProcess() T {
	return h.warp(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		ok, err := h.manager.GracefulReload(name, true)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "重启成功",
			"data": ok,
		})
	})
}

// GetStdoutLog 获取标准输出日志
func (h *ProcessHandler[T]) GetStdoutLog() T {
	return h.warp(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		proc := h.manager.Find(name)
		if proc == nil {
			errorResponse(w, http.StatusNotFound, "进程不存在")
			return
		}

		logs, err := h.readLastLines(proc.GetStdoutLogfile(), 100)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"data": logs,
		})
	})
}

// GetStderrLog 获取错误输出日志
func (h *ProcessHandler[T]) GetStderrLog() T {
	return h.warp(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		proc := h.manager.Find(name)
		if proc == nil {
			errorResponse(w, http.StatusNotFound, "进程不存在")
			return
		}

		logs, err := h.readLastLines(proc.GetStderrLogfile(), 100)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"data": logs,
		})
	})
}

// 读取文件最后几行
func (h *ProcessHandler[T]) readLastLines(filename string, n int) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, n)

	for scanner.Scan() {
		if len(lines) >= n {
			lines = lines[1:]
		}
		lines = append(lines, scanner.Text())
	}

	return strings.Join(lines, "\n"), scanner.Err()
}

// 获取最后一行日志
func (h *ProcessHandler[T]) getLastLog(name string) string {
	proc := h.manager.Find(name)
	if proc == nil {
		return ""
	}

	logs, err := h.readLastLines(proc.GetStdoutLogfile(), 1)
	if err != nil {
		return ""
	}

	return logs
}

// SetupRoutes sets up the HTTP routes for the ProcessHandler
func (h *ProcessHandler[T]) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	setupRoute := func(path string, handler func() T) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			h := handler()
			if fn, ok := any(h).(func(http.ResponseWriter, *http.Request)); ok {
				fn(w, r)
			} else {
				errorResponse(w, http.StatusInternalServerError, "Handler type mismatch")
			}
		})
	}

	setupRoute("/processes", h.ListProcesses)
	setupRoute("/process/create", h.CreateProcess)
	setupRoute("/process/delete", h.DeleteProcess)
	setupRoute("/process/start", h.StartProcess)
	setupRoute("/process/stop", h.StopProcess)
	setupRoute("/process/restart", h.RestartProcess)
	setupRoute("/process/stdout", h.GetStdoutLog)
	setupRoute("/process/stderr", h.GetStderrLog)

	return mux
}

// jsonResponse is a helper function to send JSON responses
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// errorResponse is a helper function to send error responses in JSON format
func errorResponse(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]interface{}{
		"code": -1,
		"msg":  message,
	})
}

/*
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/darkit/process"
	"github.com/darkit/process/handlers"
)

func main() {
	// 创建进程管理器
	manager := process.NewManager()

	// 创建处理器
	Http := NewProcessHandler(manager, func(h http.HandlerFunc) http.HandlerFunc {
		return h
	})

	// 获取处理器接口
	HttpHandlers := Http.GetHandlers()

	// 创建路由
	mux := http.NewServeMux()
    mux.HandleFunc("GET /processes", HttpHandlers.ListProcesses())
    mux.HandleFunc("POST /process/create", HttpHandlers.CreateProcess())
    mux.HandleFunc("DELETE /process/delete", HttpHandlers.DeleteProcess())
    mux.HandleFunc("POST /process/start", HttpHandlers.StartProcess())
    mux.HandleFunc("POST /process/stop", HttpHandlers.StopProcess())
    mux.HandleFunc("POST /process/restart", HttpHandlers.RestartProcess())
    mux.HandleFunc("GET /process/stdout", HttpHandlers.GetStdoutLog())
    mux.HandleFunc("GET /process/stderr", HttpHandlers.GetStderrLog())

	// 启动服务器
	fmt.Println("Server is running on http://localhost:8080")
	_=http.ListenAndServe(":8080", mux)


	// 创建处理器
	Gin := NewProcessHandler(manager, func(h http.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			h(c.Writer, c.Request)
		}
	})

	// 获取处理器接口
	GinHandlers := Gin.GetHandlers()

	// 创建 Gin 路由
	r := gin.Default()

	// 设置路由
	r.GET("/processes", GinHandlers.ListProcesses())
	r.POST("/process/create", GinHandlers.CreateProcess())
	r.DELETE("/process/delete", GinHandlers.DeleteProcess())
	r.POST("/process/start", GinHandlers.StartProcess())
	r.POST("/process/stop", GinHandlers.StopProcess())
	r.POST("/process/restart", GinHandlers.RestartProcess())
	r.GET("/process/stdout", GinHandlers.GetStdoutLog())
	r.GET("/process/stderr", GinHandlers.GetStderrLog())

	// 启动服务器
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(r.Run(":8080"))
}
*/
