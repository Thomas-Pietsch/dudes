// Package runstate keeps engine run progress in git (restart-safe, CORE §10): the state is a note
// on the newest commit, refs/notes/dudes/run. The newest state note in HEAD's history is the current
// state, so the state follows the history — reset the branch and the run state goes back with it.
// Every save is a commit on the notes ref: its log is the history of the run.
package runstate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

// NotesRef holds the run state: refs/notes/dudes/run.
const NotesRef = "dudes/run"

const (
	resetNote = "reset"             // written by Clear: no run from here on
	depth     = 1000                // commits searched back for the newest state note
	legacyRel = ".dudes/state.json" // before run state moved into git; migrated on first Load
)

// Status of a run.
type Status string

const (
	StatusRunning         Status = "running"
	StatusWaiting         Status = "waiting_human"
	StatusWaitingDecision Status = "waiting_decision"
	StatusSealed          Status = "sealed"
	StatusFailed          Status = "failed"
)

// State is durable run progress (not chat history).
type State struct {
	RunID     string   `json:"run_id"`
	Phase     phase.ID `json:"phase"`
	Status    Status   `json:"status"`
	Prompt    string   `json:"prompt,omitempty"`
	Feedback  string   `json:"feedback,omitempty"` // last human reject text
	Cycle     int      `json:"cycle,omitempty"`    // increments on each reject
	TaskID    string   `json:"task_id,omitempty"`  // current in-phase task (F65)
	Completed []string `json:"completed_tasks,omitempty"`
	Approved  []string `json:"approved,omitempty"` // task ids the human decided (F66)
	Decision  string   `json:"decision,omitempty"` // last decide note
	// Escalation is the worker's open BLOCKED:/CONTRACT: report (F74); cleared when the work succeeds.
	Escalation string `json:"escalation,omitempty"`
	// Redo holds, per task the human sent back (dudes redo), why; a note goes to that task's next
	// worker only and clears when the task completes again.
	Redo      map[string]string `json:"redo,omitempty"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// HasCompleted reports whether the in-phase task finished.
func (st *State) HasCompleted(taskID string) bool {
	if st == nil || taskID == "" {
		return false
	}
	for _, a := range st.Completed {
		if a == taskID {
			return true
		}
	}
	return false
}

// MarkCompleted records a finished in-phase task.
func (st *State) MarkCompleted(taskID string) {
	if st == nil || taskID == "" || st.HasCompleted(taskID) {
		return
	}
	st.Completed = append(st.Completed, taskID)
	st.TaskID = taskID
	delete(st.Redo, taskID)
}

// Reopen sends a finished task back with the human's note: it is no longer completed or approved,
// the next step runs it again and its worker gets the note (a red test that tests the wrong thing).
func (st *State) Reopen(taskID, note string) {
	if st == nil || taskID == "" {
		return
	}
	st.Completed = without(st.Completed, taskID)
	st.Approved = without(st.Approved, taskID)
	if st.Redo == nil {
		st.Redo = map[string]string{}
	}
	st.TaskID, st.Redo[taskID] = taskID, note
	if st.Status == StatusWaitingDecision {
		st.Status = StatusRunning
	}
}

// TaskFeedback is the human feedback the current task's worker gets: a run-level reject and, when
// this task was sent back, its redo note.
func (st *State) TaskFeedback() string {
	if st == nil {
		return ""
	}
	if note := st.Redo[st.TaskID]; note != "" {
		return strings.TrimSpace(st.Feedback + "\n\n" + note)
	}
	return st.Feedback
}

func without(ids []string, id string) []string {
	var out []string
	for _, x := range ids {
		if x != id {
			out = append(out, x)
		}
	}
	return out
}

// HasApproved reports whether taskID was recorded by dudes decide.
func (st *State) HasApproved(taskID string) bool {
	if st == nil || taskID == "" {
		return false
	}
	for _, a := range st.Approved {
		if a == taskID {
			return true
		}
	}
	return false
}

// Approve records a human decision for a task (F66).
func (st *State) Approve(taskID, note string) {
	if st == nil || taskID == "" {
		return
	}
	if !st.HasApproved(taskID) {
		st.Approved = append(st.Approved, taskID)
	}
	st.Decision = note
	if st.Status == StatusWaitingDecision {
		st.Status = StatusRunning
	}
}

// Path names where the run state lives, for humans.
func Path(workspace string) string { return "refs/notes/" + NotesRef }

// Load returns the current run state: the newest state note in HEAD's history. No note (or a reset
// note) → nil, false, nil. A state.json from before the move into git is migrated once.
func Load(workspace string) (*State, bool, error) {
	_, text, ok, err := gitx.LatestNote(workspace, NotesRef, depth)
	if err != nil {
		return nil, false, fmt.Errorf("read run state: %w", err)
	}
	if !ok {
		return migrate(workspace)
	}
	if strings.TrimSpace(text) == resetNote {
		return nil, false, nil
	}
	var st State
	if err := json.Unmarshal([]byte(text), &st); err != nil {
		return nil, false, fmt.Errorf("parse run state: %w", err)
	}
	return &st, true, nil
}

// Save records st as the state note on HEAD. A repository without commits gets an empty start
// commit first — the start of the run is the first entry of the record.
func Save(workspace string, st *State) error {
	if st == nil {
		return fmt.Errorf("nil state")
	}
	st.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	if err := ensureCommit(workspace, st.RunID); err != nil {
		return err
	}
	if err := gitx.AddNote(workspace, NotesRef, string(data)+"\n"); err != nil {
		return fmt.Errorf("write run state: %w", err)
	}
	return nil
}

func ensureCommit(workspace, runID string) error {
	return gitx.EnsureHead(workspace, fmt.Sprintf("chore(run): start %s [🔹engine]", runID))
}

// migrate moves a state.json from before the move into git into a note, then removes the file.
func migrate(workspace string) (*State, bool, error) {
	legacy := filepath.Join(workspace, filepath.FromSlash(legacyRel))
	data, err := os.ReadFile(legacy)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read run state: %w", err)
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, false, fmt.Errorf("parse run state: %w", err)
	}
	if err := Save(workspace, &st); err != nil {
		return nil, false, err
	}
	_ = os.Remove(legacy)
	return &st, true, nil
}

// NewRun creates initial state at the registry's first phase.
func NewRun(runID, prompt string, reg *phase.Registry) *State {
	return &State{
		RunID:  runID,
		Phase:  reg.First(),
		Status: StatusRunning,
		Prompt: prompt,
	}
}

// Clear ends the current run state: a reset note on HEAD. Earlier states stay in the notes history.
func Clear(workspace string) error {
	_ = os.Remove(filepath.Join(workspace, filepath.FromSlash(legacyRel)))
	if head, _ := gitx.Head(workspace); head == "" {
		return nil // nothing recorded yet
	}
	if err := gitx.AddNote(workspace, NotesRef, resetNote+"\n"); err != nil {
		return fmt.Errorf("clear run state: %w", err)
	}
	return nil
}
