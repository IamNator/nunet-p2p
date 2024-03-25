package pkg

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// RunCmd executes the given command with the provided arguments
func RunCmd(name string, args ...string) (outputs []string, pid int, err error) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic")
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	fmt.Printf("Executing command: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)

	// get the outputs
	outputs = []string{}

	// Attach the stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, 0, fmt.Errorf("error attaching stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, 0, fmt.Errorf("error attaching stderr pipe: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	// get the outputs
	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := stdout.Read(buf)
			if n > 0 {
				outputs = append(outputs, "Info: "+strings.TrimSpace(string(buf[:n])))
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := stderr.Read(buf)
			if n > 0 {
				outputs = append(outputs, "Error: "+strings.TrimSpace(string(buf[:n])))
			}
			if err != nil {
				break
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, 0, fmt.Errorf("error starting command: %w", err)
	}

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	defer wg.Wait()
	select {
	case <-time.After(time.Minute / 2):
		cmd.Process.Kill()
		return outputs, 0, fmt.Errorf("command timed out")
	case err := <-done:
		if err != nil {
			return outputs, 0, fmt.Errorf("error waiting for command to finish: %w", err)
		}
	}

	fmt.Println("Command executed successfully")
	return outputs, cmd.ProcessState.Pid(), nil
}
