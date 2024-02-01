package main

import (
	"io"
	"os/exec"
)

// Starts a command for winows OS
func startCmd(cmd string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	c := exec.Command("cmd", "/c", cmd)
	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	err = c.Start()
	return c, stdout, stderr, err
}
