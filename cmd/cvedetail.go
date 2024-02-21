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
	cveDetailExample = `  {{.appName}} cve detail -p <ProjectID> -c <CVE-ID>      view cve detail`
)

var CVEDetailCmd = &cobra.Command{
	Use:     "detail",
	Aliases: []string{"d", "dtl", "view"},
	Short:   "View the specified CVE details",
	Long:    `View the details of the specified CVE of project that the current user is participating in.`,
	Example: formatExample(cveDetailExample),
	Run: func(cmd *cobra.Command, args []string) {
		executeCVESubCommand("detail")
	},
}

func init() {
	CVEDetailCmd.Flags().StringVarP(&CVEID, "cveId", "c", "", "CVE ID")
	CVEDetailCmd.Flags().Uint64VarP(&ProjectID, "projectId", "p", 0, "Project ID")

	CVEDetailCmd.MarkFlagRequired("cveId")
	CVEDetailCmd.MarkFlagRequired("projectId")

	cveCmd.AddCommand(CVEDetailCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// projectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//projectCmd.Flags().BoolP("projectID", "id", false, "project ID")
}
