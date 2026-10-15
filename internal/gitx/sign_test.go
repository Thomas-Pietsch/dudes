package gitx_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
)

// isolatedGit keeps the developer's own git configuration out of a test: a signing key, commit.gpgsign
// or an identity of theirs would change what these tests see.
func isolatedGit(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "") // registers the restore; the variable itself goes for the test
		os.Unsetenv(k)
	}
}

// sealRepo is a repository with one commit and the identity Ada Sealer <ada@example.com>.
func sealRepo(t *testing.T) string {
	t.Helper()
	isolatedGit(t)
	dir := t.TempDir()
	gitOut(t, dir, "init", "-q", "-b", "main")
	gitOut(t, dir, "config", "user.email", "ada@example.com")
	gitOut(t, dir, "config", "user.name", "Ada Sealer")
	commitFile(t, dir, "a.md", "first")
	return dir
}

func commitFile(t *testing.T, dir, name, message string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gitx.AddCommit(dir, []string{name}, message); err != nil {
		t.Fatal(err)
	}
}

// tagOn makes an annotated tag with a tag date of its own: tags made in the same second would have
// no order.
func tagOn(t *testing.T, dir, name, date string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "tag", "-a", name, "-m", "m")
	cmd.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+date)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git tag %s: %v\n%s", name, err, out)
	}
}

// gitAtLeast reports whether the git on this machine is major.minor or newer.
func gitAtLeast(t *testing.T, major, minor int) bool {
	t.Helper()
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		t.Fatal(err)
	}
	var maj, min int
	if _, err := fmt.Sscanf(strings.TrimPrefix(strings.TrimSpace(string(out)), "git version "), "%d.%d", &maj, &min); err != nil {
		t.Fatalf("git --version: %q: %v", out, err)
	}
	return maj > major || (maj == major && min >= minor)
}

// sshKey makes an SSH key pair without a passphrase in a folder of its own. It returns the private
// key's path and the public line. The test is skipped where this machine cannot sign with SSH keys.
func sshKey(t *testing.T, comment string) (key, pub string) {
	t.Helper()
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("no ssh-keygen on PATH")
	}
	if !gitAtLeast(t, 2, 34) {
		t.Skip("this git cannot sign with SSH keys (2.34)")
	}
	key = filepath.ToSlash(filepath.Join(t.TempDir(), "key"))
	if out, err := exec.Command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-C", comment, "-f", key).CombinedOutput(); err != nil {
		t.Fatalf("ssh-keygen: %v\n%s", err, out)
	}
	line, err := os.ReadFile(key + ".pub")
	if err != nil {
		t.Fatal(err)
	}
	return key, strings.TrimSpace(string(line))
}

func TestIdentity(t *testing.T) {
	dir := sealRepo(t)
	if got, err := gitx.Identity(dir); err != nil || got != "Ada Sealer <ada@example.com>" {
		t.Fatalf("identity %q, %v", got, err)
	}
	// git that is told not to guess has none to give: the error says what to set
	bare := t.TempDir()
	gitOut(t, bare, "init", "-q")
	gitOut(t, bare, "config", "user.useConfigOnly", "true")
	if got, err := gitx.Identity(bare); err == nil || !strings.Contains(err.Error(), "user.name and user.email") {
		t.Fatalf("no identity: got %q, %v", got, err)
	}
}

// The sealer's git can sign when user.signingkey is set and gpg.format is one git knows.
func TestCanSign(t *testing.T) {
	const unsigned = "set user.signingkey to sign seals"
	isolatedGit(t)
	for _, tc := range []struct {
		name   string
		config [][2]string
		ok     bool
		format string
		why    string
	}{
		{"nothing set", nil, false, "openpgp", unsigned},
		{"a key, git's default format", [][2]string{{"user.signingkey", "ABCD1234"}}, true, "openpgp", ""},
		{"a key, ssh", [][2]string{{"user.signingkey", "key"}, {"gpg.format", "ssh"}}, true, "ssh", ""},
		{"a key, x509", [][2]string{{"user.signingkey", "key"}, {"gpg.format", "x509"}}, true, "x509", ""},
		{"a key, a format git does not know", [][2]string{{"user.signingkey", "key"}, {"gpg.format", "pgp"}}, false, "pgp", `gpg.format "pgp" is none of openpgp, ssh, x509`},
		{"a format but no key", [][2]string{{"gpg.format", "ssh"}}, false, "ssh", unsigned},
		{"a key set to nothing", [][2]string{{"user.signingkey", ""}}, false, "openpgp", unsigned},
		// git that signs every tag does it with the default key of the identity: the advice before the
		// engine signed seals, so such a seal is not reported unsigned
		{"every tag signed, no key", [][2]string{{"tag.gpgsign", "true"}}, true, "openpgp", ""},
		{"every tag signed, ssh", [][2]string{{"tag.gpgsign", "yes"}, {"gpg.format", "ssh"}}, true, "ssh", ""},
		{"tag.gpgSign off", [][2]string{{"tag.gpgsign", "false"}}, false, "openpgp", unsigned},
		{"tag.gpgSign no boolean", [][2]string{{"tag.gpgsign", "maybe"}}, false, "openpgp", unsigned},
	} {
		dir := t.TempDir() // CanSign reads git's configuration: no commit, no identity needed
		gitOut(t, dir, "init", "-q")
		for _, kv := range tc.config {
			gitOut(t, dir, "config", kv[0], kv[1])
		}
		got := gitx.CanSign(dir)
		if got.OK != tc.ok || got.Format != tc.format || got.Why != tc.why {
			t.Errorf("%s: got %+v, want ok=%v format=%s why=%q", tc.name, got, tc.ok, tc.format, tc.why)
		}
	}
}

// A key that cannot sign makes no tag.
func TestSignedTagWithoutAWorkingKey(t *testing.T) {
	dir := sealRepo(t)
	gitOut(t, dir, "config", "gpg.format", "ssh")
	gitOut(t, dir, "config", "user.signingkey", filepath.ToSlash(filepath.Join(t.TempDir(), "no-such-key")))
	if err := gitx.SignedTag(dir, "dudes/seal/demo/run-1", "m"); err == nil {
		t.Fatal("a tag was signed with a key that is not there")
	}
	if gitx.TagExists(dir, "dudes/seal/demo/run-1") {
		t.Fatal("a failed signature left a tag behind")
	}
}

// SignedTag signs with the sealer's key and git's own verifier accepts it. It does not accept the tag
// once its text is changed with the signature kept, nor a signature by a key that is not an allowed
// signer: who counts as trusted is the verifying git's setting.
func TestSignedTagVerifies(t *testing.T) {
	dir := sealRepo(t)
	key, pub := sshKey(t, "ada@example.com")
	allowed := filepath.ToSlash(filepath.Join(t.TempDir(), "allowed_signers"))
	trusted := "ada@example.com " + pub + "\n"
	if err := os.WriteFile(allowed, []byte(trusted), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "config", "gpg.format", "ssh")
	gitOut(t, dir, "config", "user.signingkey", key)
	gitOut(t, dir, "config", "gpg.ssh.allowedSignersFile", allowed)

	const tag = "dudes/seal/demo/run-1"
	if err := gitx.SignedTag(dir, tag, "DUDES human seal for demo (run-1)"); err != nil {
		t.Fatal(err)
	}
	object := gitOut(t, dir, "rev-parse", tag)
	raw, err := exec.Command("git", "-C", dir, "cat-file", "tag", tag).Output()
	if err != nil || !strings.Contains(string(raw), "BEGIN SSH SIGNATURE") {
		t.Fatalf("the tag carries no SSH signature: %v\n%s", err, raw)
	}
	verify := func() (said string, ok bool) {
		t.Helper()
		said, ok, err := gitx.VerifyTag(dir, tag)
		if err != nil {
			t.Fatal(err)
		}
		return said, ok
	}
	if said, ok := verify(); !ok || !strings.Contains(said, "Good") {
		t.Fatalf("a signed tag: ok=%v said=%q", ok, said)
	}

	// the text changed after it was signed, the signature kept
	forged := strings.Replace(string(raw), "for demo", "for evil", 1)
	if forged == string(raw) {
		t.Fatal("nothing to tamper with")
	}
	file := filepath.Join(t.TempDir(), "tag")
	if err := os.WriteFile(file, []byte(forged), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "update-ref", "refs/tags/"+tag, gitOut(t, dir, "hash-object", "-t", "tag", "-w", file))
	if said, ok := verify(); ok {
		t.Fatalf("a tag changed after it was signed was accepted: %q", said)
	}
	gitOut(t, dir, "update-ref", "refs/tags/"+tag, object)
	if _, ok := verify(); !ok {
		t.Fatal("the genuine tag, put back, should verify again")
	}

	// a key that is not an allowed signer
	_, eve := sshKey(t, "eve@example.com")
	if err := os.WriteFile(allowed, []byte("eve@example.com "+eve+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if said, ok := verify(); ok {
		t.Fatalf("a signature by a key that is not an allowed signer was accepted: %q", said)
	}
}

// An unsigned tag, a tag that is no tag object and a name that is no name: git says no, in words.
func TestVerifyTagRefuses(t *testing.T) {
	dir := sealRepo(t)
	if err := gitx.AnnotatedTag(dir, "dudes/seal/demo/run-1", "m"); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "tag", "light")
	said, ok, err := gitx.VerifyTag(dir, "dudes/seal/demo/run-1")
	if err != nil || ok || !strings.Contains(said, "no signature found") {
		t.Fatalf("unsigned: ok=%v err=%v said=%q", ok, err, said)
	}
	for _, name := range []string{"light", "no-such-tag"} {
		if said, ok, err := gitx.VerifyTag(dir, name); err != nil || ok || said == "" {
			t.Errorf("%s: ok=%v err=%v said=%q, want a refusal with git's words", name, ok, err, said)
		}
	}
	if _, _, err := gitx.VerifyTag(dir, "--help"); err == nil {
		t.Error("an option was taken for a tag name")
	}
}

func TestTagCommitPeelsTheTag(t *testing.T) {
	dir := sealRepo(t)
	head, _ := gitx.HeadFull(dir)
	if err := gitx.AnnotatedTag(dir, "dudes/seal/demo/run-1", "m"); err != nil {
		t.Fatal(err)
	}
	commitFile(t, dir, "b.md", "second")
	if got, ok := gitx.TagCommit(dir, "dudes/seal/demo/run-1"); !ok || got != head {
		t.Fatalf("got %q %v, want the commit it was made at, %s", got, ok, head)
	}
	if _, ok := gitx.TagCommit(dir, "no-such-tag"); ok {
		t.Fatal("a tag that is not there")
	}
	if _, ok := gitx.TagCommit(dir, "main"); ok {
		t.Fatal("a branch is not a tag")
	}
}

// The newest first by tag date, each with the subject of the commit it leads to — an annotated tag's
// commit and a lightweight tag's alike; a tag on a branch HEAD does not contain is not reachable.
func TestTagsMergedNewestFirst(t *testing.T) {
	dir := sealRepo(t)
	tagOn(t, dir, "demo-v1.0.0", "2026-01-01T10:00:00Z")
	commitFile(t, dir, "b.md", "second")
	tagOn(t, dir, "dudes/seal/demo/run-2", "2026-02-01T10:00:00Z")
	commitFile(t, dir, "c.md", "third: with a colon, a tab\tand a quote \"")
	gitOut(t, dir, "tag", "light")
	tagOn(t, dir, "other", "2026-03-01T10:00:00Z")
	gitOut(t, dir, "checkout", "-q", "-b", "side")
	commitFile(t, dir, "d.md", "on a branch")
	tagOn(t, dir, "dudes/seal/demo/side", "2026-04-01T10:00:00Z")
	gitOut(t, dir, "checkout", "-q", "main")

	got, err := gitx.TagsMerged(dir)
	want := []gitx.Tag{
		{Name: "other", Subject: "third: with a colon, a tab\tand a quote \""},
		{Name: "dudes/seal/demo/run-2", Subject: "second"},
		{Name: "demo-v1.0.0", Subject: "first"},
	}
	// the lightweight tag is dated by its commit, which is the newest in this repository: it comes
	// first, or among the first, whatever the clock says — what is fixed is its subject
	for i, tag := range got {
		if tag.Name == "light" {
			if tag.Subject != want[0].Subject {
				t.Errorf("a lightweight tag leads to %q, want %q", tag.Subject, want[0].Subject)
			}
			got = append(got[:i], got[i+1:]...)
			break
		}
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, %v, want %+v", got, err, want)
	}
}

// Trailers are git's own reading of the last paragraph of a message.
func TestTrailers(t *testing.T) {
	dir := sealRepo(t)
	commitFile(t, dir, "b.md", "chore: x\n\nSealed-by: Ada Sealer <ada@example.com>\nOther: 1")
	sealed, _ := gitx.HeadFull(dir)
	commitFile(t, dir, "c.md", "chore: y\n\nSealed-by: Mallory <m@example.com>\n\nmore text, so that is no trailer")
	body, _ := gitx.HeadFull(dir)

	if got, err := gitx.Trailers(dir, sealed, "sealed-by"); err != nil || !reflect.DeepEqual(got, []string{"Ada Sealer <ada@example.com>"}) {
		t.Fatalf("trailers %q, %v", got, err)
	}
	if got, err := gitx.Trailers(dir, sealed, "Missing"); err != nil || len(got) != 0 {
		t.Fatalf("a trailer that is not there: %q, %v", got, err)
	}
	if got, err := gitx.Trailers(dir, body, "Sealed-by"); err != nil || len(got) != 0 {
		t.Fatalf("a line in the body is no trailer: %q, %v", got, err)
	}
}

// plumb runs git with input on its standard input and returns what it printed.
func plumb(t *testing.T, dir, input string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

// A note is replaced when the commit is saved again, and the notes ref keeps every save: all the
// versions of one commit's note come back, newest first, and none of another commit's.
func TestNoteVersions(t *testing.T) {
	dir := sealRepo(t)
	first, _ := gitx.HeadFull(dir)
	commitFile(t, dir, "b.md", "second")
	second, _ := gitx.HeadFull(dir)

	if got, err := gitx.NoteVersions(dir, "dudes/run", first); err != nil || got != nil {
		t.Fatalf("no notes ref yet: %q, %v", got, err)
	}
	// no notes ref is no note; a git that fails for another reason is an error, not "no note"
	if got, err := gitx.NoteVersions(t.TempDir(), "dudes/run", first); err == nil {
		t.Fatalf("not a repository: %q, want an error", got)
	}
	for _, text := range []string{"v1", "v2", "v3"} {
		gitOut(t, dir, "checkout", "-q", first)
		if err := gitx.AddNote(dir, "dudes/run", text); err != nil {
			t.Fatal(err)
		}
	}
	gitOut(t, dir, "checkout", "-q", second)
	if err := gitx.AddNote(dir, "dudes/run", "of the second"); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "checkout", "-q", "main")

	if got, err := gitx.NoteVersions(dir, "dudes/run", first); err != nil || !reflect.DeepEqual(got, []string{"v3", "v2", "v1"}) {
		t.Fatalf("versions of the first commit: %q, %v", got, err)
	}
	if got, err := gitx.NoteVersions(dir, "dudes/run", second); err != nil || !reflect.DeepEqual(got, []string{"of the second"}) {
		t.Fatalf("versions of the second commit: %q, %v", got, err)
	}
	if got, err := gitx.NoteVersions(dir, "dudes/other", first); err != nil || got != nil {
		t.Fatalf("another notes ref: %q, %v", got, err)
	}
	// the note is on the commit, not on a commit with a similar hash: the head of the branch has none
	gitOut(t, dir, "commit", "-q", "--allow-empty", "-m", "third")
	third, _ := gitx.HeadFull(dir)
	if got, err := gitx.NoteVersions(dir, "dudes/run", third); err != nil || got != nil {
		t.Fatalf("a commit without a note: %q, %v", got, err)
	}
}

// Git splits the notes of a big repository into directories named by the first characters of the
// commit's hash; the versions are found in a tree as git writes it then, one level or two.
func TestNoteVersionsInAFannedOutNotesTree(t *testing.T) {
	dir := sealRepo(t)
	commit, _ := gitx.HeadFull(dir)
	note := func(text string) string { return plumb(t, dir, text, "hash-object", "-w", "--stdin") }
	tree := func(entries ...string) string { return plumb(t, dir, strings.Join(entries, "\n")+"\n", "mktree") }
	save := func(root, parent, ref string) string {
		args := []string{"commit-tree", root, "-m", "notes"}
		if parent != "" {
			args = append(args, "-p", parent)
		}
		sha := plumb(t, dir, "", args...)
		gitOut(t, dir, "update-ref", ref, sha)
		return sha
	}
	for name, path := range map[string][]string{
		"one level":  {commit[:2], commit[2:]},
		"two levels": {commit[:2], commit[2:4], commit[4:]},
	} {
		// a notes tree that holds the commit's note at path, as git lays it out
		layout := func(blob string) string {
			leaf := fmt.Sprintf("100644 blob %s\t%s", blob, path[len(path)-1])
			node := tree(leaf)
			for i := len(path) - 2; i >= 0; i-- {
				node = tree(fmt.Sprintf("040000 tree %s\t%s", node, path[i]))
			}
			return node
		}
		ref := "refs/notes/fan" + strings.ReplaceAll(name, " ", "")
		first := save(layout(note("old\n")), "", ref)
		save(layout(note("new\n")), first, ref)
		got, err := gitx.NoteVersions(dir, strings.TrimPrefix(ref, "refs/notes/"), commit)
		if err != nil || !reflect.DeepEqual(got, []string{"new", "old"}) {
			t.Errorf("%s: %q, %v, want both versions, newest first", name, got, err)
		}
	}
}
