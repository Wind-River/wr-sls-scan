package cmd

import (
	"github.com/spf13/cobra"
)

const (
	projectUpdateSbomExample = `  {{.appName}} project updateSbom -p <ProjectID> -f "path/sbomfile"			update SBOM file for a project`
)

var projectUpdateSbomCmd = &cobra.Command{
	Use:              "updateSbom",
	Aliases:          []string{"updateSbom"},
	Short:            "update SBOM file for a project",
	Long:             `update SBOM file for a project to sls-scan cli tool commands.`,
	Example:          formatExample(projectUpdateSbomExample),
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		executeProjectSubCommand("updateSbom")
	},
}

func init() {
	projectUpdateSbomCmd.Flags().Uint64VarP(&ProjectID, "projectId", "p", 0, "Project ID")
	projectUpdateSbomCmd.Flags().StringVarP(&SBOMFile, "SBOMFile", "f", "", "SBOM file to be uploaded")

	projectUpdateSbomCmd.MarkFlagRequired("projectId")
	projectUpdateSbomCmd.MarkFlagRequired("SBOMFile")
	projectUpdateSbomCmd.MarkFlagsRequiredTogether("projectId", "SBOMFile")
	projectCmd.AddCommand(projectUpdateSbomCmd)
}
