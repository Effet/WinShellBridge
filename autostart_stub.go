//go:build !windows
// +build !windows

package main

// EnsureAutoStart is a no-op on non-Windows platforms to keep builds happy.
func EnsureAutoStart() error {
	return nil
}
