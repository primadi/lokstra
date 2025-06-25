package iface

// HasSetting defines read-only access to a setting map.
type HasSetting interface {
	GetSetting(key string) any
}

// SettingAccessor provides helpers for working with setting maps.
type SettingAccessor struct {
	Settings map[string]any
}

func (s *SettingAccessor) SetSetting(key string, value any) {
	if s.Settings == nil {
		s.Settings = make(map[string]any)
	}
	s.Settings[key] = value
}

func (s *SettingAccessor) GetSetting(key string) any {
	return s.Settings[key]
}

func (s *SettingAccessor) GetString(key string, def string) string {
	if val, ok := s.Settings[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return def
}

func (s *SettingAccessor) GetBool(key string, def bool) bool {
	if val, ok := s.Settings[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return def
}

func (s *SettingAccessor) GetInt(key string, def int) int {
	if val, ok := s.Settings[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return def
}
