// Package awake keeps the host from going to sleep while the engine runs long external work.
// Standby freezes the WSL2 VM that runs Docker, so a worker or test gate stalls for as long
// as the laptop sleeps (seen in an instance trial).
package awake

// Hold asks the OS to keep the system awake until release is called. Release is idempotent.
// Best effort: where the OS has no such request, or it fails, Hold does nothing.
func Hold() (release func()) {
	_, release = hold()
	return release
}
