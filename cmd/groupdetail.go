/*
Copyright © 2024 Wind River Systems, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/spf13/cobra"
)

const (
	groupDetailExample = `  {{.appName}} group detail -g <GroupID>     view details of one group`
)

var groupDetailCmd = &cobra.Command{
	Use:     "detail",
	Aliases: []string{"d", "dtl", "view"},
	Short:   "View the specified group details",
	Long:    `View detailed information of the specified group that the current user is a member of.`,
	Example: formatExample(groupDetailExample),
	Run: func(cmd *cobra.Command, args []string) {
		executeGroupSubCommand("detail")
	},
}

func init() {
	groupDetailCmd.Flags().Int64VarP(&GroupID, "groupId", "g", -1, "Group ID")
	groupDetailCmd.MarkFlagRequired("groupId")
	groupCmd.AddCommand(groupDetailCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
