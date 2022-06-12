/*
 * @Author: TicktW wxjpython@gmail.com
 * @Description: MIT License Copyright (C) 2022 TicktW@https://github.com/TicktW/kubefay
 */
// MIT License Copyright (C) 2022 TicktW@https://github.com/TicktW/kubefay

package main

import (
	"flag"
	"log"
	"time"

	crdinformers "github.com/TicktW/kubefay/pkg/client/informers/externalversions"
	"github.com/TicktW/kubefay/pkg/utils/k8s"
	"github.com/TicktW/kubefay/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/client-go/informers"
	componentbaseconfig "k8s.io/component-base/config"
	klog "k8s.io/klog/v2"
)

func main() {

	cmd := newAgentCommand()
	cmd.Execute()

}

func newAgentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "kubefay-agent",
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(); err != nil {
				klog.Fatalf("Error running agent: %v", err)
			}
		},
		Version: version.GetFullVersionWithRuntimeInfo(),
	}

	flags := cmd.Flags()
	klog.InitFlags(nil)
	// Install log flags
	flags.AddGoFlagSet(flag.CommandLine)
	return cmd
}

const informerDefaultResync = 12 * time.Hour

func run() error {
	klog.Infof("kubefay agent (version %s)", version.GetFullVersion())
	// logs.GlogSetter("5")

	// klog.V(1).Infof("kubefay agent 1 (version %s)", version.GetFullVersion())
	// klog.V(2).Infof("kubefay agent 2 (version %s)", version.GetFullVersion())
	// klog.V(3).Infof("kubefay agent 3 (version %s)", version.GetFullVersion())
	// klog.V(4).Infof("kubefay agent 4 (version %s)", version.GetFullVersion())
	// klog.V(5).Infof("kubefay agent 5 (version %s)", version.GetFullVersion())
	k8sConfig := componentbaseconfig.ClientConnectionConfiguration{}
	k8sClient, crdClient, err := k8s.CreateClients(k8sConfig)

	if err != nil {
		log.Fatalf("error to create k8s client: %v", err)
	}

	informerFactory := informers.NewSharedInformerFactory(k8sClient, informerDefaultResync)
	crdInformerFactory := crdinformers.NewSharedInformerFactory(crdClient, informerDefaultResync)

	subnetInformer := crdInformerFactory.Ipam().V1alpha1().SubNets()
	namespaceInformer := informerFactory.Core().V1().Namespaces()
	podInformer := informerFactory.Core().V1().Pods()

	return nil
}
