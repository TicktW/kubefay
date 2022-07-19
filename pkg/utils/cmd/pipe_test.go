package cmd

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestPipeCmd(t *testing.T) {
	ls := exec.Command("ls")
	grep := exec.Command("grep", "test")
	fmt.Println("before pipe")
	out, err := PipeCmd(ls, grep)
	fmt.Println("\n=========")
	fmt.Println(err)
	fmt.Println(out)
	fmt.Println("=========")
}

func TestPipeCmdStr(t *testing.T) {
	out, err := PipeCmdStr("ls | grep test")
	fmt.Println("\n=========")
	fmt.Println(err)
	fmt.Println(out)
	fmt.Println("=========")
}
