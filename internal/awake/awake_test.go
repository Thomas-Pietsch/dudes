package awake

import "testing"

func TestHoldNestedRelease(t *testing.T) {
	outer := Hold()
	inner := Hold()
	inner()
	inner() // idempotent
	outer()
}
