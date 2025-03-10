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
	CVEQueryCmd.Flags().StringArrayVarP(&Severity, "severity", "v", []string{}, "CVE severity: [Critical,High,Medium,Low,None]")
	CVEQueryCmd.Flags().StringArrayVarP(&Status, "status", "u", []string{}, "CVE Status: [Allowlisted,Ignored,Needs Review,Resolved,Unaffected,Vulnerable,Won't fix]")
	CVEQueryCmd.Flags().StringVarP(&ModifiedDateEnd, "modifiedDateEnd", "D", "", "CVE modified time period query condition that end date. Format: yyyy-MM-dd")
	CVEQueryCmd.Flags().StringVarP(&ModifiedDateBegin, "modifiedDateBegin", "M", "", "CVE modified time period query condition that beginning date. Format: yyyy-MM-dd")
	CVEQueryCmd.Flags().StringArrayVarP(&scoreComparator, "scoreComparator", "s", []string{}, "CVSS score comparison condition, can be specify mutiple values. Format: \"Comparator Score\". Supported Comparator: NUMERIC_EQUAL, NUMERIC_GREATER_THAN, NUMERIC_LESS_THAN, NUMERIC_NOT_EQUAL, NUMERIC_GREATER_THAN_OR_EQUAL, NUMERIC_LESSER_THAN_OR_EQUAL. Example: --scoreComparator \"NUMERIC_LESS_THAN 8.0\" --scoreComparator \"NUMERIC_GREATER_THAN_OR_EQUAL 7.0\"")
	CVEQueryCmd.Flags().StringVarP(&PublishedDateEnd, "publishedDateEnd", "E", "", "CVE published time period query condition that end date. Format: yyyy-MM-dd")
	CVEQueryCmd.Flags().StringVarP(&PublishedDateBegin, "publishedDateBegin", "B", "", "CVE published time period query condition that beginning date. Format: yyyy-MM-dd")
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
