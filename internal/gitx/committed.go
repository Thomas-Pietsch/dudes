package gitx

import (
	"os"
	"path/filepath"
	"strings"
)

// ReadCommitted returns a file as its repository's HEAD holds it — what a run may rely on — not
// the working copy someone is still editing. A file outside any git repository is read from disk;
// a file its repository has not committed does not exist (os.ErrNotExist).
func ReadCommitted(abs string) ([]byte, error) {
	dir := filepath.Dir(abs)
	top, _, err := run(dir, "rev-parse", "--show-toplevel")
	if err != nil || strings.TrimSpace(top) == "" {
		return os.ReadFile(abs)
	}
	rel, err := filepath.Rel(filepath.Clean(strings.TrimSpace(top)), filepath.Clean(abs))
	if err != nil {
		return os.ReadFile(abs)
	}
	out, _, err := run(dir, "show", "HEAD:"+filepath.ToSlash(rel))
	if err != nil {
		return nil, &os.PathError{Op: "read committed", Path: abs, Err: os.ErrNotExist}
	}
	return []byte(out), nil
}
