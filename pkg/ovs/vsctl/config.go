package vsctl

import "github.com/TicktW/kubefay/pkg/executor"

const BASECOMMAND = "ovs-vsctl"

var vsctlExecutor *executor.Executor

func init() {
	if vsctlExecutor == nil {
		vsctlExecutor = executor.NewEngine(BASECOMMAND, BASECOMMAND)
	}
}
