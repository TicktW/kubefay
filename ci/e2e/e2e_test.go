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
	Convey("Kubefay E2E Test", t, func() {

		FAY_INTERNET := "qq.com"
		if env := os.Getenv("FAY_INTERNET"); env != "" {
			FAY_INTERNET = env
		}

		Convey("\nPods used shoud be in correct state", func() {

			Convey("Kube-system namespace:kubefay-agent shoudld be in running state", func() {
				res, err := cmd.PipeCmdStr("kubectl -n kube-system get pod | grep kubefay-agent | grep Running | wc -l")
				So(err, ShouldBeNil)
				runAgentCnt, err := strconv.Atoi(res)
				So(err, ShouldBeNil)
				So(runAgentCnt, ShouldEqual, 2)

			})

			Convey("Default namespace:test-master pods are running", func() {
				res, err := cmd.PipeCmdStr("kubectl get pod | grep test-master |grep Running | wc -l")
				So(err, ShouldBeNil)
				podCnt, err := strconv.Atoi(res)
				So(err, ShouldBeNil)
				So(podCnt, ShouldEqual, 2)
			})

			Convey("Default namespace:test-worker pod are running", func() {

				res, err := cmd.PipeCmdStr("kubectl get pod | grep test-worker |grep Running | wc -l")
				So(err, ShouldBeNil)
				podCnt, err := strconv.Atoi(res)
				So(err, ShouldBeNil)
				So(podCnt, ShouldEqual, 2)

			})

			Convey("Kubefay-test namespace: test-master pods are running", func() {

				res, err := cmd.PipeCmdStr("kubectl -n kubefay-test get pod | grep test-master |grep Running | wc -l")
				So(err, ShouldBeNil)
				podCnt, err := strconv.Atoi(res)
				So(err, ShouldBeNil)
				So(podCnt, ShouldEqual, 2)

			})

			Convey("Kubefay-test namespace: test-worker pods are running", func() {

				res, err := cmd.PipeCmdStr("kubectl -n kubefay-test get pod | grep test-worker |grep Running | wc -l")
				So(err, ShouldBeNil)
				podCnt, err := strconv.Atoi(res)
				So(err, ShouldBeNil)
				So(podCnt, ShouldEqual, 2)

			})

		})

		Convey("\nNetwork of all kinds of pods should be ready", func() {
			Convey("All namespace:test pods should ping internet successfully", func() {

				res, err := cmd.PipeCmdStr(`kubectl get pod | grep test | awk '{print $1}' | xargs`)
				So(err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				for _, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl exec -it %s -- ping -c1 %s", testPod, FAY_INTERNET)
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
				}

				res, err = cmd.PipeCmdStr(`kubectl -n kubefay-test get pod | grep test | awk '{print $1}' | xargs`)
				So(err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for _, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- ping -c1 %s", testPod, FAY_INTERNET)
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
				}
			})

			Convey("Test pods on master node should ping the others successfully", func() {

				podGetCmd := `kubectl get pod | grep test-master | awk '{print $1}' | xargs`
				res, err := cmd.PipeCmdStr(podGetCmd)
				SoMsg(podGetCmd, err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				res, err = cmd.PipeCmdStr(`kubectl get pod -A -o wide | grep test-master | awk '{print $7}' | xargs`)
				SoMsg(fmt.Sprintln(err), err, ShouldBeNil)
				testIPs := strings.Split(res, " ")
				for idx, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl exec -it %s -- ping -c1 %s", testPod, testIPs[len(testIPs)-idx-1])
					res, err = cmd.PipeCmdStr(kubeCmd)
					SoMsg(kubeCmd, err, ShouldBeNil)
					So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
				}

				podGetCmd = `kubectl -n kubefay-test get pod | grep test-master | awk '{print $1}' | xargs`
				res, err = cmd.PipeCmdStr(podGetCmd)
				SoMsg(podGetCmd, err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for idx, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- ping -c1 %s", testPod, testIPs[len(testIPs)-idx-1])
					res, err = cmd.PipeCmdStr(kubeCmd)
					SoMsg(kubeCmd, err, ShouldBeNil)
					So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
				}

			})

			Convey("Test pods on worker node should ping the others successfully", func() {

				wokerNodeCmd := `kubectl get pod | grep test-worker | awk '{print $1}' | xargs`
				res, err := cmd.PipeCmdStr(wokerNodeCmd)
				SoMsg(wokerNodeCmd, err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				res, err = cmd.PipeCmdStr(`kubectl get pod -A -o wide | grep test-master | awk '{print $7}' | xargs`)
				So(err, ShouldBeNil)
				testIPs := strings.Split(res, " ")

				for idx, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl exec -it %s -- ping -c1 %s", testPod, testIPs[len(testIPs)-idx-1])
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
				}

				wokerNodeCmd = `kubectl -n kubefay-test get pod | grep test-worker | awk '{print $1}' | xargs`
				res, err = cmd.PipeCmdStr(wokerNodeCmd)
				SoMsg(wokerNodeCmd, err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for idx, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- ping -c1 %s", testPod, testIPs[len(testIPs)-idx-1])
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

				res, err = cmd.PipeCmdStr(`kubectl get -A pod -o wide | grep test-worker | awk '{print $7}' | xargs`)
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

				masterNodeCmd = `kubectl -n kubefay-test get pod | grep test-master | awk '{print $1}' | xargs`
				res, err = cmd.PipeCmdStr(masterNodeCmd)
				SoMsg(masterNodeCmd, err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, workerIP := range workerIPs {
						kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- ping -c1 %s", testPod, workerIP)
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

				res, err = cmd.PipeCmdStr(`kubectl get pod -A -o wide | grep test-master | awk '{print $7}' | xargs`)
				So(err, ShouldBeNil)
				masterIPs := strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, masterIP := range masterIPs {
						kubeCmd := fmt.Sprintf("kubectl exec -it %s -- ping -c1 %s", testPod, masterIP)
						// fmt.Println(kubeCmd)
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
					}
				}

				wokerNodeCmd = `kubectl -n kubefay-test get pod | grep test-worker | awk '{print $1}' | xargs`
				res, err = cmd.PipeCmdStr(wokerNodeCmd)
				SoMsg(wokerNodeCmd, err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, masterIP := range masterIPs {
						kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- ping -c1 %s", testPod, masterIP)
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, "1 packets transmitted, 1 packets received, 0% packet loss")
					}
				}

			})

		})
	})
}
