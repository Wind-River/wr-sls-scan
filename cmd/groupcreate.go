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
	groupCreateExample = `  {{.appName}} group create -n "Group Name" -d "Description of the Group"                           create new group`
)

var groupCreateCmd = &cobra.Command{
	Use:              "create",
	Aliases:          []string{"c", "creat", "new", "add"},
	Short:            "Create new group",
	Long:             `To create a new group, used for sharing projects among users.`,
	Example:          formatExample(groupCreateExample),
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		executeGroupSubCommand("create")
	},
}

func init() {
	groupCreateCmd.Flags().StringVarP(&GroupName, "name", "n", "", "Group Name")
	groupCreateCmd.Flags().StringVarP(&GroupDescription, "description", "d", "", "Group Description")
	groupCreateCmd.MarkFlagRequired("groupName")
	groupCmd.AddCommand(groupCreateCmd)
}
