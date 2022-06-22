package executor

import "fmt"

type ExecutorPool struct {
	name  string
	store map[string]Executor
}


func NewExecutorPool(name string) Pool {
	return &ExecutorPool{
		name: name,
		store: new(map[string]Executor),
	}
}

func (pool *ExecutorPool) Get(name string) (Executor, error) {
	res, ok := pool.store[name]
	if ok {
		return res, nil
	}
	return nil, fmt.Errorf("Executor %s does not exist in pool %s", name, pool.name)
}

func (pool *ExecutorPool) Register(executor Executor) error {
	_, ok := pool.store[executor.GetName()]
	if ok {
		return fmt.Errorf("Executor %s exists in the pool %s", executor.GetName(), pool.name)
	}
	pool.store[executor.GetName()] = executor
	return nil
}


var AgentPool Pool
func init()  {
	AgentPool = NewExecutorPool("agent-pool")
}

