/*
 * @Author: TicktW wxjpython@gmail.com
 * @Description: MIT License Copyright (C) 2022 TicktW@https://github.com/TicktW/kubefay
 */
package ovs

import (
	"fmt"
	"os/exec"
	"strings"
)

func RunCmd(cmdStr string) (string, error) {
	cmdList := strings.Split(cmdStr, " ")

	var cmd *exec.Cmd
	if len(cmdList) == 1 {
		cmd = exec.Command(cmdList[0])
	} else if len(cmdList) > 1 {
		cmd = exec.Command(cmdList[0], cmdList[1:]...)
	} else {
		return "", fmt.Errorf("execute command error: %s", cmdStr)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), err
}
