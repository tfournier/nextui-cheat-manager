package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSave(t *testing.T) {
	p := Path(t.TempDir())

	s, err := Load(p)
	if err != nil || s.Language != "" {
		t.Fatalf("first launch: got %+v, %v", s, err)
	}

	if err := Save(p, Settings{Language: "fr"}); err != nil {
		t.Fatal(err)
	}
	s, err = Load(p)
	if err != nil || s.Language != "fr" {
		t.Fatalf("after save: got %+v, %v", s, err)
	}
	if _, err := os.Stat(p + ".tmp"); !os.IsNotExist(err) {
		t.Fatal(".tmp file not cleaned up")
	}
}

func TestLoadCorrupted(t *testing.T) {
	p := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(p, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Fatal("expected an error for a corrupted file")
	}
}
