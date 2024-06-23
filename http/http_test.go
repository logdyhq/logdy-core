package http

import (
	"testing"
)

func TestNormalizeHttpPrefix(t *testing.T) {

	cases := []struct {
		p string
		a string
	}{
		{p: "", a: "/"},
		{p: "/log", a: "/log/"},
		{p: "log", a: "/log/"},
		{p: "//log", a: "//log/"},
	}

	for _, tt := range cases {

		cfg := &Config{
			HttpPathPrefix: tt.p,
		}
		normalizeHttpPathPrefix(cfg)
		if cfg.HttpPathPrefix != tt.a {
			t.Errorf("expected %s, actual %s", tt.a, cfg.HttpPathPrefix)
		}
	}
}
