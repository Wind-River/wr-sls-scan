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
	projectExportExample = `  {{.appName}} project export -p <ProjectID>     export SBOM file under one project`
)

var projectExportCmd = &cobra.Command{
	Use:              "export",
	Aliases:          []string{"e", "exp", "out"},
	Short:            "Export the specified project sbom file",
	Long:             `Export the SBOM file of the specified project. SBOM files are available in two formats: SPDX and CycloneDX. You can specify the file format using the --sbomFormat parameter.`,
	Example:          formatExample(projectExportExample),
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		executeProjectSubCommand("export")
	},
}

func init() {
	projectExportCmd.Flags().Uint64VarP(&ProjectID, "projectId", "p", 0, "Project ID")
	projectExportCmd.Flags().StringVarP(&ExportFilePath, "outFile", "o", "", "Out File Name")
	projectExportCmd.Flags().StringVarP(&SBOMFormat, "sbomFormat", "b", "SPDX", "Export SBOM File format, Two formats available: SPDX and CycloneDX")

	projectExportCmd.MarkFlagRequired("projectId")

	projectCmd.AddCommand(projectExportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

}
