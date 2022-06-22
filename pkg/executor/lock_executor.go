package executor

import "sync"

type LockExecutor struct {
	baseExecutor Executor
	mux          sync.Mutex
}

func NewLockExecutor(baseExecutor *BaseExecutor) *LockExecutor {
	return &LockExecutor{
		baseExecutor: baseExecutor,
		mux: sync.Mutex{},
	}
}

func (lockExecutor *LockExecutor) Execute(args ...string) Result {
	lockExecutor.mux.Lock()
	defer lockExecutor.mux.Unlock()
	return lockExecutor.baseExecutor.Execute(args...)
}

func (lockExecutor *LockExecutor) GetName() string {
	return lockExecutor.baseExecutor.GetName()
}
