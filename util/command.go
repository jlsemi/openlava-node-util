package util

import (
	"bytes"
	"os/exec"
)

func ExecCommand(command string, args []string) (string, error) {
	var out, errOut bytes.Buffer

	cmd := exec.Command(command, args...)
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()

	if err != nil {
		return errOut.String(), err
	}

	return out.String(), err
}
