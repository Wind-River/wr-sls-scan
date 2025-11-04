package cmd

import (
	"github.com/spf13/cobra"
)

const (
	cyclonedxExportExample = `  {{.appName}} cve cyclonedxExport -p <ProjectID> -o "path/out/filename.json"     export CycloneDX SBOM and VEX Report`
)

var CyclonedxExportCmd = &cobra.Command{
	Use:     "cyclonedxExport",
	Aliases: []string{"e", "exp", "out"},
	Short:   "Export the specified project CycloneDX SBOM and VEX Report",
	Long:    `Export the specified project CycloneDX SBOM and VEX Report, The exported file format is in JSON format.`,
	Example: formatExample(cyclonedxExportExample),
	Run: func(cmd *cobra.Command, args []string) {
		executeCVESubCommand("cyclonedxExport")
	},
}

func init() {
	CyclonedxExportCmd.Flags().Uint64VarP(&ProjectID, "projectId", "p", 0, "Project ID")
	CyclonedxExportCmd.Flags().StringVarP(&CVEFilePath, "outFile", "o", "", "Out File Name")

	CyclonedxExportCmd.MarkFlagRequired("projectId")
	CyclonedxExportCmd.MarkFlagRequired("outFile")

	cveCmd.AddCommand(CyclonedxExportCmd)
}
