package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func GetBasePath() string {
	// 1. Find from executable location (production mode)
	exePath, err := os.Executable()
	if err == nil {
		exePath, _ = filepath.EvalSymlinks(exePath)
		path := filepath.Dir(exePath)

		// detect if not running from /tmp/go-build (which is go run temp dir)
		if !strings.Contains(path, string(os.PathSeparator)+"go-build") {
			return path
		}
	}

	// 2. find main.go location
	for i := 2; i < 15; i++ {
		_, filename, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if strings.Contains(filename, "/vendor/") {
			return filename[:strings.Index(filename, "/vendor/")]
		}

		if strings.HasSuffix(filename, "main.go") {
			return filepath.Dir(filename)
		}
	}

	// 3. Finally, fallback to current working directory
	path, _ := os.Getwd()
	return path
}

func NormalizeWithBasePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(GetBasePath(), path)
}
