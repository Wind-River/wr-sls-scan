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
	projectListExample = `  {{.appName}} project list        list all the projects`
)

// projectCmd represents the project command
var projectListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls", "query", "qry"},
	Short:   "List or query projects",
	Long:    `List or query projects that the current user has participated in.`,
	Example: formatExample(projectListExample),
	Run: func(cmd *cobra.Command, args []string) {
		executeProjectSubCommand("list")
	},
}

var age string

func init() {
	projectListCmd.Flags().Uint64VarP(&ProjectID, "projectId", "p", 0, "Project ID")
	projectListCmd.Flags().Int64VarP(&GroupID, "groupId", "g", -1, "Group ID")
	projectListCmd.Flags().StringVarP(&ProjectName, "name", "n", "", "Project Name")
	//projectCmd.Flags().StringVarP(&ProjectName, "projectName", "n", "", "Project Name")
	//projectCmd.Flags().StringVarP(&SBOMFile, "SBOMFile", "f", "", "SBOM file to be uploaded")
	projectCmd.AddCommand(projectListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// projectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//projectCmd.Flags().BoolP("projectID", "id", false, "project ID")
}
