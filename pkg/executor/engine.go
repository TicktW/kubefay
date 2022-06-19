package executor

import (
	"fmt"
	"os/exec"
	"sync"
)

type Executor struct {
	name    string
	baseCmd string
	mux     sync.RWMutex
	cache   *map[string]interface{}
}

func NewEngine(name string, baseCmd string) *Executor {
	var mux sync.RWMutex
	executor := Executor{
		name:    name,
		baseCmd: baseCmd,
		mux:     mux,
		cache:   new(map[string]interface{}),
	}
	return &executor
}

func (executor *Executor) Execute(args ...string) (string, error, int) {

	cmd := exec.Command(executor.baseCmd, args...)
	fmt.Println(cmd.String())
	out, err := cmd.CombinedOutput()
	status := cmd.ProcessState.ExitCode()

	return string(out), err, status
}

func (executor *Executor) ReadExecute(args ...string) (string, error, int) {
	executor.mux.RLock()
	defer executor.mux.RUnlock()

	return executor.Execute(args...)
}

func (executor *Executor) WriteExecute(args ...string) (string, error, int) {
	executor.mux.Lock()
	defer executor.mux.Unlock()

	return executor.Execute(args...)
}
