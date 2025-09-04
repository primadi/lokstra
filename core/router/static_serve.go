package router

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/primadi/lokstra/common/file_system"
)

type StaticSource struct {
	FromDisk bool
	RootPath string
	FS       fs.FS
}

func (s StaticSource) Handler() http.Handler {
	if s.FromDisk {
		return http.FileServer(http.Dir(s.RootPath))
	}
	return http.FileServer(http.FS(s.FS))
}

type StaticFallback struct {
	Servers []StaticSource
}

// NewStaticFallback creates a new StaticFallback instance.
// sources can be string, http.Dir, embed.FS, or fs.FS
func NewStaticFallback(sources ...any) (*StaticFallback, error) {
	sf := &StaticFallback{}
	for _, src := range sources {
		switch s := src.(type) {
		case http.Dir:
			sf.Servers = append(sf.Servers, StaticSource{FromDisk: true, RootPath: string(s)})
		case string:
			sf.Servers = append(sf.Servers, StaticSource{FromDisk: true, RootPath: s})
		case file_system.FSWithSubFS:
			subFs, err := fs.Sub(s.FS, s.SubFS)
			if err != nil {
				return nil, fmt.Errorf("failed to create sub FS from %s: %v", s.SubFS, err)
			}
			sf.Servers = append(sf.Servers, StaticSource{FromDisk: false, FS: subFs})
		case embed.FS:
			// embed.FS always needs subFirstDir to remove the embed prefix
			subFS, err := file_system.SubFirstCommonFolder(s)
			if err != nil {
				return nil, fmt.Errorf("failed to process embed.FS: %v", err)
			}
			sf.Servers = append(sf.Servers, StaticSource{FromDisk: false, FS: subFS})
		case fs.FS:
			// Regular fs.FS doesn't need subFirstDir
			sf.Servers = append(sf.Servers, StaticSource{FromDisk: false, FS: s})
		default:
			return nil, fmt.Errorf("unsupported source type for StaticFallback: %T", s)
		}
	}
	return sf, nil
}

func (sf StaticFallback) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Remove leading slash for file system access
		if path != "" && path[0] == '/' {
			path = path[1:]
		}
		if path == "" {
			path = "index.html"
		}

		// Try each source in order
		for i, s := range sf.Servers {
			if s.FromDisk {
				// Try to open file from disk
				file, err := http.Dir(s.RootPath).Open(path)
				if err == nil {
					file.Close()
					// File exists, serve it
					fmt.Printf("[DEBUG] Serving %s from disk source %d: %s\n", path, i, s.RootPath)
					http.FileServer(http.Dir(s.RootPath)).ServeHTTP(w, r)
					return
				} else {
					fmt.Printf("[DEBUG] Not found in disk source %d (%s): %v\n", i, s.RootPath, err)
				}
			} else {
				// Try to open file from embed.FS
				_, err := s.FS.Open(path)
				if err == nil {
					// File exists, serve it
					fmt.Printf("[DEBUG] Serving %s from embed.FS source %d\n", path, i)
					http.FileServer(http.FS(s.FS)).ServeHTTP(w, r)
					return
				} else {
					fmt.Printf("[DEBUG] Not found in embed.FS source %d: %v\n", i, err)
				}
			}
		}

		// Not found in any source
		fmt.Printf("[DEBUG] File %s not found in any source, returning 404\n", path)
		http.NotFound(w, r)
	})
}

// Removed responseRecorder as it's not needed with the new implementation

func (sf *StaticFallback) AddRootDir(root string) {
	sf.Servers = append(sf.Servers, StaticSource{FromDisk: true, RootPath: root})
}

func (sf *StaticFallback) AddFS(fs fs.FS) {
	sf.Servers = append(sf.Servers, StaticSource{FromDisk: false, FS: fs})
}

func (sf *StaticFallback) AddFSWithSubFS(fsys fs.FS, subFS string) {
	fs, err := fs.Sub(fsys, subFS)
	if err != nil {
		fmt.Printf("Failed to create sub FS from %s: %v\n", subFS, err)
		return
	}
	sf.Servers = append(sf.Servers, StaticSource{FromDisk: false, FS: fs})
}
