package skill

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// The naming contract of the framework repo's skills (CORE §5): the name is the task, the folder
// the level — framework/, lead/, spec/, dev/, shared/ — each folder with its README.md. A name is
// the agentskills.io name of its directory and unique across the folders (skills are loaded by
// name). A framework skill (framework/) says it is DevOps only and no recipe skill, no recipe names
// one, and the engine refuses to hand one to a worker — by its folder, so a new one needs no list.
func TestSkillNamesAndFolders(t *testing.T) {
	_, this, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(this), "..", "..")
	folders, err := os.ReadDir(filepath.Join(root, Dir))
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{FrameworkFolder: true, "lead": true, "spec": true, "dev": true, "shared": true}
	seen := map[string]string{}
	var framework []string
	for _, f := range folders {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".go") {
			continue // skills.go embeds the level folders into the binary
		}
		if !f.IsDir() || !allowed[f.Name()] {
			t.Errorf("skills/%s: the framework repo groups skills in framework/, lead/, spec/, dev/, shared/", f.Name())
			continue
		}
		if _, err := os.Stat(filepath.Join(root, Dir, f.Name(), "README.md")); err != nil {
			t.Errorf("skills/%s: the folder needs its README.md (what its skills share)", f.Name())
		}
		subs, _ := os.ReadDir(filepath.Join(root, Dir, f.Name()))
		for _, s := range subs {
			if !s.IsDir() {
				continue
			}
			name := s.Name()
			if prev, dup := seen[name]; dup {
				t.Errorf("%s is in skills/%s/ and skills/%s/ — names are unique, skills load by name", name, prev, f.Name())
			}
			seen[name] = f.Name()
			data, err := os.ReadFile(filepath.Join(root, Dir, f.Name(), name, FileName))
			if err != nil {
				t.Errorf("skills/%s/%s: %v", f.Name(), name, err)
				continue
			}
			body := strings.ReplaceAll(string(data), "\r\n", "\n")
			if !ValidName(name) || !strings.Contains(body, "\nname: "+name+"\n") {
				t.Errorf("skills/%s/%s: frontmatter name must equal the directory, agentskills.io charset", f.Name(), name)
			}
			isFramework := f.Name() == FrameworkFolder
			if strings.Contains(body, "Workspace DevOps only") != isFramework {
				t.Errorf("skills/%s/%s: a DevOps-only description belongs in skills/%s/ and only there", f.Name(), name, FrameworkFolder)
			}
			if !isFramework {
				continue
			}
			framework = append(framework, name)
			if !strings.Contains(strings.ToLower(body), "recipe") {
				t.Errorf("%s must forbid recipe use", name)
			}
			if _, err := LoadMany(root, []string{name}); err == nil {
				t.Errorf("LoadMany handed framework skill %s to a worker", name)
			}
		}
	}
	if len(framework) == 0 {
		t.Fatal("no framework skills found in skills/framework/")
	}
	// The workspace AGENTS.md table is how DevOps finds them: every framework skill is listed,
	// every listed one exists.
	agents, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range framework {
		if !strings.Contains(string(agents), "`skills/"+FrameworkFolder+"/"+name+"/SKILL.md`") {
			t.Errorf("AGENTS.md §DevOps skills does not list %s", name)
		}
	}
	for _, m := range regexp.MustCompile("`skills/"+FrameworkFolder+"/([a-z0-9-]+)/SKILL.md`").FindAllStringSubmatch(string(agents), -1) {
		if seen[m[1]] != FrameworkFolder {
			t.Errorf("AGENTS.md lists skills/%s/%s, which does not exist", FrameworkFolder, m[1])
		}
	}
	err = filepath.WalkDir(filepath.Join(root, "recipes"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
			return err
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, name := range framework {
			if strings.Contains(string(data), `"`+name+`"`) {
				t.Errorf("recipe %s names framework skill %s", path, name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"-tdd", "tdd-", "pm--review", "PM-review", strings.Repeat("a", 65)} {
		if ValidName(bad) {
			t.Errorf("ValidName(%q) = true", bad)
		}
	}
}

// An instance repo is one level: its skills sit directly in skills/, and the engine finds them there.
func TestRelPathInstanceAndFramework(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{"skills/lab-check", "skills/dev/tdd"} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(p)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(p), FileName), []byte("---\nname: x\n---\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got := RelPath(dir, "lab-check"); got != "skills/lab-check/SKILL.md" {
		t.Errorf("instance skill: %s", got)
	}
	if got := RelPath(dir, "tdd"); got != "skills/dev/tdd/SKILL.md" {
		t.Errorf("framework skill: %s", got)
	}
}
