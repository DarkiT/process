package process

type State int

const (
	Stopped  State = iota // Stopped 已停止
	Starting       = 10   // Starting 启动中
	Running        = 20   // Running 运行中
	Backoff        = 30   // Backoff 已挂起
	Stopping       = 40   // Stopping 停止中
	Exited         = 100  // Exited 已退出
	Fatal          = 200  // Fatal 启动失败
	Unknown        = 1000 // Unknown 未知状态
)

// String 把进程状态转换成可识别的字符串
func (p State) String() string {
	switch p {
	case Stopped:
		return "Stopped"
	case Starting:
		return "Starting"
	case Running:
		return "Running"
	case Backoff:
		return "Backoff"
	case Stopping:
		return "Stopping"
	case Exited:
		return "Exited"
	case Fatal:
		return "Fatal"
	case Unknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}
