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
	cveQueryExample = `  {{.appName}} cve query -p <ProjectID> -z "<FuzzyQuery>"                                 query CVE thru. fuzzy query conditions under one project`
)

var CVEQueryCmd = &cobra.Command{
	Use:     "query",
	Aliases: []string{"q", "qry", "list", "ls", "l"},
	Short:   "Query the specified project CVEs",
	Long:    `Query the CVE of the specified project that the current user has participated in.`,
	Example: formatExample(cveQueryExample),
	Run: func(cmd *cobra.Command, args []string) {
		executeCVESubCommand("query")
	},
}

func init() {
	CVEQueryCmd.Flags().StringVarP(&FuzzyQuery, "fuzzyQuery", "z", "", "Fuzzy query conditions, including CVE, Package and Package Group")
	CVEQueryCmd.Flags().Uint64VarP(&ProjectID, "projectId", "p", 0, "Project ID")

	CVEQueryCmd.MarkFlagRequired("projectId")
	//CVEQueryCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	cveCmd.AddCommand(CVEQueryCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

}
