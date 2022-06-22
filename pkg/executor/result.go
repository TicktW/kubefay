package executor

type ExeResult struct{
	status int
	output string
	err error
	cmd string
}

func (result *ExeResult) GetStatus() int {
	return result.status
}

func (result *ExeResult) GetOutput() string {
	return result.output
}

func (result *ExeResult) GetCmdStr() string {
	return result.cmd
}

func (result *ExeResult) GetError() error{
	return result.err
}