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
	cveExportExample = `  {{.appName}} cve export -p <ProjectID> -z "<FuzzyQuery>"  -o "path/out/filename.xlsx"     export cve information thru. fuzzy query conditions under one project`
)

var CVEExportCmd = &cobra.Command{
	Use:     "export",
	Aliases: []string{"e", "exp", "out"},
	Short:   "Export the specified project CVEs",
	Long:    `Export the specified project CVEs, The exported file format is in Excel format.`,
	Example: formatExample(cveExportExample),
	Run: func(cmd *cobra.Command, args []string) {
		executeCVESubCommand("export")
	},
}

func init() {
	CVEExportCmd.Flags().StringVarP(&FuzzyQuery, "fuzzyQuery", "z", "", "Fuzzy Query")
	CVEExportCmd.Flags().StringVarP(&CVEFilePath, "outFile", "o", "", "Out File Name")
	CVEExportCmd.Flags().Uint64VarP(&ProjectID, "projectId", "p", 0, "Project ID")
	CVEExportCmd.MarkFlagRequired("projectId")

	cveCmd.AddCommand(CVEExportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

}
