package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type StrStrMap struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewStrStrMap() *StrStrMap {
	return &StrStrMap{
		data: make(map[string]string),
	}
}

func (m *StrStrMap) Set(key string, val string) {
	m.mu.Lock()
	m.data[key] = val
	m.mu.Unlock()
}

func (m *StrStrMap) Sets(data map[string]string) {
	m.mu.Lock()
	for k, v := range data {
		m.data[k] = v
	}
	m.mu.Unlock()
}

func (m *StrStrMap) Get(key string) string {
	m.mu.RLock()
	val := m.data[key]
	m.mu.RUnlock()
	return val
}

func (m *StrStrMap) Size() int {
	m.mu.RLock()
	length := len(m.data)
	m.mu.RUnlock()
	return length
}

func (m *StrStrMap) Map() map[string]string {
	m.mu.RLock()
	data := make(map[string]string, len(m.data))
	for k, v := range m.data {
		data[k] = v
	}
	m.mu.RUnlock()
	return data
}

type AnyAnyMap struct {
	mu   sync.RWMutex
	data map[interface{}]interface{}
}

func NewAnyAnyMap() *AnyAnyMap {
	return &AnyAnyMap{
		data: make(map[interface{}]interface{}),
	}
}

func (m *AnyAnyMap) Set(key interface{}, val interface{}) {
	m.mu.Lock()
	m.data[key] = val
	m.mu.Unlock()
}

func (m *AnyAnyMap) Get(key interface{}) interface{} {
	m.mu.RLock()
	val := m.data[key]
	m.mu.RUnlock()
	return val
}

func SearchBinary(binary string) string {
	if filepath.IsAbs(binary) {
		if Exists(binary) {
			return binary
		}
		return ""
	}

	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	for _, path := range paths {
		file := filepath.Join(path, binary)
		if Exists(file) {
			return file
		}
	}
	return ""
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func RealPath(path string) string {
	if path == "" {
		return ""
	}

	if path[0] != '~' {
		path, _ = filepath.Abs(path)
		return path
	}

	home, _ := os.UserHomeDir()
	if len(path) > 1 {
		return filepath.Join(home, path[1:])
	}
	return home
}

func GetBytes(size string, defaultSize int) int {
	size = strings.ToUpper(strings.TrimSpace(size))
	if size == "" {
		return defaultSize
	}

	unit := size[len(size)-2:]
	value := size[:len(size)-2]

	var multiplier int
	switch unit {
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	default:
		return defaultSize
	}

	bytes := 0
	_, err := fmt.Sscanf(value, "%d", &bytes)
	if err != nil {
		return defaultSize
	}

	return bytes * multiplier
}

func SplitAndTrim(str, sep string) []string {
	parts := strings.Split(str, sep)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func SetMap(m map[string]string) error {
	for k, v := range m {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

func Map() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			env[e[:i]] = e[i+1:]
		}
	}
	return env
}

func All() []string {
	return os.Environ()
}
