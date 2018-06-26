package main

import (
	"bitbucket.org/stack-rox/apollo/generated/api/v1"
	"bitbucket.org/stack-rox/apollo/pkg/central"
	"github.com/spf13/cobra"
)

func dockerBasedOrchestrator(shortName, longName string, cluster v1.ClusterType) *cobra.Command {
	swarmConfig := new(central.SwarmConfig)

	c := orchestratorCommand(shortName, longName, cluster)
	c.PersistentPreRun = func(*cobra.Command, []string) {
		cfg.SwarmConfig = swarmConfig
		cfg.ClusterType = cluster
	}
	c.RunE = func(*cobra.Command, []string) error {
		if err := validateConfig(cfg); err != nil {
			return err
		}
		return outputZip(cfg)
	}
	c.AddCommand(externalVolume(cluster))
	c.AddCommand(hostPathVolume(cluster))

	// Adds swarm specific flags
	c.PersistentFlags().StringVarP(&swarmConfig.ClairifyImage, "clairify-image", "", "stackrox.io/"+clairifyImage, "Clairify image to use")
	c.PersistentFlags().StringVarP(&swarmConfig.PreventImage, "prevent-image", "i", "stackrox.io/"+preventImage, "Prevent image to use")
	c.PersistentFlags().StringVarP(&swarmConfig.NetworkMode, "mode", "m", "ingress", "network mode to use (ingress or host)")
	c.PersistentFlags().IntVarP(&swarmConfig.PublicPort, "port", "p", 443, "public port to expose")
	return c
}
