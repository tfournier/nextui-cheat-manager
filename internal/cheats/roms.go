package cheats

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// System is a ROM folder of the SD card.
type System struct {
	Label string // display name
	Dir   string // full path of the ROM folder
	Tag   string // tag between parentheses, also the folder name under /Cheats
}

// Entry is an item of a ROM folder.
type Entry struct {
	Name   string // file or folder name
	IsDir  bool
	Cheats int // number of existing cheats (files only)
}

var (
	tagRe      = regexp.MustCompile(`\(([^()]*)\)\s*$`)
	leadingNum = regexp.MustCompile(`^\d+[)._ -]+`)
)

var ignoredExt = map[string]bool{
	".txt": true, ".png": true, ".jpg": true, ".jpeg": true, ".bmp": true, ".xml": true,
	".json": true, ".log": true, ".db": true, ".srm": true, ".sav": true, ".state": true,
	".cht": true, ".bak": true, ".tmp": true, ".pdf": true, ".md": true,
}

func hidden(name string) bool { return strings.HasPrefix(name, ".") }

func lessFold(a, b string) bool { return strings.ToLower(a) < strings.ToLower(b) }

// CheatPath returns the .cht path NextUI uses for a ROM:
// <cheatsRoot>/<TAG>/<rom file name>.cht
func CheatPath(cheatsRoot, tag, romName string) string {
	return filepath.Join(cheatsRoot, tag, romName+".cht")
}

// ListSystems returns the folders of romsPath that have a "(TAG)" suffix and
// at least one visible item.
func ListSystems(romsPath string) []System {
	des, err := os.ReadDir(romsPath)
	if err != nil {
		return nil
	}
	var out []System
	for _, de := range des {
		if !de.IsDir() || hidden(de.Name()) {
			continue
		}
		m := tagRe.FindStringSubmatch(de.Name())
		if m == nil || m[1] == "" {
			continue
		}
		dir := filepath.Join(romsPath, de.Name())
		if !hasVisible(dir) {
			continue
		}
		label := leadingNum.ReplaceAllString(de.Name(), "")
		if label == "" {
			label = de.Name()
		}
		out = append(out, System{Label: label, Dir: dir, Tag: m[1]})
	}
	sort.Slice(out, func(i, j int) bool { return lessFold(out[i].Label, out[j].Label) })
	return out
}

func hasVisible(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return false
	}
	defer f.Close()
	for {
		names, err := f.Readdirnames(32)
		for _, n := range names {
			if !hidden(n) {
				return true
			}
		}
		if err != nil {
			return false
		}
	}
}

// ListEntries lists the sub-folders then the ROMs of a folder, with the number
// of cheats found in cheatDir for each ROM.
func ListEntries(dir, cheatDir string) []Entry {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var dirs, files []Entry
	for _, de := range des {
		n := de.Name()
		if hidden(n) {
			continue
		}
		isDir := de.IsDir()
		if de.Type()&fs.ModeSymlink != 0 {
			if st, err := os.Stat(filepath.Join(dir, n)); err == nil {
				isDir = st.IsDir()
			}
		}
		if isDir {
			dirs = append(dirs, Entry{Name: n, IsDir: true})
			continue
		}
		if ignoredExt[strings.ToLower(filepath.Ext(n))] {
			continue
		}
		files = append(files, Entry{Name: n, Cheats: Count(filepath.Join(cheatDir, n+".cht"))})
	}
	sort.Slice(dirs, func(i, j int) bool { return lessFold(dirs[i].Name, dirs[j].Name) })
	sort.Slice(files, func(i, j int) bool { return lessFold(files[i].Name, files[j].Name) })
	return append(dirs, files...)
}

// ListImportable returns the .cht files found under cheatsRoot (relative
// paths), those of the given tag first, excluding the target file.
func ListImportable(cheatsRoot, tag, exclude string) []string {
	var same, other []string
	_ = filepath.WalkDir(cheatsRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if p != cheatsRoot && hidden(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(p), ".cht") || p == exclude {
			return nil
		}
		rel, err := filepath.Rel(cheatsRoot, p)
		if err != nil {
			return nil
		}
		if strings.HasPrefix(rel, tag+string(filepath.Separator)) {
			same = append(same, rel)
		} else {
			other = append(other, rel)
		}
		return nil
	})
	sort.Slice(same, func(i, j int) bool { return lessFold(same[i], same[j]) })
	sort.Slice(other, func(i, j int) bool { return lessFold(other[i], other[j]) })
	return append(same, other...)
}
