package e2e

import (
	"fmt"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestE2E(t *testing.T) {
	// Only pass t into top-level Convey calls
	Convey("Kubefay Network E2E Test", t, func() {

		Convey("Agent shoudld in running state", func() {
			kubectl := exec.Command("kubectl", "-n", "kube-system", "get", "pod")
			filterPod := exec.Command("grep", "kubefay-agent")
			fmt.Print("hello")
			fmt.Printf(PipeCmd(kubectl, filterPod))
			// err := kubectl.Run()
			// So(err, ShouldBeNil)
			// out, err := kubectl.CombinedOutput()
			// So(err, ShouldBeNil)
			// So(out, ShouldContainSubstring, "Running")

		})
	})
}

func TestPipe(t *testing.T) {
	ls := exec.Command("ls")
	grep := exec.Command("grep", "test")
	out,err := PipeCmd(ls, grep)
	fmt.Println("=========")
	fmt.Println(err)
	fmt.Println(out)
	fmt.Println("=========")
}
