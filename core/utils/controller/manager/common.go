package manager

import (
	"errors"
	"strings"
	"time"

	"github.com/masx200/multinodewatchpanel/core/utils/cmd"
	"github.com/masx200/multinodewatchpanel/core/utils/ssh"
)

func handlerErr(out string, err error) error {
	if err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	return nil
}

func run(client *ssh.SSHClient, name string, args ...string) (string, error) {
	if client == nil {
		return cmd.NewCommandMgr(cmd.WithTimeout(10*time.Second)).RunWithStdoutBashCf("LANGUAGE=en_US:en %s %s", name, strings.Join(args, " "))
	}
	return client.Runf("LANGUAGE=en_US:en %s %s", name, strings.Join(args, " "))
}
