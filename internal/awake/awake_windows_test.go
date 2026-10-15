//go:build windows

package awake

import "testing"

func TestHoldWindowsSucceeds(t *testing.T) {
	prev, release := hold()
	defer release()
	if prev == 0 {
		t.Fatal("SetThreadExecutionState failed (returned 0)")
	}
}
