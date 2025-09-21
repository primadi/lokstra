package htmx_fsmanager

import "io/fs"

type IContainer interface {
	GetHtmxFsManager() *HtmxFsManager
}

type HtmxFsManager struct {
	LayoutsFiles []fs.FS
	PagesFiles   []fs.FS
	StaticFiles  []fs.FS
}

// Create a sub filesystem from a given fs.FS and directory
func SubEmbedFS(fsys fs.FS, dir string) (fs.FS, error) {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// Create a new HtmxFsManager instance
func New() *HtmxFsManager {
	return &HtmxFsManager{}
}

// Combine two HtmxFsManager instances by appending their file systems
func (fm *HtmxFsManager) Merge(other *HtmxFsManager) *HtmxFsManager {
	if other == nil {
		return fm
	}

	merged := &HtmxFsManager{
		LayoutsFiles: append(fm.LayoutsFiles, other.LayoutsFiles...),
		PagesFiles:   append(fm.PagesFiles, other.PagesFiles...),
		StaticFiles:  append(fm.StaticFiles, other.StaticFiles...),
	}

	return merged
}

// Add file systems for Layouts to the manager
func (fm *HtmxFsManager) AddLayoutFiles(fsList ...fs.FS) *HtmxFsManager {
	fm.LayoutsFiles = append(fm.LayoutsFiles, fsList...)
	return fm
}

// Add file systems for Pages to the manager
func (fm *HtmxFsManager) AddPageFiles(fsList ...fs.FS) *HtmxFsManager {
	fm.PagesFiles = append(fm.PagesFiles, fsList...)
	return fm
}

// Add file systems for Static files to the manager
func (fm *HtmxFsManager) AddStaticFiles(fsList ...fs.FS) *HtmxFsManager {
	fm.StaticFiles = append(fm.StaticFiles, fsList...)
	return fm
}

// Reads a layout file by searching through the registered layout filesystems
func (fm *HtmxFsManager) ReadLayoutFile(name string) ([]byte, error) {
	var err error
	for _, fSys := range fm.LayoutsFiles {
		var data []byte
		data, err = fs.ReadFile(fSys, name)
		if err == nil {
			return data, nil
		}
	}
	return nil, err
}

// Reads a page file by searching through the registered page filesystems
func (fm *HtmxFsManager) ReadPageFile(name string) ([]byte, error) {
	var err error
	for _, fSys := range fm.PagesFiles {
		var data []byte
		data, err = fs.ReadFile(fSys, name)
		if err == nil {
			return data, nil
		}
	}
	return nil, err
}

// Reads a static file by searching through the registered static filesystems
func (fm *HtmxFsManager) ReadStaticFile(name string) ([]byte, error) {
	var err error
	for _, fSys := range fm.StaticFiles {
		var data []byte
		data, err = fs.ReadFile(fSys, name)
		if err == nil {
			return data, nil
		}
	}
	return nil, err
}
