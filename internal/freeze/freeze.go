// Package freeze records a spec freeze before implementation (CORE frozen-spec intent): the hash
// of every normative input, kept as a git note on the commit it was frozen at (refs/notes/dudes/freeze,
// CORE §10). The newest freeze note in HEAD's history is the freeze in force.
package freeze

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/artifact"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/itempack"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

// NotesRef holds freeze records: refs/notes/dudes/freeze.
const NotesRef = "dudes/freeze"

// legacyRel is where freezes lived before they moved into git; migrated on first Load.
const legacyRel = ".dudes/spec_freeze.json"

// Input is one frozen normative file with provenance (F52, F59).
type Input struct {
	ID       string `json:"id"`
	Root     string `json:"root,omitempty"`
	Path     string `json:"path"`
	SHA256   string `json:"sha256"`
	Repo     string `json:"repo,omitempty"`
	Revision string `json:"revision,omitempty"`
}

// Record is the durable freeze evidence.
type Record struct {
	Product  string    `json:"product"`
	PRD      string    `json:"prd"`
	PRDRoot  string    `json:"prd_root,omitempty"`
	TA       string    `json:"ta,omitempty"`
	SHA256   string    `json:"sha256"`
	Inputs   []Input   `json:"inputs,omitempty"`
	FrozenAt time.Time `json:"frozen_at"`
	RunID    string    `json:"run_id,omitempty"`
}

// Path names where freeze records live, for humans.
func Path() string { return "refs/notes/" + NotesRef }

// HashBytes returns hex SHA-256 of data.
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// HashFile hashes a spec input as its repository's HEAD holds it (gitx.ReadCommitted): the
// freeze pins committed content, so a working copy someone is still editing is not drift.
func HashFile(abs string) (string, error) {
	data, err := gitx.ReadCommitted(abs)
	if err != nil {
		return "", err
	}
	return HashBytes(data), nil
}

// resolvePrimaryPRD hashes the instance PRD from the spec-set when present,
// otherwise the engine-named workspace file (F13/F59). A registered PRD
// does not fall back to {product}_PRD.md — that would invent a second PRD.
func resolvePrimaryPRD(workspaceRoot, product string) (root, rel, sum string, err error) {
	set, ok, err := specset.Load(workspaceRoot, product)
	if err != nil {
		return "", "", "", err
	}
	if ok && set != nil {
		if e, found := set.Find("prd"); found {
			abs, rerr := set.ResolveAbs(workspaceRoot, e)
			if rerr != nil {
				return "", "", "", rerr
			}
			sum, herr := HashFile(abs)
			if herr != nil {
				return "", e.Path, "", herr
			}
			root = e.Root
			if root == "" {
				root = "workspace"
			}
			return root, e.Path, sum, nil
		}
	}
	rel, sum, err = HashPRD(workspaceRoot, product)
	return "workspace", rel, sum, err
}

// HashPRD hashes the workspace PRD as committed (HashFile).
func HashPRD(workspaceRoot, product string) (rel, sum string, err error) {
	rel, ok := artifact.Path(product, phase.Spec)
	if !ok {
		return "", "", fmt.Errorf("no PRD path")
	}
	abs, err := workspace.Resolve(workspaceRoot, rel)
	if err != nil {
		return rel, "", err
	}
	sum, err = HashFile(abs)
	return rel, sum, err
}

// Write creates/updates the freeze record from current PRD (+ optional TA path).
func Write(workspaceRoot, product, runID string) (*Record, error) {
	return write(workspaceRoot, product, runID, false)
}

// WriteAll freezes the PRD plus every spec-set / contract input that exists (F59).
func WriteAll(workspaceRoot, product, runID string) (*Record, error) {
	return write(workspaceRoot, product, runID, true)
}

func write(workspaceRoot, product, runID string, all bool) (*Record, error) {
	if err := RequireConsistent(workspaceRoot, product); err != nil {
		return nil, err
	}
	prdRoot, prdRel, sum, err := resolvePrimaryPRD(workspaceRoot, product)
	if err != nil {
		return nil, fmt.Errorf("freeze: %w", err)
	}
	rec := &Record{
		Product:  product,
		PRD:      prdRel,
		PRDRoot:  prdRoot,
		SHA256:   sum,
		FrozenAt: time.Now().UTC(),
		RunID:    runID,
	}
	if ta := product + "_TA.md"; artifact.Exists(workspaceRoot, ta) {
		rec.TA = ta
	}
	if all {
		inputs, err := collectInputs(workspaceRoot, product, rec)
		if err != nil {
			return nil, err
		}
		rec.Inputs = inputs
	}
	if err := save(workspaceRoot, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// save attaches rec as the freeze note on HEAD (a repository without commits gets its first one).
func save(workspaceRoot string, rec *Record) error {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	if err := gitx.EnsureHead(workspaceRoot, fmt.Sprintf("chore(spec): freeze %s [🔹engine]", rec.Product)); err != nil {
		return fmt.Errorf("freeze: %w", err)
	}
	if err := gitx.AddNote(workspaceRoot, NotesRef, string(data)+"\n"); err != nil {
		return fmt.Errorf("freeze: %w", err)
	}
	return nil
}

func collectInputs(workspaceRoot, product string, rec *Record) ([]Input, error) {
	seen := map[string]bool{}
	var out []Input
	add := func(in Input) {
		key := in.Root + "|" + in.Path
		if in.Path == "" || seen[key] {
			return
		}
		seen[key] = true
		out = append(out, in)
	}

	set, _, err := specset.Load(workspaceRoot, product)
	if err != nil {
		return nil, err
	}

	prdRoot := rec.PRDRoot
	if prdRoot == "" {
		prdRoot = "workspace"
	}
	wsRepo, wsRev := gitx.Provenance(workspaceRoot)
	prdRepo, prdRev := wsRepo, wsRev
	if set != nil {
		if rootAbs, rerr := set.RootAbs(workspaceRoot, prdRoot); rerr == nil {
			prdRepo, prdRev = gitx.Provenance(rootAbs)
		}
	}
	add(Input{
		ID:       "prd",
		Root:     prdRoot,
		Path:     rec.PRD,
		SHA256:   rec.SHA256,
		Repo:     prdRepo,
		Revision: prdRev,
	})
	if rec.TA != "" && artifact.Exists(workspaceRoot, rec.TA) {
		abs, err := workspace.Resolve(workspaceRoot, rec.TA)
		sum := ""
		if err == nil {
			sum, err = HashFile(abs)
		}
		if err == nil {
			add(Input{
				ID:       "ta",
				Root:     "workspace",
				Path:     rec.TA,
				SHA256:   sum,
				Repo:     wsRepo,
				Revision: wsRev,
			})
		}
	}
	if set != nil {
		for _, e := range set.Entries {
			abs, err := set.ResolveAbs(workspaceRoot, e)
			if err != nil {
				continue
			}
			sum, err := HashFile(abs)
			if err != nil {
				continue
			}
			rootAbs, err := set.RootAbs(workspaceRoot, e.Root)
			if err != nil {
				rootAbs = workspaceRoot
			}
			repo, rev := gitx.Provenance(rootAbs)
			id := e.ID
			if id == "" {
				id = e.Kind
			}
			add(Input{ID: id, Root: e.Root, Path: e.Path, SHA256: sum, Repo: repo, Revision: rev})
		}
	}

	con, ok, err := contract.Load(workspaceRoot)
	if err != nil {
		return nil, err
	}
	if ok && con != nil {
		for _, src := range con.Sources {
			rootName := src.Root
			if rootName == "" {
				rootName = "workspace"
			}
			var abs string
			if set != nil {
				abs, err = set.ResolveAbs(workspaceRoot, specset.Entry{Root: rootName, Path: src.Path})
			} else {
				abs, err = workspace.Resolve(workspaceRoot, src.Path)
			}
			if err != nil {
				continue
			}
			sum, err := HashFile(abs)
			if err != nil {
				continue
			}
			rootAbs := workspaceRoot
			if set != nil {
				if r, err := set.RootAbs(workspaceRoot, rootName); err == nil {
					rootAbs = r
				}
			}
			repo, rev := gitx.Provenance(rootAbs)
			if src.Rev != "" {
				rev = src.Rev
			}
			add(Input{ID: src.ID, Root: rootName, Path: src.Path, SHA256: sum, Repo: repo, Revision: rev})
		}
	}
	return out, nil
}

// RequireConsistent refuses a freeze while the spec-set has consistency
// findings (F69). Without .dudes/profile.json there is nothing to check.
func RequireConsistent(workspaceRoot, product string) error {
	s, ok, err := itempack.OpenCommitted(workspaceRoot, product)
	if err != nil {
		return fmt.Errorf("freeze: %w", err)
	}
	if !ok {
		return nil
	}
	if f := s.Check(); len(f) > 0 {
		return fmt.Errorf("freeze: spec-set has %d consistency finding(s) — run: dudes item check (first: %s)", len(f), f[0])
	}
	return nil
}

// Load returns the freeze in force: the newest freeze note in HEAD's history. A spec_freeze.json
// from before the move into git is migrated once.
func Load(workspaceRoot string) (*Record, bool, error) {
	_, text, ok, err := gitx.LatestNote(workspaceRoot, NotesRef, 1000)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return migrate(workspaceRoot)
	}
	var rec Record
	if err := json.Unmarshal([]byte(text), &rec); err != nil {
		return nil, false, err
	}
	return &rec, true, nil
}

func migrate(workspaceRoot string) (*Record, bool, error) {
	legacy := filepath.Join(workspaceRoot, filepath.FromSlash(legacyRel))
	data, err := os.ReadFile(legacy)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, false, err
	}
	if err := save(workspaceRoot, &rec); err != nil {
		return nil, false, err
	}
	_ = os.Remove(legacy)
	return &rec, true, nil
}

// RequireValid enforces freeze exists and matches current inputs (impl gate).
func RequireValid(workspaceRoot, product string) error {
	rec, ok, err := Load(workspaceRoot)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("spec-freeze: missing %s — complete spec phase (auto-freeze) or run: dudes freeze", Path())
	}
	if rec.Product != "" && product != "" && rec.Product != product {
		return fmt.Errorf("spec-freeze: product mismatch frozen=%s config=%s", rec.Product, product)
	}
	_, _, sum, err := resolvePrimaryPRD(workspaceRoot, product)
	if err != nil {
		if len(rec.Inputs) == 0 {
			return fmt.Errorf("spec-freeze: %w", err)
		}
	} else if sum != rec.SHA256 {
		return fmt.Errorf("spec-freeze: PRD changed since freeze (re-run dudes freeze after updating %s)", rec.PRD)
	}
	if len(rec.Inputs) == 0 {
		return nil
	}
	set, _, err := specset.Load(workspaceRoot, product)
	if err != nil {
		return err
	}
	for _, in := range rec.Inputs {
		abs, err := resolveFrozen(workspaceRoot, set, in)
		if err != nil {
			return fmt.Errorf("spec-freeze: input %s: %w", in.ID, err)
		}
		got, err := HashFile(abs)
		if err != nil {
			return fmt.Errorf("spec-freeze: input %s: %w", in.ID, err)
		}
		if got != in.SHA256 {
			return fmt.Errorf("spec-freeze: %s changed since freeze (re-run dudes freeze)", in.Path)
		}
	}
	return nil
}

func resolveFrozen(workspaceRoot string, set *specset.Set, in Input) (string, error) {
	if set != nil {
		abs, err := set.ResolveAbs(workspaceRoot, specset.Entry{Root: in.Root, Path: in.Path})
		if err == nil {
			return abs, nil
		}
		if in.Root != "" && in.Root != "workspace" {
			return "", err
		}
	}
	return workspace.Resolve(workspaceRoot, in.Path)
}
