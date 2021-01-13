package main

import "testing"

func TestParseArg(t *testing.T) {
	tt := map[string][]string{
		"":                                {"", "", "", ""},
		"stuart-warren":                   {"", "", "stuart-warren", ""},
		"stuart-warren/.dotfiles":         {"", "", "stuart-warren", ".dotfiles"},
		"bob":                             {"", "", "bob", ""},
		"https://gitlab.com/bob/dotfiles": {"https", "gitlab.com", "bob", "dotfiles"},
	}

	for arg, expected := range tt {
		scheme, host, org, name := parseRepo(arg)
		if scheme != expected[0] {
			t.Errorf("scheme - got: %q, expected: %q", scheme, expected[0])
		}
		if host != expected[1] {
			t.Errorf("host - got: %q, expected: %q", host, expected[1])
		}
		if org != expected[2] {
			t.Errorf("org - got: %q, expected: %q", org, expected[2])
		}
		if name != expected[3] {
			t.Errorf("name - got: %q, expected: %q", name, expected[3])
		}
	}
}

func TestBuildRepo(t *testing.T) {
	tt := map[string]string{
		"":                                "https://github.com/stuart-warren/dotfiles",
		"stuart-warren":                   "https://github.com/stuart-warren/dotfiles",
		"stuart-warren/.dotfiles":         "https://github.com/stuart-warren/.dotfiles",
		"bob":                             "https://github.com/bob/dotfiles",
		"https://gitlab.com/bob/dotfiles": "https://gitlab.com/bob/dotfiles",
	}

	for arg, expected := range tt {
		out := buildRepo(arg)
		if out.String() != expected {
			t.Errorf("got: %s, expected: %s", out.String(), expected)
		}
	}
}
