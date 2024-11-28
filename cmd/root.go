package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "portscan",
	Short: "Use masscan or nmap to do the port scan, and return the result",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
