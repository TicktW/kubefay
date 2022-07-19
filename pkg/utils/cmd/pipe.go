package cmd

import (
	"bytes"
	"os/exec"
	"strings"
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
	return &CmdList{
		head:   node,
		tail:   node,
		length: 0,
	}
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
	cmdList := NewCmdList()
	cmdList.AppendList(cmds...)

	cur := cmdList.GetHead()

	for cur != nil && cur.next != nil {
		p, err := cur.cmd.StdoutPipe()
		if err != nil {
			return "", err
		}
		cur.next.cmd.Stdin = p
		cur.cmd.Start()
		cur = cur.next
	}

	var buf bytes.Buffer
	cur.cmd.Stdout = &buf
	cur.cmd.Start()

	cur = cmdList.GetHead()
	for cur != nil {
		// fmt.Printf("%s\n", cur.cmd.String())
		err := cur.cmd.Wait()
		if err != nil {
			return "", err
		}
		cur = cur.next
	}

	return strings.Trim(buf.String(), "\n"), nil
}

func PipeCmdStr(cmd string) (string, error) {
	// fmt.Println(cmd)
	if strings.Contains(cmd, "awk") {
		bashCmd := exec.Command("bash", "-c", cmd)
		out, err := bashCmd.CombinedOutput()
		res := strings.Trim(string(out), "\n")
		return res, err
	}

	res := strings.Split(cmd, "|")
	var cmds []*exec.Cmd

	for _, cmdStr := range res {
		s := strings.Split(strings.Trim(cmdStr, " "), " ")
		name := s[0]
		args := s[1:]
		// if name == "awk" {
		// 	for idx, arg := range args {
		// 		args[idx] = strings.Trim(arg, "'")
		// 	}
		// }
		cmds = append(cmds, exec.Command(name, args...))
	}

	return PipeCmd(cmds...)
}
