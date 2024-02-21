/*
Copyright © 2024 Wind River Systems, Inc.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const versionTemplate = `  {{.appName}} {{.version}} \r\n  © Wind River Systems, Inc.`

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v", "ver"},
	Short:   "Provides the version information of " + APP_NAME,
	Long:    `Provides the version information of ` + APP_NAME + ` and quit.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Tprintf(versionTemplate, map[string]interface{}{
			"appName": APP_NAME,
			"version": VERSION,
		}))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
