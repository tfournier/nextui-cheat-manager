package l10n

import "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/i18n"

// All UI messages, with their English text. Every ID must exist in each
// locales/active.<lang>.toml file (checked by the tests).

type msg = i18n.Message

// Footer buttons
var (
	BtnQuit     = &msg{ID: "button_quit", Other: "Quit"}
	BtnSelect   = &msg{ID: "button_select", Other: "Select"}
	BtnBack     = &msg{ID: "button_back", Other: "Back"}
	BtnToggle   = &msg{ID: "button_toggle", Other: "ON/OFF"}
	BtnCancel   = &msg{ID: "button_cancel", Other: "Cancel"}
	BtnEdit     = &msg{ID: "button_edit", Other: "Edit"}
	BtnSave     = &msg{ID: "button_save", Other: "Save"}
	BtnOK       = &msg{ID: "button_ok", Other: "OK"}
	BtnDelete   = &msg{ID: "button_delete", Other: "Delete"}
	BtnImport   = &msg{ID: "button_import", Other: "Import"}
	BtnConfirm  = &msg{ID: "button_confirm", Other: "Confirm"}
	BtnLanguage = &msg{ID: "button_language", Other: "Language"}
)

// Screen titles
var (
	TitleNewCheat    = &msg{ID: "title_new_cheat", Other: "New cheat"}
	TitleCheat       = &msg{ID: "title_cheat", Other: "Cheat {{.Number}}"}
	TitleImportFrom  = &msg{ID: "title_import_from", Other: "Import from..."}
	TitleImportCount = &msg{ID: "title_import_count", One: "{{.Count}} cheat to import", Other: "{{.Count}} cheats to import"}
	TitleLanguage    = &msg{ID: "title_language", Other: "Language"}
	LanguagePrompt   = &msg{ID: "language_prompt", Other: "Choose the interface language."}
)

// List items and form fields
var (
	ItemAddCheat      = &msg{ID: "item_add_cheat", Other: "+ Add a cheat"}
	ItemImport        = &msg{ID: "item_import", Other: "+ Import a .cht file"}
	ItemDeleteAll     = &msg{ID: "item_delete_all", Other: "- Delete all cheats"}
	ItemNoName        = &msg{ID: "item_no_name", Other: "(no name)"}
	ItemEmpty         = &msg{ID: "item_empty", Other: "(empty)"}
	FieldName         = &msg{ID: "field_name", Other: "Name"}
	FieldCode         = &msg{ID: "field_code", Other: "Code"}
	FieldEnabled      = &msg{ID: "field_enabled", Other: "Enabled"}
	OptionNo          = &msg{ID: "option_no", Other: "No"}
	OptionYes         = &msg{ID: "option_yes", Other: "Yes"}
	ItemDeleteCheat   = &msg{ID: "item_delete_cheat", Other: "Delete this cheat"}
	ItemImportFile    = &msg{ID: "item_import_file", Other: "From a local file"}
	ItemImportURL     = &msg{ID: "item_import_url", Other: "From a URL"}
	ItemImportAdd     = &msg{ID: "item_import_add", Other: "Add to existing cheats"}
	ItemImportReplace = &msg{ID: "item_import_replace", Other: "Replace existing cheats"}
)

// Messages
var (
	MsgNoSystem         = &msg{ID: "msg_no_system", Other: "No system found in\n{{.Path}}"}
	MsgNoGame           = &msg{ID: "msg_no_game", Other: "No game in this folder"}
	MsgReadError        = &msg{ID: "msg_read_error", Other: "Cannot read file:\n{{.Error}}"}
	MsgWriteError       = &msg{ID: "msg_write_error", Other: "Write error:\n{{.Error}}"}
	MsgError            = &msg{ID: "msg_error", Other: "Error:\n{{.Error}}"}
	MsgConfirmDelete    = &msg{ID: "msg_confirm_delete", Other: "Delete the cheat\n\"{{.Name}}\"?"}
	MsgConfirmDeleteAll = &msg{ID: "msg_confirm_delete_all", Other: "Delete ALL cheats of\n{{.Game}}?\n\n(a .bak copy is kept)"}
	MsgNameEmpty        = &msg{ID: "msg_name_empty", Other: "The name is empty."}
	MsgCodeEmpty        = &msg{ID: "msg_code_empty", Other: "The code is empty.\n\nMultiple codes: separate them with +"}
	MsgNoChtFound       = &msg{ID: "msg_no_cht_found", Other: "No .cht file found in\n{{.Path}}\n\nCopy some there first (e.g. Cheats/_import/)"}
	MsgInvalidCht       = &msg{ID: "msg_invalid_cht", Other: "This file contains no valid cheat."}
	MsgImported         = &msg{ID: "msg_imported", One: "{{.Count}} cheat imported", Other: "{{.Count}} cheats imported"}
	MsgURLEmpty         = &msg{ID: "msg_url_empty", Other: "The URL is empty."}
	MsgDownloadError    = &msg{ID: "msg_download_error", Other: "Download failed:\n{{.Error}}"}
)

// All lists every message, for the tests.
var All = []*msg{
	BtnQuit, BtnSelect, BtnBack, BtnToggle, BtnCancel, BtnEdit, BtnSave, BtnOK,
	BtnDelete, BtnImport, BtnConfirm, BtnLanguage,
	TitleNewCheat, TitleCheat, TitleImportFrom, TitleImportCount, TitleLanguage, LanguagePrompt,
	ItemAddCheat, ItemImport, ItemDeleteAll, ItemNoName, ItemEmpty, FieldName, FieldCode,
	FieldEnabled, OptionNo, OptionYes, ItemDeleteCheat, ItemImportFile, ItemImportURL,
	ItemImportAdd, ItemImportReplace,
	MsgNoSystem, MsgNoGame, MsgReadError, MsgWriteError, MsgError, MsgConfirmDelete,
	MsgConfirmDeleteAll, MsgNameEmpty, MsgCodeEmpty, MsgNoChtFound, MsgInvalidCht, MsgImported,
	MsgURLEmpty, MsgDownloadError,
}
