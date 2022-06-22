package executor

type Result interface {
	GetStatus() int
	GetOutput() string
	GetError() error
	GetCmdStr() string
}

type Executor interface {
	GetName() string
	Execute(args ...string) Result
}

type Pool interface {
	Register(executor Executor) error
	Get(name string) (Executor, error)
}
