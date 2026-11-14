// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package i18n

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/flamego/flamego"
	"go.yaml.in/yaml/v3"
)

func readLocale(t *testing.T, name string) map[string]map[string]string {
	t.Helper()
	data, err := localeFiles.ReadFile("locales/" + name + ".yaml")
	if err != nil {
		t.Fatal(err)
	}
	var sections map[string]map[string]string
	if err := yaml.Unmarshal(data, &sections); err != nil {
		t.Fatal(err)
	}
	return sections
}

var verbPattern = regexp.MustCompile(`%[-+# 0]*\d*(?:\.\d+)?[a-zA-Z%]`)

// Every locale must define the same keys with the same format verbs as the default one.
func TestLocalesMatch(t *testing.T) {
	base := readLocale(t, DefaultLanguage)
	for _, language := range Languages[1:] {
		other := readLocale(t, language.Name)
		for section, keys := range base {
			for key, value := range keys {
				translated, ok := other[section][key]
				if !ok {
					t.Errorf("%s: missing %s::%s", language.Name, section, key)
					continue
				}
				if want, got := verbPattern.FindAllString(value, -1), verbPattern.FindAllString(translated, -1); strings.Join(want, " ") != strings.Join(got, " ") {
					t.Errorf("%s: %s::%s has verbs %v, want %v", language.Name, section, key, got, want)
				}
			}
		}
		for section, keys := range other {
			for key := range keys {
				if _, ok := base[section][key]; !ok {
					t.Errorf("%s: unknown key %s::%s", language.Name, section, key)
				}
			}
		}
	}
}

// Every "section::key" literal in the server code must exist in the default locale.
func TestKeysUsedInCode(t *testing.T) {
	base := readLocale(t, DefaultLanguage)
	pattern := regexp.MustCompile(`"([a-z_]+)::([a-z_]+)"`)
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range pattern.FindAllStringSubmatch(string(data), -1) {
			if match[1] == "section" {
				continue // Placeholder in doc comments.
			}
			if _, ok := base[match[1]][match[2]]; !ok {
				t.Errorf("%s: %s::%s is not defined in %s.yaml", path, match[1], match[2], DefaultLanguage)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnglishAndMessages(t *testing.T) {
	if got := English().Translate("execution::invalid_status", "bad"); got != `Invalid status "bad"` {
		t.Fatalf("unexpected translation %q", got)
	}
	if got := English().Translate("validation::username"); !strings.Contains(got, "'_', '.' or '-'") {
		t.Fatalf("quotes must survive the INI conversion, got %q", got)
	}

	err := Errorf("validation::password_length", 8, 72)
	if err.Error() != "Password must contain 8-72 UTF-8 bytes" {
		t.Fatalf("unexpected English message %q", err.Error())
	}
	if got := TranslateError(nil, err); got != err.Error() {
		t.Fatalf("nil locale must keep the English text, got %q", got)
	}
	if got := TranslateError(English(), errors.New("plain")); got != "plain" {
		t.Fatalf("plain errors must keep their text, got %q", got)
	}
}

func TestMiddleware(t *testing.T) {
	f := flamego.New()
	f.Use(I18n())
	f.Get("/", func(c flamego.Context, l Locale) string {
		message := TranslateError(l, Errorf("validation::password_length", 8, 72))
		return l.Lang() + "|" + l.Translate("auth::invalid_credentials") + "|" + message
	})

	for _, test := range []struct {
		name, query, cookie, acceptLanguage string
		want                                string
	}{
		{"default", "", "", "", "en-US|Invalid username or password|Password must contain 8-72 UTF-8 bytes"},
		{"accept language", "", "", "zh-CN,zh;q=0.9", "zh-CN|用户名或密码错误|密码长度须为 8-72 个 UTF-8 字节"},
		{"unsupported language", "", "", "fr-FR", "en-US|Invalid username or password|Password must contain 8-72 UTF-8 bytes"},
		{"query wins over cookie", "?lang=en-US", "zh-CN", "zh-CN", "en-US|Invalid username or password|Password must contain 8-72 UTF-8 bytes"},
		{"cookie wins over header", "", "zh-CN", "en-US", "zh-CN|用户名或密码错误|密码长度须为 8-72 个 UTF-8 字节"},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/"+test.query, nil)
			if test.cookie != "" {
				r.AddCookie(&http.Cookie{Name: "lang", Value: test.cookie})
			}
			if test.acceptLanguage != "" {
				r.Header.Set("Accept-Language", test.acceptLanguage)
			}
			w := httptest.NewRecorder()
			f.ServeHTTP(w, r)
			if got := w.Body.String(); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
			if test.cookie == "" && w.Header().Get("Set-Cookie") == "" {
				t.Fatal("the chosen language should be remembered in a cookie")
			}
		})
	}
}
