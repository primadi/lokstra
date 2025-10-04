package utils

import (
	"io/fs"
	"os"
)

type FsFallback struct {
	FSList []fs.FS
}

var (
	_ fs.FS        = (*FsFallback)(nil)
	_ fs.StatFS    = (*FsFallback)(nil)
	_ fs.ReadDirFS = (*FsFallback)(nil)
	_ fs.SubFS     = (*FsFallback)(nil)
)

// Create a new FsFallback with given fs.FS list.
// The order of fs.FS matters: the first one has higher priority.
// If all fs.FS return os.ErrNotExist, the final result is os.ErrNotExist.
// If any fs.FS return other error, that error is returned immediately.
func NewFsFallback(fses ...fs.FS) *FsFallback {
	return &FsFallback{FSList: fses}
}

// adds a new fs.FS to the end of the list (lowest priority)
func (f *FsFallback) AddFs(fsys fs.FS) {
	f.FSList = append(f.FSList, fsys)
}

// adds a new fs.FS to the start of the list (highest priority)
func (f *FsFallback) AddFirstFs(fsys fs.FS) {
	f.FSList = append([]fs.FS{fsys}, f.FSList...)
}

// ---- fs.FS ----
func (f *FsFallback) Open(name string) (fs.File, error) {
	for _, filesystem := range f.FSList {
		file, err := filesystem.Open(name)
		if err == nil {
			return file, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return nil, os.ErrNotExist
}

// ---- fs.StatFS ----
func (f *FsFallback) Stat(name string) (fs.FileInfo, error) {
	for _, filesystem := range f.FSList {
		// Direct support
		if statFS, ok := filesystem.(fs.StatFS); ok {
			info, err := statFS.Stat(name)
			if err == nil {
				return info, nil
			}
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}
		// Fallback: open + stat
		file, err := filesystem.Open(name)
		if err == nil {
			defer file.Close()
			return file.Stat()
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return nil, os.ErrNotExist
}

// ---- fs.ReadDirFS ----
func (f *FsFallback) ReadDir(name string) ([]fs.DirEntry, error) {
	for _, filesystem := range f.FSList {
		// Direct support
		if rd, ok := filesystem.(fs.ReadDirFS); ok {
			entries, err := rd.ReadDir(name)
			if err == nil {
				return entries, nil
			}
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}
		// Fallback: open as ReadDirFile
		file, err := filesystem.Open(name)
		if err == nil {
			defer file.Close()
			if rdf, ok := file.(fs.ReadDirFile); ok {
				return rdf.ReadDir(-1)
			}
			return nil, fs.ErrInvalid
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return nil, os.ErrNotExist
}

// ---- fs.SubFS ----
func (f *FsFallback) Sub(dir string) (fs.FS, error) {
	var subList []fs.FS
	for _, filesystem := range f.FSList {
		if sub, ok := filesystem.(fs.SubFS); ok {
			subfs, err := sub.Sub(dir)
			if err == nil {
				subList = append(subList, subfs)
				continue
			}
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}
	if len(subList) == 0 {
		return nil, os.ErrNotExist
	}
	return &FsFallback{FSList: subList}, nil
}
