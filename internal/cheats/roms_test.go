package cheats

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func touch(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListSystemsAndEntries(t *testing.T) {
	sd := t.TempDir()
	roms := filepath.Join(sd, "Roms")
	cheatsRoot := filepath.Join(sd, "Cheats")
	gba := filepath.Join(roms, "Game Boy Advance (MGBA)")

	touch(t, filepath.Join(gba, "Metroid.gba"))
	touch(t, filepath.Join(gba, "readme.txt"))
	touch(t, filepath.Join(gba, ".hidden.gba"))
	touch(t, filepath.Join(gba, "Hacks", "Hack.gba"))
	touch(t, filepath.Join(roms, "Super Nintendo (SFC)", "Mario.sfc"))
	if err := os.MkdirAll(filepath.Join(roms, "Empty (GB)"), 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, filepath.Join(roms, "No tag", "x.zip"))

	systems := ListSystems(roms)
	var labels []string
	for _, s := range systems {
		labels = append(labels, s.Label+"|"+s.Tag)
	}
	want := []string{"Game Boy Advance (MGBA)|MGBA", "Super Nintendo (SFC)|SFC"}
	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("systems = %v, want %v", labels, want)
	}

	cf := &File{Cheats: []Cheat{{Desc: "A", Code: "1"}, {Desc: "B", Code: "2"}}}
	if err := Save(CheatPath(cheatsRoot, "MGBA", "Metroid.gba"), cf); err != nil {
		t.Fatal(err)
	}

	entries := ListEntries(gba, filepath.Join(cheatsRoot, "MGBA"))
	wantEntries := []Entry{{Name: "Hacks", IsDir: true}, {Name: "Metroid.gba", Cheats: 2}}
	if !reflect.DeepEqual(entries, wantEntries) {
		t.Fatalf("entries = %+v, want %+v", entries, wantEntries)
	}

	touch(t, filepath.Join(cheatsRoot, "_import", "Other.cht"))
	target := CheatPath(cheatsRoot, "MGBA", "Hack.gba")
	got := ListImportable(cheatsRoot, "MGBA", target)
	wantImport := []string{filepath.Join("MGBA", "Metroid.gba.cht"), filepath.Join("_import", "Other.cht")}
	if !reflect.DeepEqual(got, wantImport) {
		t.Fatalf("importable = %v, want %v", got, wantImport)
	}
}
