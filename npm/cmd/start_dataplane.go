package main

import (
	"context"
	"fmt"

	"github.com/Azure/azure-container-networking/npm"
	npmconfig "github.com/Azure/azure-container-networking/npm/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

func newStartNPMDataplaneCmd() *cobra.Command {
	startNPMDataplaneCmd := &cobra.Command{
		Use:   "dataplane",
		Short: "Starts the Azure NPM dataplane process",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &npmconfig.Config{}
			err := viper.Unmarshal(config)
			if err != nil {
				return fmt.Errorf("failed to load config with error: %w", err)
			}

			return startDataplane(*config)
		},
	}

	return startNPMDataplaneCmd
}

func startDataplane(config npmconfig.Config) error {
	klog.Infof("loaded config: %+v", config)
	klog.Infof("Start NPM version: %s", version)

	err := initLogging()
	if err != nil {
		klog.Errorf("failed to init logging : %v", err)
		return err
	}

	n, err := npm.NewNetworkPolicyDataplane(context.Background(), config)
	if err != nil {
		klog.Errorf("failed to create dataplane : %v", err)
		return err
	}

	return n.Start(config, wait.NeverStop)
}
