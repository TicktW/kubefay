package executor

import "sync"

type RLockExecutor struct {
	baseExecutor Executor
	mux          sync.RWMutex
}

func NewRLockExecutor(baseExecutor *BaseExecutor) *RLockExecutor {
	return &RLockExecutor{
		baseExecutor: baseExecutor,
		mux:          sync.RWMutex{},
	}
}

func (lockExecutor *RLockExecutor) Execute(args ...string) Result {
	lockExecutor.mux.Lock()
	defer lockExecutor.mux.Unlock()
	return lockExecutor.baseExecutor.Execute(args...)
}

func (lockExecutor *RLockExecutor) GetName() string {
	return lockExecutor.baseExecutor.GetName()
}

func (lockExecutor *RLockExecutor) Rexecute(args ...string) Result{
	lockExecutor.mux.RLock()
	defer lockExecutor.mux.RUnlock()
	return lockExecutor.baseExecutor.Execute(args...)
}