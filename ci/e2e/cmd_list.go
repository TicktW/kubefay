package e2e

import (
	"fmt"
	"io"
	"os/exec"
)

type CmdNode struct {
	cmd  *exec.Cmd
	next *CmdNode
}

type CmdList struct {
	head   *CmdNode
	tail   *CmdNode
	length int
}

func NewCmdList() *CmdList {
	node := &CmdNode{}
	head := &CmdList{
		head:   node,
		tail:   node,
		length: 0,
	}
	return head
}

func (cmdList *CmdList) Append(cmd *exec.Cmd) {
	node := &CmdNode{
		cmd: cmd,
	}
	cmdList.tail.next = node
	cmdList.tail = node
	cmdList.length++
}

func (cmdList *CmdList) AppendList(cmds ...*exec.Cmd) {
	for _, cmd := range cmds {
		cmdList.Append(cmd)
	}
}

func (cmdList *CmdList) GetHead() *CmdNode {
	return cmdList.head.next
}

func PipeCmd(cmds ...*exec.Cmd) (string, error) {
	// var r *io.PipeReader
	// var w *io.PipeWriter
	cmdList := NewCmdList()
	cmdList.AppendList(cmds...)
	fmt.Print(cmdList.length)

	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	cur := cmdList.GetHead()
	for cur != nil && cur.next != nil {

		cur.cmd.Stdout = w
		if err := cur.cmd.Run(); err != nil {
			return "", err
		}
		cur.next.cmd.Stdin = r
		cur = cur.next
	}
	res, err := cur.cmd.CombinedOutput()

	if err != nil {
		return "", err
	}

	return string(res), nil
}
