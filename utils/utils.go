package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// DefaultTrimChars are the characters which are stripped by Trim* functions in default.
var DefaultTrimChars = string([]byte{
	'\t', // Tab.
	'\v', // Vertical tab.
	'\n', // New line (line feed).
	'\r', // Carriage return.
	'\f', // New page.
	' ',  // Ordinary space.
	0x00, // NUL-byte.
	0x85, // Delete.
	0xA0, // Non-breaking space.
})

// GetBytes 通过可识别字符串，返回数字容量
//
//	logSize=1MB
//	logSize=1GB
//	logSize=1KB
//	logSize=1024
func GetBytes(value string, defValue int) int {
	if len(value) > 2 {
		lastTwoBytes := value[len(value)-2:]
		if lastTwoBytes == "MB" {
			return toInt(value[:len(value)-2], 1024*1024, defValue)
		} else if lastTwoBytes == "GB" {
			return toInt(value[:len(value)-2], 1024*1024*1024, defValue)
		} else if lastTwoBytes == "KB" {
			return toInt(value[:len(value)-2], 1024, defValue)
		}
		return toInt(value, 1, defValue)
	}
	return defValue
}

func toInt(s string, factor int, defValue int) int {
	i, err := strconv.Atoi(s)
	if err == nil {
		return i * factor
	}
	return defValue
}

// RealPath converts the given `path` to its absolute path
// and checks if the file path exists.
// If the file does not exist, return an empty string.
func RealPath(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	if !Exists(p) {
		return ""
	}
	return p
}

// Exists checks whether given `path` exist.
func Exists(path string) bool {
	if stat, err := os.Stat(path); stat != nil && !os.IsNotExist(err) {
		return true
	}
	return false
}

// Trim strips whitespace (or other characters) from the beginning and end of a string.
// The optional parameter `characterMask` specifies the additional stripped characters.
func Trim(str string, characterMask ...string) string {
	trimChars := DefaultTrimChars
	if len(characterMask) > 0 {
		trimChars += characterMask[0]
	}
	return strings.Trim(str, trimChars)
}

// SplitAndTrim splits string `str` by a string `delimiter` to an array,
// and calls Trim to every element of this array. It ignores the elements
// which are empty after Trim.
func SplitAndTrim(str, delimiter string, characterMask ...string) []string {
	array := make([]string, 0)
	for _, v := range strings.Split(str, delimiter) {
		v = Trim(v, characterMask...)
		if v != "" {
			array = append(array, v)
		}
	}
	return array
}

// SearchBinary searches the binary `file` in current working folder and PATH environment.
func SearchBinary(file string) string {
	// Check if it is absolute path of exists at current working directory.
	if Exists(file) {
		return file
	}
	return SearchBinaryPath(file)
}

// SearchBinaryPath searches the binary `file` in PATH environment.
func SearchBinaryPath(file string) string {
	array := ([]string)(nil)
	switch runtime.GOOS {
	case "windows":
		envPath := os.Getenv("Path")
		if strings.Contains(envPath, ";") {
			array = SplitAndTrim(envPath, ";")
		} else if strings.Contains(envPath, ":") {
			array = SplitAndTrim(envPath, ":")
		}
		if Ext(file) != ".exe" {
			file += ".exe"
		}

	default:
		array = SplitAndTrim(os.Getenv("Path"), ":")
	}
	if len(array) > 0 {
		path := ""
		for _, v := range array {
			path = v + string(filepath.Separator) + file
			if Exists(path) && IsFile(path) {
				return path
			}
		}
	}
	return ""
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func Stat(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		err = errors.New(fmt.Sprintf(`os.Stat failed for file "%s"`, path))
	}
	return info, err
}

// IsFile checks whether given `path` a file, which means it's not a directory.
// Note that it returns false if the `path` does not exist.
func IsFile(path string) bool {
	s, err := Stat(path)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

// Ext returns the file name extension used by path.
// The extension is the suffix beginning at the final dot
// in the final element of path; it is empty if there is
// no dot.
// Note: the result contains symbol '.'.
//
// Example:
// Ext("main.go")  => .go
// Ext("api.json") => .json
func Ext(path string) string {
	ext := filepath.Ext(path)
	if p := strings.IndexByte(ext, '?'); p != -1 {
		ext = ext[0:p]
	}
	return ext
}
