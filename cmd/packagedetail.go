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
	packageDetailExample = `  {{.appName}} package detail -p <ProjectID> -m <PakcageID>           view all CVEs under cetain package (thru. package id) under one project
  {{.appName}} package detail -p <ProjectID> -n "<PakcageName>"       view all CVEs under cetain package (thru. package name) under one project`
)

var packageDetailCmd = &cobra.Command{
	Use:     "detail",
	Aliases: []string{"d", "dtl", "view"},
	Short:   "View the specified package details",
	Long:    `View the details of the specified package of project that the current user is participating in.`,
	Example: formatExample(packageDetailExample),
	Run: func(cmd *cobra.Command, args []string) {
		executePackageSubCommand("detail")
	},
}

func init() {
	packageDetailCmd.Flags().Uint64VarP(&ProjectID, "projectId", "p", 0, "Project ID")
	packageDetailCmd.Flags().StringVarP(&PackageName, "name", "n", "", "Package Name")
	packageDetailCmd.Flags().Uint64VarP(&PackageID, "packageId", "m", 0, "Package ID")
	packageDetailCmd.MarkFlagRequired("projectId")

	packageCmd.AddCommand(packageDetailCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// projectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//projectCmd.Flags().BoolP("projectID", "id", false, "project ID")
}
