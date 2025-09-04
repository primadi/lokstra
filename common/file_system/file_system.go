package file_system

import (
	"fmt"
	"io/fs"
	"strings"
)

type FSWithSubFS struct {
	fs.FS
	SubFS string
}

func findCommonFolder(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	if len(paths) == 1 {
		idx := strings.LastIndex(paths[0], "/")
		if idx == -1 {
			return ""
		}
		return paths[0][:idx]
	}

	splitPaths := make([][]string, len(paths))
	for i, p := range paths {
		splitPaths[i] = strings.Split(p, "/")
	}
	var common []string
	for i := 0; ; i++ {
		var folder string
		for j, sp := range splitPaths {
			if i >= len(sp) {
				return strings.Join(common, "/")
			}
			if j == 0 {
				folder = sp[i]
			} else if sp[i] != folder {
				return strings.Join(common, "/")
			}
		}
		common = append(common, folder)
	}
}

// SubFirstCommonFolder takes an fs.FS and returns a sub fs.FS that starts from the first common folder
// If no common folder is found, it returns an error
func SubFirstCommonFolder(fsys fs.FS) (fs.FS, error) {
	var paths []string
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if len(paths) == 0 {
		return fsys, nil // No files, return original FS
	}

	commonFolder := findCommonFolder(paths)
	if commonFolder != "" && commonFolder != "." {
		return fs.Sub(fsys, commonFolder)
	}
	return nil, fmt.Errorf("no commonFolder found in FS")
}
