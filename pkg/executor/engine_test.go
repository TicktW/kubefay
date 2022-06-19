package executor

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRunCmd(t *testing.T) {
	Convey("Execute command", t, func() {
		Convey("Execute error command", func() {
			res, err := RunCmd("l")
			So(res, ShouldBeBlank)
			So(err, ShouldBeError)
		})

		Convey("Execute right command", func() {
			res, err := RunCmd("ls engine.go")
			So(res, ShouldEqual, "engine.go\n")
			So(err, ShouldBeNil)
		})

		Convey("Execute empty command", func() {
			res, err := RunCmd("")
			So(res, ShouldBeBlank)
			So(err, ShouldBeError)
		})

		Convey("Run ovs", func() {
			res, _ := RunCmd("ovs-vsctl -h")
			fmt.Println(res)

		})
	})
	// res, err := RunCmd("l")
	// fmt.Println(res)
	// fmt.Print(err)
}
