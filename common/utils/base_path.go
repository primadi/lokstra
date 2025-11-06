package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

func GetBasePath() string {
	// 1. Find from executable location (production mode)
	exePath, err := os.Executable()
	if err == nil {
		path := filepath.Dir(exePath)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 2. If not found, fallback to source location (debug mode)
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		path := filepath.Dir(filename)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 3. Finally, fallback to current working directory
	path, _ := os.Getwd()
	return path
}
