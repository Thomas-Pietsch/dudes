//go:build !windows

package awake

func hold() (prev uint32, release func()) { return 0, func() {} }
