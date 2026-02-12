package opener

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Open opens a file or URL using the default application for the current OS.
func Open(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// 'cmd /c start' handles file associations automatically
		// "start" command helps detaching the process so the CLI doesn't block if we want.
		// However, for consistency with 'exec.Command', we might want to wait or starts Detached.
		// 'start "" "path"' is the syntax. The empty string is for the window title (required if path is quoted).
		cmd = exec.Command("cmd", "/c", "start", "", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Connect/Disconnect standard streams?
	// For 'start' on windows it generally detaches nicely.
	// For 'xdg-open' it also usually delegates.
	// We run it and wait for the 'start' command itself to finish (which is instant),
	// not the application it opens.
	return cmd.Run()
}
