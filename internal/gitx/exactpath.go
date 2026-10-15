package gitx

import (
	"path/filepath"
	"runtime"
	"strings"
)

// exactPath addresses rel inside dir by its exact name. A worker in a Linux container can create
// names Win32 path handling rewrites — a trailing dot or space is stripped, by filepath.Abs too
// (GetFullPathName) — so on Windows only the directory is resolved and the name goes to NTFS
// unparsed, through the \\?\ prefix. A character Win32 forbids in a name (* ? " < > | :) cannot
// be the name on disk either: WSL stores it as U+F000 plus its code (DrvFs), so that is the name
// asked for — in a replay of real product slices a reply sentence with *must* in it, as a file
// name, stopped the step.
func exactPath(dir, rel string) string {
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if runtime.GOOS != "windows" {
		return p
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return p
	}
	return `\\?\` + abs + `\` + filepath.FromSlash(wslName(rel))
}

// wslName maps the characters Win32 forbids in a name to the private-use code points WSL stores
// them as (U+F000 plus the character's code); / stays the separator.
func wslName(rel string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(`*?"<>|:`, r) || (r > 0 && r < 0x20) {
			return 0xF000 + r
		}
		return r
	}, rel)
}
