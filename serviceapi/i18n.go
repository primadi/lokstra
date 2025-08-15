package serviceapi

type I18n interface {
	T(lang, code string, params map[string]any) string
}
