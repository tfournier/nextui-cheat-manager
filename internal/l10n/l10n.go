// Package l10n holds the UI translations. Messages are declared in
// messages.go with their English text; translations live in locales/*.toml
// and are embedded in the binary.
package l10n

import (
	"embed"
	"fmt"
	"io/fs"
	"path"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/i18n"
)

//go:embed locales/*.toml
var localeFS embed.FS

// Language is a supported UI language.
type Language struct {
	Code    string // BCP 47 code, also the locale file suffix
	Name    string // native name
	Welcome string // welcome text in that language
}

// Languages lists the supported languages, in display order.
var Languages = []Language{
	{"en", "English", "Welcome to Cheat Manager!"},
	{"de", "Deutsch", "Willkommen bei Cheat Manager!"},
	{"es", "Español", "¡Bienvenido a Cheat Manager!"},
	{"fr", "Français", "Bienvenue dans Cheat Manager !"},
	{"it", "Italiano", "Benvenuto in Cheat Manager!"},
	{"pt", "Português", "Bem-vindo ao Cheat Manager!"},
	{"ja", "日本語", "Cheat Managerへようこそ！"},
}

// DefaultLanguage is used when no language has been chosen yet.
const DefaultLanguage = "en"

// IsSupported reports whether code is one of Languages.
func IsSupported(code string) bool {
	for _, l := range Languages {
		if l.Code == code {
			return true
		}
	}
	return false
}

// MessageFiles returns the embedded locale files.
func MessageFiles() ([]i18n.MessageFile, error) {
	entries, err := fs.ReadDir(localeFS, "locales")
	if err != nil {
		return nil, err
	}
	var files []i18n.MessageFile
	for _, e := range entries {
		content, err := localeFS.ReadFile(path.Join("locales", e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read locale %s: %w", e.Name(), err)
		}
		files = append(files, i18n.MessageFile{Name: e.Name(), Content: content})
	}
	return files, nil
}

// Init loads the translations and selects the given language.
func Init(code string) error {
	files, err := MessageFiles()
	if err != nil {
		return err
	}
	if err := i18n.InitI18NFromBytes(files); err != nil {
		return err
	}
	return SetLanguage(code)
}

// SetLanguage switches the UI language (falls back to English if unsupported).
func SetLanguage(code string) error {
	if !IsSupported(code) {
		code = DefaultLanguage
	}
	return i18n.SetWithCode(code)
}

// T returns the message in the current language.
func T(m *i18n.Message) string {
	return i18n.Localize(m, nil)
}

// Tf returns the message with template values, e.g. Tf(MsgNoSystem, "Path", p).
func Tf(m *i18n.Message, kv ...any) string {
	return i18n.Localize(m, data(kv))
}

// Tn returns the plural form of the message for count. {{.Count}} is set.
func Tn(m *i18n.Message, count int, kv ...any) string {
	d := data(kv)
	d["Count"] = count
	return i18n.LocalizePlural(m, count, d)
}

func data(kv []any) map[string]any {
	d := make(map[string]any, len(kv)/2+1)
	for i := 0; i+1 < len(kv); i += 2 {
		d[fmt.Sprint(kv[i])] = kv[i+1]
	}
	return d
}
