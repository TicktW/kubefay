package executor

import (
	"fmt"
	"os/exec"
)

type BaseExecutor struct {
	name    string
	cache   *map[string]interface{}
}

func NewExecutor(name string) *BaseExecutor {
	return &BaseExecutor{
		name:    name,
		cache:   new(map[string]interface{}),
	}
}

func (executor *BaseExecutor) Execute(args ...string) Result{

	cmd := exec.Command(executor.name, args...)
	fmt.Println(cmd.String())
	out, err := cmd.CombinedOutput()

	return &ExeResult{
		status: cmd.ProcessState.ExitCode(),
		output: string(out),
		err: err,
		cmd: cmd.String(),
	}
}

func (executor *BaseExecutor) GetName() string {
	return executor.name
}
