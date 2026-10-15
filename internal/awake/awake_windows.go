//go:build windows

package awake

import (
	"runtime"
	"sync"
	"syscall"
)

const (
	esContinuous     = 0x80000000
	esSystemRequired = 0x00000001
)

var setThreadExecutionState = syscall.NewLazyDLL("kernel32.dll").NewProc("SetThreadExecutionState")

// hold returns the thread's previous execution state (0: the request failed). The state
// belongs to the thread that set it, so a goroutine locked to its own thread sets, holds and
// clears it — the caller's goroutine may move between threads meanwhile.
func hold() (prev uint32, release func()) {
	if setThreadExecutionState.Find() != nil {
		return 0, func() {}
	}
	set, stop, done := make(chan uint32), make(chan struct{}), make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(done)
		r, _, _ := setThreadExecutionState.Call(esContinuous | esSystemRequired)
		set <- uint32(r)
		<-stop
		setThreadExecutionState.Call(esContinuous)
	}()
	prev = <-set
	var once sync.Once
	return prev, func() {
		once.Do(func() { close(stop); <-done })
	}
}
