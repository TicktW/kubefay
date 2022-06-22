package vsctl

import "github.com/TicktW/kubefay/pkg/executor"

const VSCTLNAME = "ovs-vsctl"

func init() {
	baseOvsctl := executor.NewExecutor(VSCTLNAME)
	LockOvsctl := executor.NewLockExecutor(baseOvsctl)
	executor.AgentPool.Register(LockOvsctl)
}
