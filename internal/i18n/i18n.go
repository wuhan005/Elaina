// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package i18n serves API messages from the embedded YAML locales through the flamego i18n middleware.
package i18n

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/flamego/flamego"
	flamegoi18n "github.com/flamego/i18n"
	"go.yaml.in/yaml/v3"
	"unknwon.dev/i18n"
)

//go:embed locales/*.yaml
var localeFiles embed.FS

// Locale translates "section::key" messages; the middleware maps it for every request.
type Locale = flamegoi18n.Locale

const DefaultLanguage = "en-US"

var Languages = []flamegoi18n.Language{
	{Name: "en-US", Description: "English"},
	{Name: "zh-CN", Description: "简体中文"},
}

var (
	iniFiles memFS
	english  Locale
)

func init() {
	files, err := loadLocales()
	if err != nil {
		panic("i18n: " + err.Error())
	}
	iniFiles = files

	locale, err := i18n.NewStore().AddLocale(DefaultLanguage, Languages[0].Description, []byte(files[DefaultLanguage+".ini"]))
	if err != nil {
		panic("i18n: load fallback locale: " + err.Error())
	}
	english = locale
}

// loadLocales renders each YAML locale as the INI file flamego/i18n reads.
func loadLocales() (memFS, error) {
	files := make(memFS, len(Languages))
	for _, language := range Languages {
		data, err := localeFiles.ReadFile("locales/" + language.Name + ".yaml")
		if err != nil {
			return nil, err
		}
		var sections map[string]map[string]string
		if err := yaml.Unmarshal(data, &sections); err != nil {
			return nil, fmt.Errorf("parse %s.yaml: %w", language.Name, err)
		}
		rendered, err := renderINI(sections)
		if err != nil {
			return nil, fmt.Errorf("%s.yaml: %w", language.Name, err)
		}
		files[language.Name+".ini"] = rendered
	}
	return files, nil
}

func renderINI(sections map[string]map[string]string) (string, error) {
	var b strings.Builder
	for _, section := range sortedKeys(sections) {
		fmt.Fprintf(&b, "[%s]\n", section)
		for _, key := range sortedKeys(sections[section]) {
			value := sections[section][key]
			// Backquotes keep ';', '#' and quotes literal; they cannot appear inside the value.
			if strings.ContainsAny(value, "`\n") {
				return "", fmt.Errorf("%s.%s: values must not contain backquotes or newlines", section, key)
			}
			fmt.Fprintf(&b, "%s = `%s`\n", key, value)
		}
	}
	return b.String(), nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// I18n returns the middleware that picks the locale from ?lang=, the lang cookie, then Accept-Language.
func I18n() flamego.Handler {
	return flamegoi18n.I18n(flamegoi18n.Options{
		FileSystem: http.FS(iniFiles),
		Languages:  Languages,
		Default:    DefaultLanguage,
		NameFormat: "%s.ini",
	})
}

// English returns the default locale, for code paths without the middleware.
func English() Locale {
	return english
}

// Chinese reports whether the locale is Chinese.
func Chinese(l Locale) bool {
	return l != nil && strings.HasPrefix(l.Lang(), "zh")
}

// Message is an error whose text comes from the locale files.
type Message struct {
	Key  string
	Args []any
}

func Errorf(key string, args ...any) error {
	return &Message{Key: key, Args: args}
}

func (m *Message) Error() string {
	return english.Translate(m.Key, m.Args...)
}

// TranslateError renders a Message in the locale; other errors keep their own text.
func TranslateError(l Locale, err error) string {
	var message *Message
	if l != nil && errors.As(err, &message) {
		return l.Translate(message.Key, message.Args...)
	}
	return err.Error()
}

// memFS is a read-only in-memory file system of rendered INI files.
type memFS map[string]string

func (m memFS) Open(name string) (fs.File, error) {
	data, ok := m[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return &memFile{Reader: bytes.NewReader([]byte(data)), name: name, size: int64(len(data))}, nil
}

type memFile struct {
	*bytes.Reader
	name string
	size int64
}

func (f *memFile) Stat() (fs.FileInfo, error) { return f, nil }
func (f *memFile) Close() error               { return nil }
func (f *memFile) Name() string               { return f.name }
func (f *memFile) Size() int64                { return f.size }
func (f *memFile) Mode() fs.FileMode          { return 0o444 }
func (f *memFile) ModTime() time.Time         { return time.Time{} }
func (f *memFile) IsDir() bool                { return false }
func (f *memFile) Sys() any                   { return nil }
