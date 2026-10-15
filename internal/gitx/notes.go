package gitx

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Notes as a record that grows (F88): appended to, never replaced, and read back whole, whichever
// worktree or branch wrote them. A notes ref is one for every worktree of a repository, so what is
// recorded there outlives the worktree that recorded it.

// AppendNote adds text to the note on HEAD under refs/notes/<ref>, after what the note holds (git puts
// a blank line between), or as the note when there is none.
func AppendNote(dir, ref, text string) error {
	f, err := os.CreateTemp("", "dudes-note-*")
	if err != nil {
		return fmt.Errorf("git notes: %w", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(text); err != nil {
		f.Close()
		return fmt.Errorf("git notes: %w", err)
	}
	f.Close()
	if _, _, err := run(dir, "notes", "--ref", ref, "append", "-F", f.Name(), "HEAD"); err != nil {
		return fmt.Errorf("git notes append: %w", err)
	}
	return nil
}

// AllNotes returns every note under refs/notes/<ref> in the repository at dir, not only those in
// HEAD's history: a note stays readable when its commit is on another branch, or gone with a deleted
// one. None: no note, or no such ref.
func AllNotes(dir, ref string) ([]Note, error) {
	out, _, err := run(dir, "notes", "--ref", ref, "list")
	if err != nil {
		return nil, fmt.Errorf("git notes list: %w", err)
	}
	var blobs, commits []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if blob, commit, ok := strings.Cut(strings.TrimSpace(line), " "); ok {
			blobs, commits = append(blobs, blob), append(commits, commit)
		}
	}
	if len(blobs) == 0 {
		return nil, nil
	}
	cmd := exec.Command("git", "cat-file", "--batch")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(strings.Join(blobs, "\n") + "\n")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git cat-file: %v %s", err, strings.TrimSpace(stderr.String()))
	}
	// --batch answers each name with "<sha> blob <size>", the content, and a newline
	r := bufio.NewReader(bytes.NewReader(data))
	notes := make([]Note, 0, len(blobs))
	for _, commit := range commits {
		header, err := r.ReadString('\n')
		if err != nil {
			return notes, fmt.Errorf("git cat-file: %w", err)
		}
		fields := strings.Fields(header)
		if len(fields) != 3 {
			continue // "<sha> missing": nothing to read
		}
		size, err := strconv.Atoi(fields[2])
		if err != nil {
			return notes, fmt.Errorf("git cat-file: %q", header)
		}
		body := make([]byte, size+1)
		if _, err := io.ReadFull(r, body); err != nil {
			return notes, fmt.Errorf("git cat-file: %w", err)
		}
		notes = append(notes, Note{Commit: commit, Text: string(body[:size])})
	}
	return notes, nil
}
