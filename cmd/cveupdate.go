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
	cveUpdateStatusExample = `  {{.appName}} cve update -p <ProjectID> -c <cveId> -a <newStatus> -N <packageName>    	Update the status of CVE.`
)

var CVEStatusUpdateCmd = &cobra.Command{
	Use:              "update",
	Aliases:          []string{"u", "upd", "edit"},
	Short:            "Update cve status",
	Long:             `To update the specified cve status that the current user is participating in.`,
	Example:          formatExample(cveUpdateStatusExample),
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		executeCVESubCommand("update")
	},
}

func init() {
	CVEStatusUpdateCmd.Flags().Uint64VarP(&ProjectID, "projectId", "p", 0, "Project ID")
	CVEStatusUpdateCmd.Flags().StringVarP(&CVEID, "cveId", "c", "", "CVE ID")
	CVEStatusUpdateCmd.Flags().StringVarP(&newStatus, "newStatus", "a", "", "New Status: [Vulnerable, Needs Review, Won't fix, Resolved, Unaffected, Allowlisted, Ignored]")
	CVEStatusUpdateCmd.Flags().StringVarP(&packageName, "packageName", "N", "", "When the CVE belongs to multiple packages, it is necessary. Format: packageName packageVersion. Example:expat 2.2.6")

	CVEStatusUpdateCmd.MarkFlagRequired("projectId")
	CVEStatusUpdateCmd.MarkFlagRequired("cveId")
	CVEStatusUpdateCmd.MarkFlagRequired("newStatus")

	cveCmd.AddCommand(CVEStatusUpdateCmd)
}
