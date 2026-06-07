package open

import (
	"os/exec"
	"runtime"
)

func Open(filename string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filename)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", filename)
	case "linux":
		cmd = exec.Command("xdg-open", filename)
	default:
		return
	}
	_ = cmd.Start()
}
