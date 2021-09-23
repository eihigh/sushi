package sushi

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	tasks []*task

	wrapFormat = "%s ; code=$?; pwd > /tmp/wd; exit $code"
	baseCmd    = []string{"bash", "-c"}
)

type task struct {
	cmd     *exec.Cmd
	command string

	stdinPipe io.WriteCloser
}

func RunTask(workdir, command string) (taskID int, err error) {
	id := len(tasks)
	t := &task{
		command: command,
	}

	// wrap command
	arg := fmt.Sprintf(wrapFormat, command)
	args := append(baseCmd, arg)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = workdir
	t.cmd = cmd

	// set stdio
	in, err := cmd.StdinPipe()
	if err != nil {
		return 0, err
	}
	t.stdinPipe = in
	cmd.Stdout = &stdoutHandler{id}
	cmd.Stderr = &stderrHandler{id}

	// run
	go func() {
		cmd.Run()

		// cleanup
		wd, _ := os.ReadFile("/tmp/wd")
		exitc <- exitMsg{id, cmd.ProcessState.ExitCode(), strings.TrimSpace(string(wd))}
	}()

	tasks = append(tasks, t)
	return id, nil
}

func WriteStdin(taskID int, p []byte) (n int, err error) {
	// TODO: error handling
	t := tasks[taskID]
	return t.stdinPipe.Write(p)
}

func CloseStdin(taskID int) error {
	// TODO: error handling
	t := tasks[taskID]
	return t.stdinPipe.Close()
}
