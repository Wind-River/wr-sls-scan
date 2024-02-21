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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	ExportFilePath string
	PackageID      uint64
	PackageName    string
)

// packageCmd represents the package command
var packageCmd = &cobra.Command{
	Use:              "package",
	Aliases:          []string{"k", "pkg"},
	Short:            "Operation commands related to the project's package",
	Long:             `Operation commands related to the project's package, Subcommands include query, detail and export.`,
	Example:          formatExample(packageExample),
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand := "query"
		if len(args) > 0 {
			subCommand = strings.ToLower(args[0])
		}

		if subCommand != "query" && subCommand != "list" && subCommand != "detail" && subCommand != "export" {
			fmt.Printf("Unknown subcommand: %s", subCommand)
			return
		}

		executePackageSubCommand(subCommand)
	},
}

func executePackageSubCommand(subCommand string) {
	packageId := PackageID
	if packageId != 0 && subCommand == "detail" {
		detailPackageById(packageId)
		return
	}

	projectId := ProjectID
	for projectId <= 0 {
		fmt.Print("Please enter project ID:")
		fmt.Scanln(&projectId)
	}

	packageName := PackageName //*packageNameFlag

	packageNames := strings.Split(packageName, ",")
	if subCommand == "query" || subCommand == "list" {
		queryPackage(projectId, packageNames)
	} else if subCommand == "detail" {
		if packageNames != nil && len(packageNames) > 0 {
			detailPackageByName(projectId, packageNames[0])
		} else {
			fmt.Println("Missing necessary parameter, Please input Package ID or Package Name.")
			return
		}
	} else if subCommand == "export" {
		exportPackage(projectId, packageNames)
	} else {
		fmt.Printf("Unrecognized subcommand: %s", subCommand)
	}
}

func queryPackage(projectId uint64, packageNames []string) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	err1 := printCVEScanProcessBar(projectId)

	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}

	var queryBean PackageQueryBean
	queryBean.ProjectId = projectId
	queryBean.Packages = packageNames
	queryBean.Strict = false

	jsonData, err := json.Marshal(queryBean)
	if err != nil {
		fmt.Println("Fatal error: ", err.Error())
		return
	}

	fmt.Println("Request to query package...")

	httpURL := getServerUrl() + "/project/package/list"
	req, err := http.NewRequest(http.MethodPost, httpURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Fatal error:", err.Error())
		return
	}

	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Fatal error:", err.Error())
			return
		}

		outJsonData(content)
		if needFormatOut() {
			var tableDataInfo PackageTableDataInfo
			json.Unmarshal(content, &tableDataInfo)
			if tableDataInfo.Code != 200 {
				fmt.Println(tableDataInfo.Msg)
				return
			}

			if tableDataInfo.Total > 0 {
				fmt.Println("Package List:")
				formatOutPackages(tableDataInfo)
			} else {
				fmt.Printf("No package %s under this project.", PackageName)
			}
		}
	} else {
		fmt.Println("Query package response status:", resp.Status)
	}
}

func detailPackageById(packageId uint64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	fmt.Println("Request the package details...")

	httpURL := getServerUrl() + "/manifest/" + strconv.FormatUint(packageId, 10)

	req, err := http.NewRequest(http.MethodGet, httpURL, nil)
	if err != nil {
		fmt.Println("Fatal error:", err.Error())
		return
	}

	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Fatal error:", err.Error())
			return
		}

		var result InfoManifestDetailResult
		json.Unmarshal(content, &result)
		if result.Code != 200 {
			fmt.Println(result.Msg)
			return
		}

		err1 := printCVEScanProcessBar(result.Data.ProjectId)

		if err1 != nil {
			fmt.Println(err1.Error())
			return
		}

		outJsonData(content)
		if needFormatOut() {
			formatOutManifestDetail(result.Data)
		}

		var queryBean CVEQueryBean
		queryBean.ProjectId = result.Data.ProjectId
		queryBean.ManifestId = packageId

		jsonData, err := json.Marshal(queryBean)
		if err != nil {
			fmt.Println("Fatal error: ", err.Error())
			return
		}

		httpURL := getServerUrl() + "/cve/cve/list"
		req, err := http.NewRequest(http.MethodPost, httpURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Fatal error:", err.Error())
			return
		}

		req.Header.Set("Cookie", cookieStr)
		req.Header.Set("content-type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)

		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Fatal error:", err.Error())
				return
			}

			outJsonData(content)
			if needFormatOut() {
				var tableDataInfo CVETableDataInfo
				json.Unmarshal(content, &tableDataInfo)
				if tableDataInfo.Code != 200 {
					fmt.Println(tableDataInfo.Msg)
					return
				}
				if tableDataInfo.Total > 0 {
					fmt.Println("Package CVE List:")
					formatOutCVEs(tableDataInfo)
				} else {
					fmt.Println("No CVE under this package.")
				}
			}
		} else {
			fmt.Println("Query CVE response status:", resp.Status)
		}
	} else {
		fmt.Println("Query CVE response status:", resp.Status)
	}
}

func detailPackageByName(projectId uint64, packageName string) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	err1 := printCVEScanProcessBar(projectId)

	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}

	var queryBean PackageQueryBean
	queryBean.ProjectId = projectId
	var packageNames = []string{packageName}
	queryBean.Packages = packageNames
	queryBean.Strict = false

	jsonData, err := json.Marshal(queryBean)
	if err != nil {
		fmt.Println("Fatal error: ", err.Error())
		return
	}

	fmt.Println("Request the package '" + packageName + "' details...")

	httpURL := getServerUrl() + "/project/package/list"
	req, err := http.NewRequest(http.MethodPost, httpURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Fatal error:", err.Error())
		return
	}

	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Fatal error:", err.Error())
			return
		}

		var tableDataInfo PackageTableDataInfo
		json.Unmarshal(content, &tableDataInfo)
		if tableDataInfo.Code != 200 {
			fmt.Println(tableDataInfo.Msg)
			return
		}

		outJsonData(content)
		if needFormatOut() {
			formatOutPackageDetail(tableDataInfo)
		}

		var queryBean CVEQueryBean
		queryBean.ProjectId = projectId
		var packageNames = []string{packageName}
		queryBean.Packages = packageNames

		jsonData, err := json.Marshal(queryBean)
		if err != nil {
			fmt.Println("Fatal error: ", err.Error())
			return
		}

		httpURL := getServerUrl() + "/cve/cve/list"
		req, err := http.NewRequest(http.MethodPost, httpURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Fatal error:", err.Error())
			return
		}

		req.Header.Set("Cookie", cookieStr)
		req.Header.Set("content-type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)

		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Fatal error:", err.Error())
				return
			}

			outJsonData(content)
			if needFormatOut() {
				var tableDataInfo CVETableDataInfo
				json.Unmarshal(content, &tableDataInfo)
				if tableDataInfo.Code != 200 {
					fmt.Println(tableDataInfo.Msg)
					return
				}

				if tableDataInfo.Total > 0 {
					fmt.Println("Package CVE List:")
					formatOutCVEs(tableDataInfo)
				} else {
					fmt.Println("No CVE under this package.")
				}
			}

		} else {
			fmt.Println("Query CVE response status:", resp.Status)
		}
	} else {
		fmt.Println("Query package response status:", resp.Status)
	}
}

func exportPackage(projectId uint64, packageNames []string) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	err1 := printCVEScanProcessBar(projectId)
	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}

	fmt.Println("Request to export package list...")

	infoProject, err := getInfoProject(projectId)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fileType := ".XLSX"
	fileSuffix := ".xlsx"

	var queryBean PackageQueryBean
	queryBean.ProjectId = projectId
	queryBean.Packages = packageNames
	queryBean.SuffixType = fileType
	queryBean.Strict = false

	//queryBean.FuzzyQuery = *fuzzyQueryFlag
	jsonData, err := json.Marshal(queryBean)
	if err != nil {
		fmt.Println("Fatal error:", err.Error())
		return
	}

	httpURL := getServerUrl() + "/project/package/export"
	req, err := http.NewRequest(http.MethodPost, httpURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Fatal error:", err.Error())
		return
	}

	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode == 200 {
		defer resp.Body.Close()

		outFileName := ExportFilePath //*outFileNameFlag
		if outFileName == "" {
			outFileName = infoProject.ProjectName + "-Package-List" + fileSuffix
		}

		outFile, err := os.Create(outFileName)
		if err != nil {
			panic(err)
		}
		defer outFile.Close()

		size, err := io.Copy(outFile, resp.Body)
		if err != nil {
			panic(err)
		}

		fmt.Printf("The out file is %s, file size is %v bytes.\n", outFileName, size)
	} else {
		fmt.Println("Export package response status:", resp.Status)
	}
}

func formatOutManifestDetail(detailData InfoManifest) {
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("%12v: %-48d %15s: %-64s\n", "Package ID", detailData.ManifestId, "License", detailData.LicenseName)
	fmt.Printf("%12v: %-48s %15s: %-64s\n", "Package Name", detailData.PackageName, "Package Version", detailData.PackageVersion)
	fmt.Printf("%12v: %-142s\n", "Url", detailData.Homepage)
	fmt.Printf("%12v: %-142s\n", "Summary", detailData.Summary)
	fmt.Printf("%12v: %-142s\n", "Description", detailData.Descript)
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")

	fmt.Println(" ")
}

func formatOutPackageDetail(tableDataInfo PackageTableDataInfo) {
	if tableDataInfo.Total > 0 {
		if len(tableDataInfo.Rows) > 0 {
			for _, infoPackage := range tableDataInfo.Rows {
				fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
				fmt.Printf("%12v: %-48s %15s: %-64s\n", "Package Name", infoPackage.PackageName, "Package Version", infoPackage.PackageVersion)
				fmt.Printf("%12v: %-142s\n", "Description", infoPackage.Descript)
				fmt.Printf("%12v: %-142s\n", "License", infoPackage.LicenseName)
				fmt.Printf("%12v: %-142s\n", "Summary", infoPackage.Summary)
				fmt.Printf("%12v: %-142s\n", "Url", infoPackage.Homepage)
				fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
			}
			fmt.Println(" ")
		}
	}
}

func formatOutPackages(tableDataInfo PackageTableDataInfo) {
	if tableDataInfo.Total > 0 {
		length := len(tableDataInfo.Rows)
		if length > 0 {
			fmt.Println("┌────────────┬──────────────────────────┬────────────────┬──────────────┬────────────────┬────────────────┬──────────────┬─────────────────┬───────────────┐")
			fmt.Println("│ PACKAGE ID │           PACKAGE        │ PACKAGE VERSION│ PACKAGE GROUP│     LICENSEs   │ VULNERABLE CVEs│ RESOLVED CVEs│ ALLOWLISTED CVEs│ PACKAGE STATUS│")
			fmt.Println("├────────────┼──────────────────────────┼────────────────┼──────────────┼────────────────┼────────────────┼──────────────┼─────────────────┼───────────────┤")
			var line, nameLines, versionLines, licenseLines int
			for index, infoPackage := range tableDataInfo.Rows {
				nameLines = getLineCount(len(infoPackage.PackageName), 25)
				versionLines = getLineCount(len(infoPackage.PackageVersion), 15)
				licenseLines = getLineCount(len(infoPackage.LicenseName), 15)
				line = max(nameLines, versionLines)
				line = max(line, licenseLines)
				if line > 1 {
					for i := 0; i < line; i++ {
						fmt.Printf("│ %-11s│ %-25s│ %-15s│ %-13s│ %-15s│ %-15s│ %-13s│ %-16s│ %-14s│\n", substring(strconv.FormatUint(infoPackage.ManifestId, 10), i*11, (i+1)*11), substring(infoPackage.PackageName, i*25, (i+1)*25), substring(infoPackage.PackageVersion, i*15, (i+1)*15), substring(infoPackage.PackageGroup, i*13, (i+1)*13), substring(infoPackage.LicenseName, i*15, (i+1)*15), substring(strconv.Itoa(infoPackage.Unresolved), i*15, (i+1)*15), substring(strconv.Itoa(infoPackage.Resolved), i*13, (i+1)*13), substring(strconv.Itoa(infoPackage.Allowlistcves), i*16, (i+1)*16), substring(infoPackage.PackageStatus, i*14, (i+1)*14))
					}
				} else {
					fmt.Printf("│ %-11d│ %-25s│ %-15s│ %-13s│ %-15s│ %-15d│ %-13d│ %-16d│ %-14s│\n", infoPackage.ManifestId, abbreviate(infoPackage.PackageName, 25), abbreviate(infoPackage.PackageVersion, 15), abbreviate(infoPackage.PackageGroup, 13), abbreviate(infoPackage.LicenseName, 15), infoPackage.Unresolved, infoPackage.Resolved, infoPackage.Allowlistcves, infoPackage.PackageStatus)
				}

				if index < length-1 {
					fmt.Println("├────────────┼──────────────────────────┼────────────────┼──────────────┼────────────────┼────────────────┼──────────────┼─────────────────┼───────────────┤")
				} else {

					fmt.Println("└────────────┴──────────────────────────┴────────────────┴──────────────┴────────────────┴────────────────┴──────────────┴─────────────────┴───────────────┘")
				}
			}
		}
		fmt.Println(" ")
	}
}

type PackageQueryBean struct {
	ProjectId        uint64   `json:"projectId"`
	Packages         []string `json:"packages"`
	PackageType      []string `json:"packageType"`
	PackageStatus    []string `json:"packageStatus"`
	License          []string `json:"license"`
	PackageGroups    []string `json:"packageGroups"`
	PackageVersion   string   `json:"packageVersion"`
	StarFlag         string   `json:"starFlag"`
	SuffixType       string   `json:"suffixType"`
	Orderby          string   `json:"orderby"`
	Sort             string   `json:"sort"`
	ScanResponseCode string   `json:"scanResponseCode"`
	Strict           bool     `json:"strict"`
	Confirm          bool     `json:"confirm"`
}

type InfoManifestDetailResult struct {
	Code  int          `json:"code"`
	Msg   string       `json:"msg"`
	Total int          `json:"total"`
	Data  InfoManifest `json:"data"`
}

type InfoManifest struct {
	ManifestId          uint64    `json:"manifestId"`
	ProjectId           uint64    `json:"projectId"`
	FileId              uint64    `json:"fileId"`
	PackageName         string    `json:"packageName"`
	PackageNameAlias    string    `json:"packageNameAlias"`
	PackageVersion      string    `json:"packageVersion"`
	PackageGroup        string    `json:"packageGroup"`
	PackageType         string    `json:"packageType"`
	PackageStatus       string    `json:"packageStatus"`
	LicenseName         string    `json:"licenseName"`
	Homepage            string    `json:"homepage"`
	Summary             string    `json:"summary"`
	Descript            string    `json:"descript"`
	apiUri              string    `json:"apiUri"`
	reportMd5           string    `json:"reportMd5"`
	licenseName         string    `json:"licenseName"`
	sourceInfo          string    `json:"sourceInfo"`
	updateTime          time.Time `json:"updateTime"`
	updateFlag          string    `json:"updateFlag"`
	starFlag            string    `json:"starFlag"`
	validFlag           string    `json:"validFlag"`
	StarFlag            string    `json:"starFlag"`
	ScanResponseMessage string    `json:"scanResponseMessage"`
	ScanResponseCode    string    `json:"scanResponseCode"`
	Confirm             bool      `json:"confirm"`
}

type PackageTableDataInfo struct {
	Code  int           `json:"code"`
	Msg   string        `json:"msg"`
	Total int           `json:"total"`
	Rows  []InfoPackage `json:"rows"`
}
type InfoPackage struct {
	ManifestId          uint64                 `json:"manifestId"`
	ProjectId           uint64                 `json:"projectId"`
	FileId              uint64                 `json:"fileId"`
	PackageName         string                 `json:"packageName"`
	PackageNameAlias    string                 `json:"packageNameAlias"`
	PackageVersion      string                 `json:"packageVersion"`
	PackageGroup        string                 `json:"packageGroup"`
	PackageType         string                 `json:"packageType"`
	PackageStatus       string                 `json:"packageStatus"`
	LicenseName         string                 `json:"licenseName"`
	Homepage            string                 `json:"homepage"`
	Summary             string                 `json:"summary"`
	Descript            string                 `json:"descript"`
	Unresolved          int                    `json:"unresolved"`
	Resolved            int                    `json:"resolved"`
	Allowlistcves       int                    `json:"allowlistcves"`
	UnresolvedCritical  int                    `json:"unresolvedCritical"`
	UnresolvedHigh      int                    `json:"unresolvedHigh"`
	UnresolvedMedium    int                    `json:"unresolvedMedium"`
	UnresolvedLow       int                    `json:"unresolvedLow"`
	UnresolvedNone      int                    `json:"unresolvedNone"`
	SecurityRisk        string                 `json:"securityRisk"`
	StarFlag            string                 `json:"starFlag"`
	ScanResponseMessage string                 `json:"scanResponseMessage"`
	ScanResponseCode    string                 `json:"scanResponseCode"`
	Confirm             bool                   `json:"confirm"`
	UnresolvedStat      InfoCVEStatsBySeverity `json:"unresolvedStat"`
}

type InfoCVEStatsBySeverity struct {
	Critical uint64 `json:"critical"`
	High     uint64 `json:"high"`
	Medium   uint64 `json:"medium"`
	Low      uint64 `json:"low"`
	None     uint64 `json:"none"`
}

//var FileType string

func init() {

	rootCmd.AddCommand(packageCmd)
	//packageCmd.Flags().StringVarP(&ExportFileName, "outFile", "o", "", "Out File Name")
	//packageCmd.Flags().StringVarP(&FileType, "fileType", "t", "Excel", "Export File Type, Three types available: Excel, SPDX, and CycloneDX")
	//pflag.Parse()
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// packageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// packageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
