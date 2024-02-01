package main

import (
	"os"
	"os/exec"
	"strconv"
)

// Starts a command for winows OS
func startCmd(cmd string) (*exec.Cmd, error) {
	c := exec.Command("cmd", "/c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := c.Start()
	return c, err
}

// Kills a process
func killCmd(cmd *exec.Cmd) error {
	pid := cmd.Process.Pid
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(pid))
	return kill.Run()
}
