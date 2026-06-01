//go:build darwin

package main

import "runtime"

// Pin the main goroutine to the OS main thread for the whole process. The macOS
// now-playing bridge starts its NSApplication run loop there (an AppKit
// requirement), and the goroutine must not have migrated off the main thread by
// the time the player command runs. Harmless for the other commands.
func init() { runtime.LockOSThread() }
