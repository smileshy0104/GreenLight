package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// 定义levels
type Level int8

// 定义各种 severity levels
const (
	LevelInfo  Level = iota // Has the value of 0.
	LevelError              // Has the value of 1.
	LevelFatal              // Has the value of 2.
	LevelOff                // Has the value of 3.
)

// 每种 level 有对应的字符串
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Logger 日志实例结构体
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// 创建 一个 Logger 实例
func NewLogger(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// PrintInfo 是一个 helper 方法，它写一个 Info 级别的日志条目
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

// PrintInfo 是一个 helper 方法，它写一个 Error 级别的日志条目
func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

// PrintFatal 是一个 helper 方法，它写一个 Fatal 级别的日志条目
func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	// 直接 退出程序
	os.Exit(1)
}

// print 方法 打印各种错误等级的日志
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	// 如果日志等级小于最小等级，则直接返回
	if level < l.minLevel {
		return 0, nil
	}

	// 声明 一个匿名结构体存放日志信息
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	// 包含 错误等级的堆栈跟踪
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// 声明一个变量，用于存储日志的 JSON 格式
	var line []byte

	// 将日志信息转换为 JSON 格式
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message:" + err.Error())
	}

	// 获取锁，并使用 defer 语句确保锁在函数返回时自动释放
	l.mu.Lock()
	defer l.mu.Unlock()

	// 将日志信息写入到 io.Writer 中，并返回写入的字节数和错误信息
	return l.out.Write(append(line, '\n'))
}

// Write 方法 实现了 io.Writer 接口
func (l *Logger) Write(message []byte) (n int, err error) {
	// 调用 print 方法，将日志信息写入到 io.Writer 中
	return l.print(LevelError, string(message), nil)
}
