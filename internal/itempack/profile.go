// Package itempack (contract: SPEC.md) assembles the complete context of one backlog work item
// from a spec-set (F68), checks spec-set consistency (F69) and derives item
// status from the QR (F70). Everything product-specific lives in a profile.
//
// The format is context-pack/1: `[[DOC:ID]]`, `[[DOC:A..B]]`, `[[DOC:§n]]`.
// DOC keys are spec-set entry IDs, so documents may live in any named root.
package itempack

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

const profileRel = ".dudes/profile.json"

// ProfilePath is the profile file relative to the workspace.
func ProfilePath() string { return profileRel }

// Owner maps an ID family to the document that defines it. First full match wins.
type Owner struct {
	Pattern string `json:"pattern"`
	Doc     string `json:"doc"`
	Note    string `json:"note,omitempty"`
}

// BacklogCfg locates work items inside the backlog document.
type BacklogCfg struct {
	Doc            string `json:"doc"`
	ItemsSection   string `json:"items_section"`   // prefix of the level-2 heading holding items
	ItemHeading    string `json:"item_heading"`    // groups: id, title
	OutlineRow     string `json:"outline_row"`     // groups: id, title
	ReleaseHeading string `json:"release_heading"` // group: release
	ReleaseField   string `json:"release_field"`   // group: release
	DoDField       string `json:"dod_field"`       // group: text
	ACLine         string `json:"ac_line,omitempty"`
}

// RulesSection names one section copied verbatim into RULES.md.
type RulesSection struct {
	Doc           string `json:"doc"`
	HeadingPrefix string `json:"heading_prefix"`
}

// StatusCfg names the QR document and its status marks.
type StatusCfg struct {
	Doc     string   `json:"doc"`
	Done    []string `json:"done"`
	Blocked []string `json:"blocked"`
	Open    []string `json:"open"`
	// NotApplicable marks an ID as deliberately out of scope ("not this item", retired) — only for
	// IDs in the row's first cell, and never over a real mark elsewhere (e.g. "—").
	NotApplicable []string `json:"not_applicable,omitempty"`
}

// RoadmapCfg finds release milestones.
type RoadmapCfg struct {
	Doc          string `json:"doc"`
	MilestoneRow string `json:"milestone_row"` // group: release
}

// TableDDLCfg attaches CREATE TABLE / TRIGGER / INDEX to table IDs.
type TableDDLCfg struct {
	Doc      string `json:"doc"`
	IDPrefix string `json:"id_prefix"`
}

// ConsistencyCfg holds the patterns of rules C3–C5.
type ConsistencyCfg struct {
	RoutePattern        string `json:"route_pattern"`
	RouteOwner          string `json:"route_owner"`
	RouteShapeInBacklog string `json:"route_shape_in_backlog"`
	DDLMarker           string `json:"ddl_marker"`
	UIAnchor            string `json:"ui_anchor"`
	PRDForbiddenRow     string `json:"prd_forbidden_row"`
	SkipLineMarker      string `json:"skip_line_marker,omitempty"`
}

// AcceptanceCfg says where the instance keeps its acceptance tests (BDD / E2E scenarios) and how a
// test names the requirement it proves (F77). Unit tests are not listed: they carry no ID.
type AcceptanceCfg struct {
	Paths []string `json:"paths"`          // workspace-relative directories of acceptance tests
	ID    string   `json:"id"`             // extended regex a test's name must contain, e.g. "BK-?[0-9]{3}"
	Gate  bool     `json:"gate"`           // true: an unnamed acceptance test fails the test gate; false: report only
	Goal  string   `json:"goal,omitempty"` // make goal that runs them (e.g. "e2e") — a hard gate after the unit tests (F35)
}

// Profile is instance data: ID ownership and the few format conventions per vertical.
type Profile struct {
	ProfileVersion         int            `json:"profile_version"`
	Product                string         `json:"product"`
	SharedDocuments        []string       `json:"shared_documents,omitempty"`
	Owners                 []Owner        `json:"owners"`
	MultiDefinitionAllowed []string       `json:"multi_definition_allowed,omitempty"`
	Backlog                BacklogCfg     `json:"backlog"`
	RulesSections          []RulesSection `json:"rules_sections,omitempty"`
	Status                 StatusCfg      `json:"status"`
	Roadmap                RoadmapCfg     `json:"roadmap"`
	TableDDL               TableDDLCfg    `json:"table_ddl"`
	ContextOrder           []string       `json:"context_order,omitempty"`
	Consistency            ConsistencyCfg `json:"consistency"`
	AcceptanceTests        AcceptanceCfg  `json:"acceptance_tests,omitempty"`
}

const defaultACLine = `^\s*[-*]\s+(?:\*\*)?(AC-\d+)(?:\*\*)?[\s:.]+(.*)$`

// ParseProfile decodes and validates a profile.
func ParseProfile(data []byte) (*Profile, error) {
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("profile: parse: %w", err)
	}
	if p.ProfileVersion != 1 {
		return nil, fmt.Errorf("profile: unsupported profile_version %d (expected 1)", p.ProfileVersion)
	}
	if len(p.Owners) == 0 {
		return nil, fmt.Errorf("profile: no owners")
	}
	if p.Backlog.Doc == "" || p.Status.Doc == "" || p.Roadmap.Doc == "" {
		return nil, fmt.Errorf("profile: backlog, status and roadmap docs are required")
	}
	if p.Backlog.ACLine == "" {
		p.Backlog.ACLine = defaultACLine
	}
	return &p, nil
}

// LoadProfile reads .dudes/profile.json. ok=false when the instance has none.
func LoadProfile(workspaceRoot string) (*Profile, bool, error) {
	abs, err := workspace.Resolve(workspaceRoot, profileRel)
	if err != nil {
		return nil, false, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("profile: %w", err)
	}
	p, err := ParseProfile(data)
	if err != nil {
		return nil, false, err
	}
	return p, true, nil
}

// compiled holds every profile regex, compiled once (RE2 syntax).
type compiled struct {
	owners       []*regexp.Regexp // ^(?:p)$ — full match
	ids          *regexp.Regexp   // (p1|p2|…) + trailing boundary
	itemHeading  *regexp.Regexp
	outlineRow   *regexp.Regexp
	releaseHead  *regexp.Regexp
	releaseField *regexp.Regexp
	dodField     *regexp.Regexp
	acLine       *regexp.Regexp
	milestoneRow *regexp.Regexp
	routePattern *regexp.Regexp
	routeShape   *regexp.Regexp
	prdForbidden *regexp.Regexp
}

func (p *Profile) compile() (*compiled, error) {
	c := &compiled{}
	var alts []string
	for _, o := range p.Owners {
		re, err := regexp.Compile(`^(?:` + o.Pattern + `)$`)
		if err != nil {
			return nil, fmt.Errorf("profile: owner %q: %w", o.Pattern, err)
		}
		c.owners = append(c.owners, re)
		alts = append(alts, o.Pattern)
	}
	var err error
	if c.ids, err = regexp.Compile(`(` + joinAlt(alts) + `)` + idTrail); err != nil {
		return nil, fmt.Errorf("profile: owners: %w", err)
	}
	b := p.Backlog
	cons := p.Consistency
	for _, x := range []struct {
		dst  **regexp.Regexp
		pat  string
		pre  string
		name string
	}{
		{&c.itemHeading, b.ItemHeading, "^(?:", "backlog.item_heading"},
		{&c.outlineRow, b.OutlineRow, "^(?:", "backlog.outline_row"},
		{&c.releaseHead, b.ReleaseHeading, "^(?:", "backlog.release_heading"},
		{&c.releaseField, b.ReleaseField, "(?:", "backlog.release_field"},
		{&c.dodField, b.DoDField, "(?m)(?:", "backlog.dod_field"},
		{&c.acLine, b.ACLine, "^(?:", "backlog.ac_line"},
		{&c.milestoneRow, p.Roadmap.MilestoneRow, "(?m)(?:", "roadmap.milestone_row"},
		{&c.routePattern, cons.RoutePattern, "(?:", "consistency.route_pattern"},
		{&c.routeShape, cons.RouteShapeInBacklog, "(?:", "consistency.route_shape_in_backlog"},
		{&c.prdForbidden, cons.PRDForbiddenRow, "^(?:", "consistency.prd_forbidden_row"},
	} {
		if x.pat == "" {
			continue
		}
		re, err := regexp.Compile(x.pre + x.pat + `)`)
		if err != nil {
			return nil, fmt.Errorf("profile: %s: %w", x.name, err)
		}
		*x.dst = re
	}
	if c.itemHeading == nil || c.dodField == nil {
		return nil, fmt.Errorf("profile: backlog.item_heading and backlog.dod_field are required")
	}
	return c, nil
}

func joinAlt(alts []string) string {
	out := ""
	for i, a := range alts {
		if i > 0 {
			out += "|"
		}
		out += a
	}
	return out
}
