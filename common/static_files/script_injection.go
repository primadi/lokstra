package static_files

import (
	"embed"
	"io"
	"io/fs"
	"strings"
)

type InjectLocation string

const (
	InjectLocationHeadStart InjectLocation = "head_start"
	InjectLocationHeadEnd   InjectLocation = "head_end"
	InjectLocationBodyEnd   InjectLocation = "body_end"
)

type ScriptData struct {
	Location InjectLocation
	Script   string
}

type ScriptInjection struct {
	LayoutScripts []*ScriptData
}

//go:embed scripts
var emScript embed.FS

var fsScript fs.FS

func init() {
	fsScript, _ = fs.Sub(emScript, "scripts")
}

func getInjectLocation(fileName string) InjectLocation {
	switch {
	case strings.HasPrefix(fileName, "body_end"):
		return InjectLocationBodyEnd
	case strings.HasPrefix(fileName, "head_start"):
		return InjectLocationHeadStart
	case strings.HasPrefix(fileName, "head_end"):
		return InjectLocationHeadEnd
	default:
		return InjectLocationBodyEnd
	}
}

// Creates a new empty script injection
func NewScriptInjection() *ScriptInjection {
	return &ScriptInjection{
		LayoutScripts: []*ScriptData{},
	}
}

// Adds all scripts from the embedded "scripts/<name>" directory
func (si *ScriptInjection) AddNamedScriptInjection(name string) *ScriptInjection {
	return si.AddCustomNamedScriptInjection(fsScript, name)
}

// Adds all scripts from the provided fs.FS at "<name>" directory
func (si *ScriptInjection) AddCustomNamedScriptInjection(fsScript fs.FS, name string) *ScriptInjection {
	files, _ := fs.ReadDir(fsScript, name)
	for _, file := range files {
		fName := file.Name()
		if file.IsDir() || !strings.HasSuffix(fName, ".js") {
			continue
		}
		if fsFile, err := fsScript.Open(name + "/" + fName); err == nil {
			if content, err := io.ReadAll(fsFile); err == nil {
				si.LayoutScripts = append(si.LayoutScripts, &ScriptData{
					Location: getInjectLocation(fName),
					Script:   string(content),
				})
			}
		}
	}
	return si
}

// Creates a default script injection with optional animation scripts
func NewDefaultScriptInjection(enableAnimation bool) *ScriptInjection {
	if enableAnimation {
		return NewScriptInjection().AddNamedScriptInjection("default").
			AddNamedScriptInjection("animation")
	}
	return NewScriptInjection().AddNamedScriptInjection("default")
}

// Adds a custom script to be injected at the end of the body
func (si *ScriptInjection) AddBodyEndScript(script string) *ScriptInjection {
	si.LayoutScripts = append(si.LayoutScripts, &ScriptData{
		Location: InjectLocationBodyEnd,
		Script:   script,
	})
	return si
}

// Adds a custom script to be injected at the start of the head
func (si *ScriptInjection) AddHeadStartScript(script string) *ScriptInjection {
	si.LayoutScripts = append(si.LayoutScripts, &ScriptData{
		Location: InjectLocationHeadStart,
		Script:   script,
	})
	return si
}

// Adds a custom script to be injected at the end of the head
func (si *ScriptInjection) AddHeadEndScript(script string) *ScriptInjection {
	si.LayoutScripts = append(si.LayoutScripts, &ScriptData{
		Location: InjectLocationHeadEnd,
		Script:   script,
	})
	return si
}

// Loads the injection scripts into the provided layout content
func (si *ScriptInjection) LoadInjectionScripts(strLayoutContent string) string {
	var headStartInjection, headEndInjection, bodyEndInjection string

	for _, script := range si.LayoutScripts {
		switch script.Location {
		case InjectLocationHeadStart:
			headStartInjection += "<script>" + script.Script + "\n</script>\n"
		case InjectLocationHeadEnd:
			headEndInjection += "<script>" + script.Script + "\n</script>\n"
		case InjectLocationBodyEnd:
			bodyEndInjection += "<script>" + script.Script + "\n</script>\n"
		}
	}

	if headStartInjection != "" {
		strLayoutContent = strings.Replace(strLayoutContent, "<head>",
			"<head>\n"+headStartInjection, 1)
	}

	if headEndInjection != "" {
		strLayoutContent = strings.Replace(strLayoutContent, "</head>",
			headEndInjection+"</head>", 1)
	}

	if bodyEndInjection != "" {
		strLayoutContent = strings.Replace(strLayoutContent, "</body>",
			bodyEndInjection+"</body>", 1)
	}

	return strLayoutContent
}
