//go:build windows

package terminal

import (
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/masx200/multinodewatchpanel/core/global"
)

const (
	DefaultCloseSignal  = syscall.SIGINT
	DefaultCloseTimeout = 10 * time.Second
)

type LocalCommand struct {
	closeSignal  syscall.Signal
	closeTimeout time.Duration

	cmd *exec.Cmd
	pty  *os.File
}

func NewCommand(script string) (*LocalCommand, error) {
	cmd := exec.Command("cmd.exe")
	if term := os.Getenv("TERM"); term != "" {
		cmd.Env = append(os.Environ(), "TERM="+term)
	}

	lcmd := &LocalCommand{
		closeSignal:  DefaultCloseSignal,
		closeTimeout: DefaultCloseTimeout,
		cmd:          cmd,
		pty:           nil,
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return lcmd, nil
}

func (lcmd *LocalCommand) Read(p []byte) (n int, err error) {
	return 0, nil // Windows 上暂不支持
}

func (lcmd *LocalCommand) Write(p []byte) (n int, err error) {
	return len(p), nil // Windows 上暂不支持
}

func (lcmd *LocalCommand) Close() error {
	if lcmd.cmd != nil && lcmd.cmd.Process != nil {
		_ = lcmd.cmd.Process.Kill()
	}
	return nil
}

func (lcmd *LocalCommand) ResizeTerminal(width int, height int) error {
	// Windows 上暂不支持 pty 调整大小
	return nil
}

func (lcmd *LocalCommand) Wait(quitChan chan bool) {
	if err := lcmd.cmd.Wait(); err != nil {
		global.LOG.Errorf("ssh session wait failed, err: %v", err)
		setQuit(quitChan)
	}
}
