package iptool

import "github.com/TicktW/kubefay/pkg/executor"

const IP_LINK_NAME = "ip link"
const IP_ADDR_NAME = "ip link"

func init() {
	baseIPLink := executor.NewExecutor(IP_LINK_NAME)
	lockIPLink := executor.NewLockExecutor(baseIPLink)
	executor.AgentPool.Register(lockIPLink)

	baseIPAddr:= executor.NewExecutor(IP_ADDR_NAME)
	lockIPAddr := executor.NewLockExecutor(baseIPAddr)
	executor.AgentPool.Register(lockIPAddr)
}
