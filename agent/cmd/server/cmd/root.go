package cmd

import (
	"github.com/masx200/multinodewatchpanel/agent/server"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "1panel-agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		server.Start()
		return nil
	},
}
