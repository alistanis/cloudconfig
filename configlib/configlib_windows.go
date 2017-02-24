// +build windows
package configlib

var (
	// %userprofile% is the traditional windows user directory shortcut, but powershell allows ~/, so we'll try both
	homeShortcuts = []string{"%userprofile%", "~/"}
)
