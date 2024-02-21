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
	groupUpdateExample = `  {{.appName}} group update   -g <GroupID> -n "New Group Name"  -d "New Description"     update group`
)

var groupUpdateCmd = &cobra.Command{
	Use:              "update",
	Aliases:          []string{"u", "upd", "edit"},
	Short:            "Update the specified group",
	Long:             `Update the specified group that the current user is a member of.`,
	Example:          formatExample(groupUpdateExample),
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		executeGroupSubCommand("update")
	},
}

func init() {
	groupUpdateCmd.Flags().Int64VarP(&GroupID, "groupId", "g", -1, "Group ID")
	groupUpdateCmd.Flags().StringVarP(&GroupName, "name", "n", "", "Group Name")
	groupUpdateCmd.Flags().StringVarP(&GroupDescription, "description", "d", "", "Group Description")
	groupUpdateCmd.MarkFlagRequired("groupId")
	groupCmd.AddCommand(groupUpdateCmd)
}
