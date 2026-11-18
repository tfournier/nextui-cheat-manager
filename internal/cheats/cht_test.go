package cheats

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sample = "# comment\r\ncheats = 3\r\n\r\ncheat0_desc = \"Infinite health\"\r\ncheat0_code = \"82000F20+0063\"\r\ncheat0_enable = false\r\ncheat0_handler = \"1\"\r\n\r\ncheat2_desc = \"Max money\"\r\ncheat2_code = \"XXXX YYYY\"\r\ncheat2_enable = true\r\n\r\ncheat1_desc = \"Debug\"\r\ncheat1_code = \"ABC-DEF\"\r\ncheat1_enable = false\r\n"

func TestParse(t *testing.T) {
	cf := Parse([]byte(sample))
	if len(cf.Cheats) != 3 {
		t.Fatalf("want 3 cheats, got %d", len(cf.Cheats))
	}
	if cf.Cheats[1].Desc != "Debug" || cf.Cheats[2].Desc != "Max money" || !cf.Cheats[2].Enable {
		t.Fatalf("wrong order or values: %+v", cf.Cheats)
	}
	if len(cf.Cheats[0].Extras) != 1 || cf.Cheats[0].Extras[0].Key != "handler" {
		t.Fatalf("handler key lost: %+v", cf.Cheats[0])
	}
	if len(cf.Header) != 1 || cf.Header[0] != "# comment" {
		t.Fatalf("header lost: %q", cf.Header)
	}
}

func TestRoundTripAndEdit(t *testing.T) {
	cf := Parse([]byte(sample))
	cf.Cheats = append(cf.Cheats[:1], cf.Cheats[2:]...) // remove "Debug"
	cf.Cheats = append(cf.Cheats, Cheat{Desc: `God "mode"`, Code: "1234+5678"})
	out := string(cf.Bytes())
	for _, want := range []string{
		"cheats = 3\n",
		"cheat0_handler = \"1\"",
		"cheat1_desc = \"Max money\"",
		"cheat1_enable = true",
		"cheat2_desc = \"God 'mode'\"",
		"cheat2_enable = false",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output is missing %q:\n%s", want, out)
		}
	}
	again := Parse([]byte(out))
	if string(again.Bytes()) != out {
		t.Errorf("re-encoding is not stable")
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "MGBA", "Game.gba.cht")
	cf, _ := Load(p)
	if len(cf.Cheats) != 0 {
		t.Fatal("a missing file should give 0 cheats")
	}
	cf.Cheats = append(cf.Cheats, Cheat{Desc: "A", Code: "1"})
	if err := Save(p, cf); err != nil {
		t.Fatal(err)
	}
	cf.Cheats = append(cf.Cheats, Cheat{Desc: "B", Code: "2"})
	if err := Save(p, cf); err != nil {
		t.Fatal(err)
	}
	if Count(p) != 2 || Count(p+".bak") != 1 {
		t.Fatal("wrong save or .bak content")
	}
	if _, err := os.Stat(p + ".tmp"); !os.IsNotExist(err) {
		t.Fatal(".tmp file not cleaned up")
	}
}
