package e2e

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/kubefay/kubefay/pkg/utils/cmd"
)

func TestE2E(t *testing.T) {
	// Only pass t into top-level Convey calls
	Convey("Kubefay E2E Test", t, func() {

		FAY_INTERNET := "qq.com"
		FAY_MASTER_NODE := "kind-control-plane"
		FAY_WORKER_NODE := "kind-control-plane"

		if internet := os.Getenv("FAY_INTERNET"); internet != "" {
			FAY_INTERNET = internet
		}

		if master := os.Getenv("FAY_MASTER_NODE"); master != "" {
			FAY_MASTER_NODE = master
		}

		if master := os.Getenv("FAY_WORKER_NODE"); master != "" {
			FAY_WORKER_NODE = master
		}

		PING_RESULT_OK := "1 packets transmitted, 1 packets received, 0% packet loss"
		WEB_RESULT_OK := "kubefay"

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
				So(podCnt, ShouldEqual, 3)

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

		Convey("\nLayer 2 network of all kinds of pods should be ready", func() {
			Convey("All namespace:test pods should ping internet successfully", func() {

				res, err := cmd.PipeCmdStr(`kubectl get pod | grep test | awk '{print $1}' | xargs`)
				So(err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				for _, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl exec -it %s -- ping -c1 %s", testPod, FAY_INTERNET)
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, PING_RESULT_OK)
				}

				res, err = cmd.PipeCmdStr(`kubectl -n kubefay-test get pod | grep test | awk '{print $1}' | xargs`)
				So(err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for _, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- ping -c1 %s", testPod, FAY_INTERNET)
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, PING_RESULT_OK)
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
					So(res, ShouldContainSubstring, PING_RESULT_OK)
				}

				podGetCmd = `kubectl -n kubefay-test get pod | grep test-master | awk '{print $1}' | xargs`
				res, err = cmd.PipeCmdStr(podGetCmd)
				SoMsg(podGetCmd, err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for idx, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- ping -c1 %s", testPod, testIPs[len(testIPs)-idx-1])
					res, err = cmd.PipeCmdStr(kubeCmd)
					SoMsg(kubeCmd, err, ShouldBeNil)
					So(res, ShouldContainSubstring, PING_RESULT_OK)
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
					So(res, ShouldContainSubstring, PING_RESULT_OK)
				}

				wokerNodeCmd = `kubectl -n kubefay-test get pod | grep test-worker | awk '{print $1}' | xargs`
				res, err = cmd.PipeCmdStr(wokerNodeCmd)
				SoMsg(wokerNodeCmd, err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for idx, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- ping -c1 %s", testPod, testIPs[len(testIPs)-idx-1])
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, PING_RESULT_OK)
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
						So(res, ShouldContainSubstring, PING_RESULT_OK)
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
						So(res, ShouldContainSubstring, PING_RESULT_OK)
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
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, PING_RESULT_OK)
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
						So(res, ShouldContainSubstring, PING_RESULT_OK)
					}
				}

			})

			Convey("Master host should ping pods successfully", func() {

				masterNodeCmd := fmt.Sprintf(
					`kubectl -n kube-system get pod -o wide | grep %s | grep kubefay-agent | awk '{print $1}' | xargs echo -n`,
					FAY_MASTER_NODE,
				)

				res, err := cmd.PipeCmdStr(masterNodeCmd)
				SoMsg(masterNodeCmd, err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				res, err = cmd.PipeCmdStr(`kubectl get pod -A -o wide | grep test- | awk '{print $7}' | xargs`)
				So(err, ShouldBeNil)
				masterIPs := strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, masterIP := range masterIPs {
						kubeCmd := fmt.Sprintf("kubectl -n kube-system exec -it %s -- ping -c1 %s", testPod, masterIP)
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, PING_RESULT_OK)
					}
				}

			})

			Convey("Worker host should ping pods successfully", func() {

				workerNodeCmd := fmt.Sprintf(
					`kubectl -n kube-system get pod -o wide | grep %s | grep kubefay-agent | awk '{print $1}' | xargs echo -n`,
					FAY_WORKER_NODE,
				)

				res, err := cmd.PipeCmdStr(workerNodeCmd)
				SoMsg(workerNodeCmd, err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				res, err = cmd.PipeCmdStr(`kubectl get pod -A -o wide | grep test- | awk '{print $7}' | xargs`)
				So(err, ShouldBeNil)
				masterIPs := strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, masterIP := range masterIPs {
						kubeCmd := fmt.Sprintf("kubectl -n kube-system exec -it %s -- ping -c1 %s", testPod, masterIP)
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, PING_RESULT_OK)
					}
				}

			})
		})

		Convey("\nLayer 3,4 network of all kinds of pods should be ready", func() {

			Convey("Test pods on master node should get the others' web service successfully", func() {

				podGetCmd := `kubectl get pod | grep test-master | awk '{print $1}' | xargs`
				res, err := cmd.PipeCmdStr(podGetCmd)
				SoMsg(podGetCmd, err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				res, err = cmd.PipeCmdStr(`kubectl get pod -A -o wide | grep test-master | awk '{print $7}' | xargs`)
				SoMsg(fmt.Sprintln(err), err, ShouldBeNil)
				testIPs := strings.Split(res, " ")
				for idx, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl exec -it %s -- kubefay-test-cli %s", testPod, testIPs[len(testIPs)-idx-1])
					res, err = cmd.PipeCmdStr(kubeCmd)
					SoMsg(kubeCmd, err, ShouldBeNil)
					So(res, ShouldContainSubstring, WEB_RESULT_OK)
				}

				podGetCmd = `kubectl -n kubefay-test get pod | grep test-master | awk '{print $1}' | xargs`
				res, err = cmd.PipeCmdStr(podGetCmd)
				SoMsg(podGetCmd, err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for idx, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- kubefay-test-cli %s", testPod, testIPs[len(testIPs)-idx-1])
					res, err = cmd.PipeCmdStr(kubeCmd)
					SoMsg(kubeCmd, err, ShouldBeNil)
					So(res, ShouldContainSubstring, WEB_RESULT_OK)
				}

			})

			Convey("Test pods on worker node should get the others' web service successfully", func() {

				wokerNodeCmd := `kubectl get pod | grep test-worker | awk '{print $1}' | xargs`
				res, err := cmd.PipeCmdStr(wokerNodeCmd)
				SoMsg(wokerNodeCmd, err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				res, err = cmd.PipeCmdStr(`kubectl get pod -A -o wide | grep test-master | awk '{print $7}' | xargs`)
				So(err, ShouldBeNil)
				testIPs := strings.Split(res, " ")

				for idx, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl exec -it %s -- kubefay-test-cli %s", testPod, testIPs[len(testIPs)-idx-1])
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, WEB_RESULT_OK)
				}

				wokerNodeCmd = `kubectl -n kubefay-test get pod | grep test-worker | awk '{print $1}' | xargs`
				res, err = cmd.PipeCmdStr(wokerNodeCmd)
				SoMsg(wokerNodeCmd, err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for idx, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- kubefay-test-cli %s", testPod, testIPs[len(testIPs)-idx-1])
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, WEB_RESULT_OK)
				}
			})

			Convey("Test pods on master node should get web service of pods on worker node successfully", func() {

				masterNodeCmd := `kubectl get pod | grep test-master | awk '{print $1}' | xargs`
				res, err := cmd.PipeCmdStr(masterNodeCmd)
				SoMsg(masterNodeCmd, err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				res, err = cmd.PipeCmdStr(`kubectl get -A pod -o wide | grep test-worker | awk '{print $7}' | xargs`)
				So(err, ShouldBeNil)
				workerIPs := strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, workerIP := range workerIPs {
						kubeCmd := fmt.Sprintf("kubectl exec -it %s -- kubefay-test-cli %s", testPod, workerIP)
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, WEB_RESULT_OK)
					}
				}

				masterNodeCmd = `kubectl -n kubefay-test get pod | grep test-master | awk '{print $1}' | xargs`
				res, err = cmd.PipeCmdStr(masterNodeCmd)
				SoMsg(masterNodeCmd, err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, workerIP := range workerIPs {
						kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- kubefay-test-cli %s", testPod, workerIP)
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, WEB_RESULT_OK)
					}
				}
			})

			Convey("Test pods on worker node should get web service of pods on master node successfully", func() {

				wokerNodeCmd := `kubectl get pod | grep test-worker | awk '{print $1}' | xargs`
				res, err := cmd.PipeCmdStr(wokerNodeCmd)
				SoMsg(wokerNodeCmd, err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				res, err = cmd.PipeCmdStr(`kubectl get pod -A -o wide | grep test-master | awk '{print $7}' | xargs`)
				So(err, ShouldBeNil)
				masterIPs := strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, masterIP := range masterIPs {
						kubeCmd := fmt.Sprintf("kubectl exec -it %s -- kubefay-test-cli %s", testPod, masterIP)
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, WEB_RESULT_OK)
					}
				}

				wokerNodeCmd = `kubectl -n kubefay-test get pod | grep test-worker | awk '{print $1}' | xargs`
				res, err = cmd.PipeCmdStr(wokerNodeCmd)
				SoMsg(wokerNodeCmd, err, ShouldBeNil)

				testPods = strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, masterIP := range masterIPs {
						kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- kubefay-test-cli %s", testPod, masterIP)
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, WEB_RESULT_OK)
					}
				}

			})

			Convey("Master host should get web service of pods successfully", func() {

				masterNodeCmd := fmt.Sprintf(
					`kubectl -n kube-system get pod -o wide | grep %s | grep kubefay-agent | awk '{print $1}' | xargs echo -n`,
					FAY_MASTER_NODE,
				)

				res, err := cmd.PipeCmdStr(masterNodeCmd)
				SoMsg(masterNodeCmd, err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				res, err = cmd.PipeCmdStr(`kubectl get pod -A -o wide | grep test- | awk '{print $7}' | xargs`)
				So(err, ShouldBeNil)
				masterIPs := strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, masterIP := range masterIPs {
						kubeCmd := fmt.Sprintf("kubectl -n kube-system exec -it %s -- kubefay-test-cli %s", testPod, masterIP)
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, WEB_RESULT_OK)
					}
				}

			})

			Convey("Worker host should get web service of pods successfully", func() {

				workerNodeCmd := fmt.Sprintf(
					`kubectl -n kube-system get pod -o wide | grep %s | grep kubefay-agent | awk '{print $1}' | xargs echo -n`,
					FAY_WORKER_NODE,
				)

				res, err := cmd.PipeCmdStr(workerNodeCmd)
				SoMsg(workerNodeCmd, err, ShouldBeNil)

				testPods := strings.Split(res, " ")

				res, err = cmd.PipeCmdStr(`kubectl get pod -A -o wide | grep test- | awk '{print $7}' | xargs`)
				So(err, ShouldBeNil)
				masterIPs := strings.Split(res, " ")

				for _, testPod := range testPods {
					for _, masterIP := range masterIPs {
						kubeCmd := fmt.Sprintf("kubectl -n kube-system exec -it %s -- kubefay-test-cli %s", testPod, masterIP)
						res, err = cmd.PipeCmdStr(kubeCmd)
						So(err, ShouldBeNil)
						So(res, ShouldContainSubstring, WEB_RESULT_OK)
					}
				}

			})

			Convey("Hosts should get web service of cluster and nodeport service successfully", func() {

				workerNodeCmd := `kubectl -n kube-system get pod -o wide | grep kubefay-agent | awk '{print $1}' | xargs echo -n`

				res, err := cmd.PipeCmdStr(workerNodeCmd)
				SoMsg(workerNodeCmd, err, ShouldBeNil)

				testNodes := strings.Split(res, " ")

				clusterServiceIP, err := cmd.PipeCmdStr(`kubectl get svc|grep test-service | awk '{print $3}'`)
				So(err, ShouldBeNil)

				clusterIPCmd := `kubectl get svc|grep test-service | awk '{print $5}' | awk -F '/' '{print $1}' | awk -F ':' '{print $1}'`
				clusterIPPort, err := cmd.PipeCmdStr(clusterIPCmd)
				So(err, ShouldBeNil)

				nodePortCmd := `kubectl get svc|grep test-service | awk '{print $5}' | awk -F '/' '{print $1}' | awk -F ':' '{print $2}'`
				nodePort, err := cmd.PipeCmdStr(nodePortCmd)
				So(err, ShouldBeNil)

				for _, testPod := range testNodes {
					kubeCmd := fmt.Sprintf("kubectl -n kube-system exec -it %s -- kubefay-test-cli %s:%s", testPod, clusterServiceIP, clusterIPPort)
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, WEB_RESULT_OK)

					kubeCmd = fmt.Sprintf("kubectl -n kube-system exec -it %s -- kubefay-test-cli %s:%s", testPod, "127.0.0.1", nodePort)
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, WEB_RESULT_OK)
				}

			})

			Convey("All pods should get cluster service successfully", func() {

				res, err := cmd.PipeCmdStr(`kubectl get pod | grep test- | awk '{print $1}' | xargs echo -n`)
				So(err, ShouldBeNil)
				testPods := strings.Split(res, " ")

				clusterServiceIP, err := cmd.PipeCmdStr(`kubectl get svc|grep test-service | awk '{print $3}'`)
				So(err, ShouldBeNil)

				clusterIPCmd := `kubectl get svc|grep test-service | awk '{print $5}' | awk -F '/' '{print $1}' | awk -F ':' '{print $1}'`
				clusterIPPort, err := cmd.PipeCmdStr(clusterIPCmd)
				So(err, ShouldBeNil)

				for _, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl exec -it %s -- kubefay-test-cli %s:%s", testPod, clusterServiceIP, clusterIPPort)
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, WEB_RESULT_OK)
				}

				res, err = cmd.PipeCmdStr(`kubectl get pod -n kubefay-test | grep test- | awk '{print $1}' | xargs echo -n`)
				So(err, ShouldBeNil)
				testPods = strings.Split(res, " ")

				for _, testPod := range testPods {
					kubeCmd := fmt.Sprintf("kubectl -n kubefay-test exec -it %s -- kubefay-test-cli %s:%s", testPod, clusterServiceIP, clusterIPPort)
					res, err = cmd.PipeCmdStr(kubeCmd)
					So(err, ShouldBeNil)
					So(res, ShouldContainSubstring, WEB_RESULT_OK)
				}
			})
		})
	})
}
