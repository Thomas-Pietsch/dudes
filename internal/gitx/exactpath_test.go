package gitx

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

// A worker in the Linux container named a file after a sentence of its reply, ending in a dot
// ("Let's look at `main.go` again."). Discard must remove exactly that file, also on Windows,
// where any normalised path (filepath.Abs included) loses the dot and the file stayed behind.
func TestDiscardTrailingDot(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	name := "Let's look at `main.go` again."
	p := dir + "/" + name
	if runtime.GOOS == "windows" {
		p = `\\?\` + dir + `\` + name // built by hand: the file must really end in a dot
	}
	if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Discard(dir, []string{name}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(p); !os.IsNotExist(err) {
		t.Fatalf("%q survived Discard: %v", name, err)
	}
}

// A worker named a file after a heading ("Shelf Schema v1 (from TA §4.2)"): ChangedPaths must
// return the name as it is — git's text status quotes and octal-escapes it — so Discard finds it,
// staged or not.
func TestChangedPathsRawName(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "t@t"}, {"config", "user.name", "t"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	name := "Shelf Schema v1 (from TA §4.2)"
	if err := os.WriteFile(dir+"/"+name, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "add", "-A").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	paths, err := ChangedPaths(dir)
	if err != nil || len(paths) != 1 || paths[0] != name {
		t.Fatalf("ChangedPaths = %q, %v; want [%q]", paths, err, name)
	}
	if err := Discard(dir, paths); err != nil {
		t.Fatal(err)
	}
	if paths, _ := ChangedPaths(dir); len(paths) != 0 {
		t.Fatalf("%q survived Discard: %q", name, paths)
	}
}

// The worker named a file after a sentence with *must* in it. In the container the name has a
// real '*'; WSL stores it on NTFS as U+F02A, and git (from the index) still names it with '*'.
// Discard must remove the stored file instead of failing on a name Windows refuses.
func TestDiscardWSLMappedName(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	name := "Wait, the `Database` struct *must* have `URL`"
	p := dir + "/" + name
	if runtime.GOOS == "windows" {
		p = `\\?\` + dir + `\` + wslName(name) // as WSL stored it
	}
	if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Discard(dir, []string{name}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(p); !os.IsNotExist(err) {
		t.Fatalf("%q survived Discard: %v", name, err)
	}
}
