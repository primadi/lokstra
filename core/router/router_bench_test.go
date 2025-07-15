package router

import (
	"lokstra/common/meta"
	"strings"
	"testing"
)

func BenchmarkCleanPrefix(b *testing.B) {
	router := &RouterImpl{
		meta: &meta.RouterMeta{
			Prefix: "/api/v1",
		},
	}

	testCases := []string{
		"users",
		"/users",
		"users/",
		"/users/",
		"users/profile",
		"/users/profile/",
		"very/long/path/with/many/segments",
		"api/v1/users/profile/settings/notifications",
		"extremely/long/path/with/many/segments/that/would/benefit/from/builder",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, prefix := range testCases {
			_ = router.cleanPrefix(prefix)
		}
	}
}

func BenchmarkCleanPrefixOld(b *testing.B) {
	router := &RouterImpl{
		meta: &meta.RouterMeta{
			Prefix: "/api/v1",
		},
	}

	cleanPrefixOld := func(r *RouterImpl, prefix string) string {
		if prefix == "/" || prefix == "" {
			return r.meta.Prefix
		}

		if r.meta.Prefix == "/" {
			return "/" + strings.Trim(prefix, "/")
		}
		return r.meta.Prefix + "/" + strings.Trim(prefix, "/")
	}

	testCases := []string{
		"users",
		"/users",
		"users/",
		"/users/",
		"users/profile",
		"/users/profile/",
		"very/long/path/with/many/segments",
		"api/v1/users/profile/settings/notifications",
		"extremely/long/path/with/many/segments/that/would/benefit/from/builder",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, prefix := range testCases {
			_ = cleanPrefixOld(router, prefix)
		}
	}
}
