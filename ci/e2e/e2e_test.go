package e2e

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/TicktW/kubefay/pkg/utils/cmd"
)

func TestE2E(t *testing.T) {
	// Only pass t into top-level Convey calls
	Convey("Kubefay Network E2E Test", t, func() {

		FAY_INTERNET := "qq.com"
		if env := os.Getenv("FAY_INTERNET"); env != "" {
			FAY_INTERNET = env
		}

		Convey("Agent shoudld be in running state", func() {
			res, err := cmd.PipeCmdStr("kubectl -n kube-system get pod | grep kubefay-agent | grep Running | wc -l")
			So(err, ShouldBeNil)
			runAgentCnt, err := strconv.Atoi(res)
			So(err, ShouldBeNil)
			So(runAgentCnt, ShouldEqual, 2)

		})

		Convey("Test pods should be in running state", func() {
			res, err := cmd.PipeCmdStr("kubectl get pod | grep test-master |grep Running | wc -l")
			So(err, ShouldBeNil)
			masterPodCnt, err := strconv.Atoi(res)
			So(err, ShouldBeNil)
			So(masterPodCnt, ShouldEqual, 2)

			res, err = cmd.PipeCmdStr("kubectl get pod | grep test-master |grep Running | wc -l")
			So(err, ShouldBeNil)
			workerPodCnt, err := strconv.Atoi(res)
			So(err, ShouldBeNil)
			So(workerPodCnt, ShouldEqual, 2)
		})

		Convey("Test pods should ping internet successfully", func() {

			res, err := cmd.PipeCmdStr(`kubectl get pod | grep test | awk '{print $1}' | xargs`)
			So(err, ShouldBeNil)

			testPods := strings.Split(res, " ")

			for _, testPod := range testPods {
				kubeCmd := fmt.Sprintf("kubectl exec -it %s -- ping -c1 %s", testPod, FAY_INTERNET)
				res, err = cmd.PipeCmdStr(kubeCmd)
				So(err, ShouldBeNil)
				So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
			}
		})

		Convey("Test pods on master node should ping the other successfully", func() {

			res, err := cmd.PipeCmdStr(`kubectl get pod | grep test-master | awk '{print $1}' | xargs`)
			So(err, ShouldBeNil)

			testPods := strings.Split(res, " ")

			res, err = cmd.PipeCmdStr(`kubectl get pod -o wide | grep test-master | awk '{print $6}' | xargs`)
			So(err, ShouldBeNil)
			testIPs := strings.Split(res, " ")

			for idx, testPod := range testPods {
				kubeCmd := fmt.Sprintf("kubectl exec -it %s -- ping -c1 %s", testPod, testIPs[len(testIPs)-idx-1])
				res, err = cmd.PipeCmdStr(kubeCmd)
				So(err, ShouldBeNil)
				So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
			}
		})

		Convey("Test pods on worker node should ping the other successfully", func() {

			wokerNodeCmd := `kubectl get pod | grep test-worker | awk '{print $1}' | xargs`
			res, err := cmd.PipeCmdStr(wokerNodeCmd)
			SoMsg(wokerNodeCmd, err, ShouldBeNil)

			testPods := strings.Split(res, " ")

			res, err = cmd.PipeCmdStr(`kubectl get pod -o wide | grep test-master | awk '{print $6}' | xargs`)
			So(err, ShouldBeNil)
			testIPs := strings.Split(res, " ")

			for idx, testPod := range testPods {
				kubeCmd := fmt.Sprintf("kubectl exec -it %s -- ping -c1 %s", testPod, testIPs[len(testIPs)-idx-1])
				res, err = cmd.PipeCmdStr(kubeCmd)
				So(err, ShouldBeNil)
				So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
			}
		})

		Convey("Test pods on master node should ping pods on worker node successfully", func() {

			masterNodeCmd := `kubectl get pod | grep test-master | awk '{print $1}' | xargs`
			res, err := cmd.PipeCmdStr(masterNodeCmd)
			SoMsg(masterNodeCmd, err, ShouldBeNil)

			testPods := strings.Split(res, " ")

			res, err = cmd.PipeCmdStr(`kubectl get pod -o wide | grep test-worker | awk '{print $6}' | xargs`)
			So(err, ShouldBeNil)
			workerIPs := strings.Split(res, " ")

			for _, testPod := range testPods {
				for _, workerIP := range workerIPs {
					kubeCmd := fmt.Sprintf("kubectl exec -it %s -- ping -c1 %s", testPod, workerIP)
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
				}
			}
		})

		Convey("Test pods on worker node should ping pods on master node successfully", func() {

			wokerNodeCmd := `kubectl get pod | grep test-worker | awk '{print $1}' | xargs`
			res, err := cmd.PipeCmdStr(wokerNodeCmd)
			SoMsg(wokerNodeCmd, err, ShouldBeNil)

			testPods := strings.Split(res, " ")

			res, err = cmd.PipeCmdStr(`kubectl get pod -o wide | grep test-master | awk '{print $6}' | xargs`)
			So(err, ShouldBeNil)
			masterIPs := strings.Split(res, " ")

			for _, testPod := range testPods {
				for _, masterIP := range masterIPs {
					kubeCmd := fmt.Sprintf("kubectl exec -it %s -- ping -c1 %s", testPod, masterIP)
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
				}
			}
		})
	})
}
