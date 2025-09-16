package static_files

import (
	"embed"
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

func NewScriptInjection() *ScriptInjection {
	return &ScriptInjection{
		LayoutScripts: []*ScriptData{},
	}
}

func (si *ScriptInjection) AddNamedScriptInjection(name string) *ScriptInjection {
	namedPath := "scripts/" + name
	files, _ := emScript.ReadDir(namedPath)
	for _, file := range files {
		fName := file.Name()
		if file.IsDir() || !strings.HasSuffix(fName, ".js") {
			continue
		}
		if content, err := emScript.ReadFile(namedPath + "/" + fName); err == nil {
			si.LayoutScripts = append(si.LayoutScripts, &ScriptData{
				Location: getInjectLocation(fName),
				Script:   string(content),
			})
		}
	}
	return si
}

func NewDefaultScriptInjection(enableAnimation bool) *ScriptInjection {
	if enableAnimation {
		return NewScriptInjection().AddNamedScriptInjection("default").
			AddNamedScriptInjection("animation")
	}
	return NewScriptInjection().AddNamedScriptInjection("default")
}

func (si *ScriptInjection) AddBodyEndScript(script string) *ScriptInjection {
	si.LayoutScripts = append(si.LayoutScripts, &ScriptData{
		Location: InjectLocationBodyEnd,
		Script:   script,
	})
	return si
}

func (si *ScriptInjection) AddHeadStartScript(script string) *ScriptInjection {
	si.LayoutScripts = append(si.LayoutScripts, &ScriptData{
		Location: InjectLocationHeadStart,
		Script:   script,
	})
	return si
}

func (si *ScriptInjection) AddHeadEndScript(script string) *ScriptInjection {
	si.LayoutScripts = append(si.LayoutScripts, &ScriptData{
		Location: InjectLocationHeadEnd,
		Script:   script,
	})
	return si
}

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
