// Cheat Manager - a NextUI tool to add, edit, toggle, delete and import
// RetroArch cheats (.cht) directly on the device.
package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"

	"github.com/tfournier/nextui-cheat-manager/internal/cheats"
	"github.com/tfournier/nextui-cheat-manager/internal/l10n"
	"github.com/tfournier/nextui-cheat-manager/internal/settings"
)

// Short aliases for the translation helpers.
var (
	T  = l10n.T
	Tf = l10n.Tf
	Tn = l10n.Tn
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

var (
	sdcard     = envOr("SDCARD_PATH", "/mnt/SDCARD")
	romsPath   = envOr("ROMS_PATH", filepath.Join(sdcard, "Roms"))
	cheatsPath = envOr("CHEATS_PATH", filepath.Join(sdcard, "Cheats"))
	platform   = envOr("PLATFORM", "tg5040")
	userdata   = envOr("USERDATA_PATH", filepath.Join(sdcard, ".userdata", platform))
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	logDir := envOr("LOGS_PATH", filepath.Join(userdata, "logs"))
	gaba.Init(gaba.Options{
		WindowTitle:    "Cheat Manager",
		ShowBackground: true,
		// CHEATS_DESKTOP=1 runs the UI outside of the device (keyboard input, no NextUI theme).
		IsNextUI: os.Getenv("CHEATS_DESKTOP") == "",
		LogPath:  filepath.Join(logDir, "Cheat Manager.txt"),
	})
	defer gaba.Close()

	slog.Info("Cheat Manager started", "version", version, "roms", romsPath, "cheats", cheatsPath)
	setupLanguage()
	_ = os.MkdirAll(cheatsPath, 0o755)
	selectSystem()
}

// ---------------------------------------------------------------------------
// Language

var settingsPath = settings.Path(userdata)

// setupLanguage loads the saved language, or asks for it on first launch.
func setupLanguage() {
	conf, err := settings.Load(settingsPath)
	if err != nil {
		slog.Warn("cannot read settings, using defaults", "path", settingsPath, "error", err)
	}
	if err := l10n.Init(l10n.DefaultLanguage); err != nil {
		slog.Error("cannot load translations", "error", err)
	}
	if l10n.IsSupported(conf.Language) {
		_ = l10n.SetLanguage(conf.Language)
		return
	}
	// First launch: the choice cannot be skipped.
	chooseLanguage(conf, true)
}

// chooseLanguage shows the language picker, applies and saves the choice.
func chooseLanguage(conf settings.Settings, firstLaunch bool) {
	options := make([]gaba.SelectionOption, len(l10n.Languages))
	initial := 0
	for i, l := range l10n.Languages {
		options[i] = gaba.SelectionOption{DisplayName: l.Name, Description: l.Welcome, Value: l.Code}
		if l.Code == conf.Language {
			initial = i
		}
	}
	title := T(l10n.LanguagePrompt)
	help := footer(constants.LeftRight, T(l10n.BtnSelect), "A", T(l10n.BtnConfirm))
	if firstLaunch {
		// Nothing is chosen yet: show the native welcome texts, in English.
		title = l10n.Languages[0].Welcome
		help = footer(constants.LeftRight, "Select", "A", "Confirm")
	} else {
		help = append(footer("B", T(l10n.BtnCancel)), help...)
	}

	r, err := gaba.SelectionMessage(title, options, help, gaba.SelectionMessageSettings{
		DisableBackButton: firstLaunch,
		InitialSelection:  initial,
	})
	if err != nil || r == nil {
		if firstLaunch {
			_ = l10n.SetLanguage(l10n.DefaultLanguage)
		}
		return
	}
	code, _ := r.SelectedValue.(string)
	_ = l10n.SetLanguage(code)
	conf.Language = code
	if err := settings.Save(settingsPath, conf); err != nil {
		slog.Error("cannot save settings", "path", settingsPath, "error", err)
	}
	slog.Info("language selected", "language", code)
}

// ---------------------------------------------------------------------------
// UI helpers

// pos remembers the position in a list, to come back to the same place.
type pos struct{ sel, start int }

func (p *pos) update(r *gaba.ListResult) {
	if len(r.Selected) > 0 {
		p.sel = r.Selected[0]
		p.start = max(p.sel-r.VisiblePosition, 0)
	}
}

func footer(items ...string) []gaba.FooterHelpItem {
	var out []gaba.FooterHelpItem
	for i := 0; i+1 < len(items); i += 2 {
		out = append(out, gaba.FooterHelpItem{ButtonName: items[i], HelpText: items[i+1]})
	}
	return out
}

// Extra list buttons.
type listButtons int

const (
	noExtra listButtons = iota
	withX               // X triggers gaba.ListActionTriggered
	withY               // Y triggers gaba.ListActionSecondaryTriggered
)

// showList displays a list and returns the selected index and action, or
// gaba.ErrCancelled when the user goes back.
func showList(title string, labels []string, p *pos, empty string, extra listButtons, help ...string) (int, gaba.ListAction, error) {
	items := make([]gaba.MenuItem, len(labels))
	for i, l := range labels {
		items[i] = gaba.MenuItem{Text: l}
	}
	opts := gaba.DefaultListOptions(title, items)
	if p.sel >= len(items) {
		p.sel, p.start = 0, 0
	}
	opts.SelectedIndex = p.sel
	opts.VisibleStartIndex = p.start
	opts.EmptyMessage = empty
	opts.FooterHelpItems = footer(help...)
	switch extra {
	case withX:
		opts.ActionButton = constants.VirtualButtonX
	case withY:
		opts.SecondaryActionButton = constants.VirtualButtonY
	}
	r, err := gaba.List(opts)
	if err != nil {
		return -1, 0, err
	}
	p.update(r)
	if len(r.Selected) == 0 && r.Action != gaba.ListActionSecondaryTriggered {
		return -1, r.Action, gaba.ErrCancelled
	}
	if len(r.Selected) == 0 {
		return -1, r.Action, nil
	}
	return r.Selected[0], r.Action, nil
}

func message(text string) {
	_, _ = gaba.ConfirmationMessage(text, footer("A", T(l10n.BtnOK)), gaba.MessageOptions{})
}

func confirm(text, okLabel string) bool {
	r, err := gaba.ConfirmationMessage(text, footer("B", T(l10n.BtnCancel), "A", okLabel), gaba.MessageOptions{})
	return err == nil && r != nil && r.Confirmed
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func save(path string, cf *cheats.File) bool {
	if err := cheats.Save(path, cf); err != nil {
		slog.Error("save failed", "path", path, "error", err)
		message(Tf(l10n.MsgWriteError, "Error", err))
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// Screens

func selectSystem() {
	systems := cheats.ListSystems(romsPath)
	if len(systems) == 0 {
		message(Tf(l10n.MsgNoSystem, "Path", romsPath))
		return
	}
	labels := make([]string, len(systems))
	for i, s := range systems {
		labels[i] = s.Label
	}
	var p pos
	for {
		i, action, err := showList("Cheat Manager", labels, &p, "", withY,
			"B", T(l10n.BtnQuit), "Y", T(l10n.BtnLanguage), "A", T(l10n.BtnSelect))
		if err != nil {
			return
		}
		if action == gaba.ListActionSecondaryTriggered {
			conf, _ := settings.Load(settingsPath)
			chooseLanguage(conf, false)
			continue
		}
		s := systems[i]
		browse(s.Dir, s.Tag, s.Label)
	}
}

func browse(dir, tag, title string) {
	cheatDir := filepath.Join(cheatsPath, tag)
	var p pos
	for {
		entries := cheats.ListEntries(dir, cheatDir)
		labels := make([]string, len(entries))
		for i, e := range entries {
			switch {
			case e.IsDir:
				labels[i] = e.Name + "/"
			case e.Cheats > 0:
				labels[i] = fmt.Sprintf("%s  [%d]", e.Name, e.Cheats)
			default:
				labels[i] = e.Name
			}
		}
		i, _, err := showList(title, labels, &p, T(l10n.MsgNoGame), noExtra, "B", T(l10n.BtnBack), "A", T(l10n.BtnSelect))
		if err != nil {
			return
		}
		e := entries[i]
		if e.IsDir {
			browse(filepath.Join(dir, e.Name), tag, e.Name)
		} else {
			manageCheats(cheats.CheatPath(cheatsPath, tag, e.Name), e.Name, tag)
		}
	}
}

func manageCheats(path, rom, tag string) {
	var p pos
	for {
		cf, err := cheats.Load(path)
		if err != nil {
			message(Tf(l10n.MsgReadError, "Error", err))
			return
		}
		labels := []string{T(l10n.ItemAddCheat), T(l10n.ItemImport)}
		for _, c := range cf.Cheats {
			state := "[OFF]"
			if c.Enable {
				state = "[ON] "
			}
			d := c.Desc
			if d == "" {
				d = T(l10n.ItemNoName)
			}
			labels = append(labels, state+" "+d)
		}
		hasFile := fileExists(path)
		if hasFile {
			labels = append(labels, T(l10n.ItemDeleteAll))
		}

		i, action, err := showList(rom, labels, &p, "", withX,
			"B", T(l10n.BtnBack), "X", T(l10n.BtnToggle), "A", T(l10n.BtnSelect))
		if err != nil {
			return
		}
		ci := i - 2 // cheat index (the first two items are actions)
		isCheat := ci >= 0 && ci < len(cf.Cheats)

		if action == gaba.ListActionTriggered {
			if isCheat {
				cf.Cheats[ci].Enable = !cf.Cheats[ci].Enable
				save(path, cf)
			}
			continue
		}

		switch {
		case i == 0:
			if c, res := cheatForm(T(l10n.TitleNewCheat), cheats.Cheat{}, false); res == formSave {
				cf.Cheats = append(cf.Cheats, c)
				if save(path, cf) {
					p.sel = len(cf.Cheats) + 1 // select the new cheat
				}
			}
		case i == 1:
			importChoice(path, tag, cf)
		case isCheat:
			c, res := cheatForm(Tf(l10n.TitleCheat, "Number", ci+1), cf.Cheats[ci], true)
			switch res {
			case formSave:
				cf.Cheats[ci] = c
				save(path, cf)
			case formDelete:
				if confirm(Tf(l10n.MsgConfirmDelete, "Name", cf.Cheats[ci].Desc), T(l10n.BtnDelete)) {
					cf.Cheats = append(cf.Cheats[:ci], cf.Cheats[ci+1:]...)
					save(path, cf)
				}
			}
		case hasFile:
			if confirm(Tf(l10n.MsgConfirmDeleteAll, "Game", rom), T(l10n.BtnDelete)) {
				if err := os.Rename(path, path+".bak"); err != nil {
					message(Tf(l10n.MsgError, "Error", err))
				}
				p = pos{}
			}
		}
	}
}

type formResult int

const (
	formCancel formResult = iota
	formSave
	formDelete
)

func textOption(v string) gaba.Option {
	display := v
	if display == "" {
		display = T(l10n.ItemEmpty)
	}
	return gaba.Option{DisplayName: display, Value: v, Type: gaba.OptionTypeKeyboard, KeyboardPrompt: v}
}

// cheatForm shows the Name / Code / Enabled form (plus Delete when editing).
// START saves, B cancels.
func cheatForm(title string, c cheats.Cheat, canDelete bool) (cheats.Cheat, formResult) {
	enabled := 0
	if c.Enable {
		enabled = 1
	}
	items := []gaba.ItemWithOptions{
		{Item: gaba.MenuItem{Text: T(l10n.FieldName)}, Options: []gaba.Option{textOption(c.Desc)}},
		{Item: gaba.MenuItem{Text: T(l10n.FieldCode)}, Options: []gaba.Option{textOption(c.Code)}},
		{Item: gaba.MenuItem{Text: T(l10n.FieldEnabled)}, SelectedOption: enabled, Options: []gaba.Option{
			{DisplayName: T(l10n.OptionNo), Value: false},
			{DisplayName: T(l10n.OptionYes), Value: true},
		}},
	}
	const deleteIndex = 3
	if canDelete {
		items = append(items, gaba.ItemWithOptions{
			Item:    gaba.MenuItem{Text: T(l10n.ItemDeleteCheat)},
			Options: []gaba.Option{{DisplayName: "", Type: gaba.OptionTypeClickable}},
		})
	}

	for {
		r, err := gaba.OptionsList(title, gaba.OptionListSettings{
			FooterHelpItems: footer("B", T(l10n.BtnCancel), "A", T(l10n.BtnEdit), "START", T(l10n.BtnSave)),
		}, items)
		if err != nil || r == nil {
			return c, formCancel
		}
		items = r.Items
		if r.Action == gaba.ListActionSelected && canDelete && r.Selected == deleteIndex {
			return c, formDelete
		}
		if r.Action != gaba.ListActionConfirmed {
			continue
		}
		out := c
		out.Desc = strings.TrimSpace(optValue(items[0]))
		out.Code = strings.TrimSpace(optValue(items[1]))
		out.Enable = items[2].SelectedOption == 1
		if out.Desc == "" {
			message(T(l10n.MsgNameEmpty))
			continue
		}
		if out.Code == "" {
			message(T(l10n.MsgCodeEmpty))
			continue
		}
		return out, formSave
	}
}

func optValue(it gaba.ItemWithOptions) string {
	if it.SelectedOption < 0 || it.SelectedOption >= len(it.Options) {
		return ""
	}
	if s, ok := it.Options[it.SelectedOption].Value.(string); ok {
		return s
	}
	return ""
}

func importChoice(target, tag string, cf *cheats.File) {
	var p pos
	choices := []string{T(l10n.ItemImportFile), T(l10n.ItemImportURL)}
	i, _, err := showList(T(l10n.TitleImportFrom), choices, &p, "", noExtra,
		"B", T(l10n.BtnCancel), "A", T(l10n.BtnSelect))
	if err != nil {
		return
	}
	switch i {
	case 0:
		importCht(target, tag, cf)
	case 1:
		importFromURL(target, cf)
	}
}

func importFromURL(target string, cf *cheats.File) {
	items := []gaba.ItemWithOptions{
		{Item: gaba.MenuItem{Text: "URL"}, Options: []gaba.Option{textOption("")}},
	}
	for {
		r, err := gaba.OptionsList(T(l10n.ItemImportURL), gaba.OptionListSettings{
			FooterHelpItems: footer("B", T(l10n.BtnCancel), "A", T(l10n.BtnEdit), "START", T(l10n.BtnConfirm)),
		}, items)
		if err != nil || r == nil {
			return
		}
		items = r.Items
		if r.Action != gaba.ListActionConfirmed {
			continue
		}
		url := strings.TrimSpace(optValue(items[0]))
		if url == "" {
			message(T(l10n.MsgURLEmpty))
			continue
		}

		data, err := downloadCht(url)
		if err != nil {
			slog.Error("download failed", "url", url, "error", err)
			message(Tf(l10n.MsgDownloadError, "Error", err))
			return
		}
		src := cheats.Parse(data)
		if len(src.Cheats) == 0 {
			message(T(l10n.MsgInvalidCht))
			return
		}
		mergeImport(target, cf, src)
		return
	}
}

const maxChtSize = 4 * 1024 * 1024 // 4 MiB

func downloadCht(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxChtSize))
}

func importCht(target, tag string, cf *cheats.File) {
	files := cheats.ListImportable(cheatsPath, tag, target)
	if len(files) == 0 {
		message(Tf(l10n.MsgNoChtFound, "Path", cheatsPath))
		return
	}
	var p pos
	i, _, err := showList(T(l10n.ItemImportFile), files, &p, "", noExtra, "B", T(l10n.BtnCancel), "A", T(l10n.BtnImport))
	if err != nil {
		return
	}
	src, err := cheats.Load(filepath.Join(cheatsPath, files[i]))
	if err != nil || len(src.Cheats) == 0 {
		message(T(l10n.MsgInvalidCht))
		return
	}
	mergeImport(target, cf, src)
}

func mergeImport(target string, cf, src *cheats.File) {
	replace := false
	if len(cf.Cheats) > 0 {
		var mp pos
		choices := []string{T(l10n.ItemImportAdd), T(l10n.ItemImportReplace)}
		m, _, err := showList(Tn(l10n.TitleImportCount, len(src.Cheats)), choices, &mp, "", noExtra,
			"B", T(l10n.BtnCancel), "A", T(l10n.BtnConfirm))
		if err != nil {
			return
		}
		replace = m == 1
	}
	if replace {
		cf.Cheats = nil
	}
	cf.Cheats = append(cf.Cheats, src.Cheats...)
	if save(target, cf) {
		message(Tn(l10n.MsgImported, len(src.Cheats)))
	}
}
