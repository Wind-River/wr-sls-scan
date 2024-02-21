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
	projectCreateExample = `  {{.appName}} project create -n "Project Name" -f "path/sbomfile"                        create a new project
  {{.appName}} project create -n "Project Name" -f "path/sbomfile" -g <GroupID>           create a new project under one group`
)

var projectCreateCmd = &cobra.Command{
	Use:              "create",
	Aliases:          []string{"c", "creat", "new", "add"},
	Short:            "Create new project",
	Long:             `To create a new project, an SBOM file needs to be uploaded and can be assigned to a specified group.`,
	Example:          formatExample(projectCreateExample),
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		executeProjectSubCommand("create")
	},
}

func init() {
	projectCreateCmd.Flags().StringVarP(&ProjectName, "name", "n", "", "Project Name")
	projectCreateCmd.Flags().StringVarP(&SBOMFile, "SBOMFile", "f", "", "SBOM file to be uploaded")
	projectCreateCmd.Flags().Int64VarP(&GroupID, "groupId", "g", -1, "Group ID")
	projectCreateCmd.Flags().StringVarP(&Description, "description", "d", "", "Description")
	projectCreateCmd.MarkFlagRequired("name")
	projectCreateCmd.MarkFlagRequired("SBOMFile")
	projectCreateCmd.MarkFlagsRequiredTogether("name", "SBOMFile")
	projectCmd.AddCommand(projectCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// projectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//projectCmd.Flags().BoolP("projectID", "id", false, "project ID")
}
