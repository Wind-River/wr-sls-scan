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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	scoreComparator    []string
	PublishedDateEnd   string
	PublishedDateBegin string
	FuzzyQuery         string
	CVEFilePath        string
	CVEID              string
	newStatus          string
	packageName        string
	Status             []string
	Severity           []string
	ModifiedDateBegin  string
	ModifiedDateEnd    string
)

// cveCmd represents the cve command
var cveCmd = &cobra.Command{
	Use:              "cve",
	Aliases:          []string{"c"},
	Short:            "Operation commands related to the project's CVE",
	Long:             `Operation commands related to the project's CVE, Subcommands include query, detail and export.`,
	Example:          formatExample(cveExample),
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand := "query"
		if len(args) > 0 {
			subCommand = strings.ToLower(args[0])
		}

		if subCommand != "query" && subCommand != "list" && subCommand != "detail" && subCommand != "export" && subCommand != "cyclonedxExport" {
			fmt.Printf("Unknown subcommand: %s", subCommand)
			return
		}

		executeCVESubCommand(subCommand)
	},
}

func executeCVESubCommand(subCommand string) {
	projectId := ProjectID
	for projectId <= 0 {
		fmt.Print("Please enter project ID:")
		fmt.Scanln(&projectId)
	}

	if subCommand == "query" || subCommand == "list" {
		queryCVE(projectId)
	} else if subCommand == "detail" {
		cveId := CVEID
		for cveId == "" {
			fmt.Print("Please enter CVE ID:")
			fmt.Scanln(&cveId)
		}

		detailCVE(projectId, cveId)
	} else if subCommand == "export" {
		exportCVE(projectId)
	} else if subCommand == "cyclonedxExport" {
		executeCyclonedxExport(projectId)
	} else if subCommand == "update" {
		cveId := CVEID
		for cveId == "" {
			fmt.Print("Please enter CVE ID:")
			fmt.Scanln(&cveId)
		}

		updateCveStatus(projectId, cveId)
	} else {
		fmt.Printf("Unrecognized subcommand: %s", subCommand)
	}
}

func updateCveStatus(projectId uint64, cveId string) {
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

	var updateBean CVEStatusUpdateBean
	updateBean.ProjectId = projectId
	updateBean.CveId = cveId
	updateBean.NewStatus = newStatus
	updateBean.PackageName = packageName
	//fmt.Printf("CVEStatusUpdateBean 对象内容：%+v\n", updateBean)
	jsonData, err := json.Marshal(updateBean)
	if err != nil {
		fmt.Println("Fatal error: ", err.Error())
		return
	}
	//fmt.Printf("JSON 数据: %s\n", string(jsonData))

	httpURL := getServerUrl() + "/cve/process/updateCveStatus"
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
			fmt.Println(err.Error())
			return
		}

		outJsonData(content)

		if needFormatOut() {
			var jaxResult AjaxResult
			json.Unmarshal(content, &jaxResult)
			if jaxResult.Code != 200 {
				fmt.Println(jaxResult.Msg)
				return
			} else {
				fmt.Println(jaxResult.Msg)
			}
		}
	} else {
		fmt.Println("Update the group response status:", resp.Status)
		return
	}
}

func queryCVE(projectId uint64) {
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

	var queryBean CVEQueryBean
	var cvssScore []ScoreComparator
	// the scoreComparator flag is of type []string, which is passed by multi-value command line
	// flag --scoreComparator, each flag value format: "Comparator Score", Comparator is string
	// type of supported value, Score is float64 type of cvss score.
	for _, scorecomparator := range scoreComparator {
		comparator := strings.Split(scorecomparator, " ")[0]
		score, err := strconv.ParseFloat(strings.Split(scorecomparator, " ")[1], 64)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		cvssScore = append(cvssScore, ScoreComparator{Comparator: comparator, Score: score})
	}
	queryBean.ProjectId = projectId
	queryBean.FuzzyQuery = FuzzyQuery
	queryBean.Status = Status
	queryBean.Severity = Severity
	queryBean.CvssScore = cvssScore
	queryBean.PublishedDateBegin = PublishedDateBegin
	queryBean.PublishedDateEnd = PublishedDateEnd
	queryBean.ModifiedDateBegin = ModifiedDateBegin
	queryBean.ModifiedDateEnd = ModifiedDateEnd
	//queryBean.Severity = *severityFlag
	jsonData, err := json.Marshal(queryBean)
	if err != nil {
		fmt.Println("Fatal error: ", err.Error())
		return
	}

	fmt.Println("Request to query CVE...")

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
				fmt.Println("CVE List:")
				formatOutCVEs(tableDataInfo)
			} else {
				fmt.Println("No CVEs under this project.")
			}
		}
	} else {
		fmt.Println("Query CVE response status:", resp.Status)
	}
}

func exportCVE(projectId uint64) {
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

	infoProject, err := getInfoProject(projectId)
	if err != nil {
		fmt.Println("Fatal error:", err.Error())
		return
	}

	var queryBean CVEQueryBean
	var cvssScore []ScoreComparator
	// the scoreComparator flag is of type []string, which is passed by multi-value command line
	// flag --scoreComparator, each flag value format: "Comparator Score", Comparator is string
	// type of supported value, Score is float64 type of cvss score.
	for _, scorecomparator := range scoreComparator {
		comparator := strings.Split(scorecomparator, " ")[0]
		score, err := strconv.ParseFloat(strings.Split(scorecomparator, " ")[1], 64)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		cvssScore = append(cvssScore, ScoreComparator{Comparator: comparator, Score: score})
	}
	queryBean.ProjectId = projectId
	queryBean.FuzzyQuery = FuzzyQuery
	queryBean.Status = Status
	queryBean.Severity = Severity
	queryBean.CvssScore = cvssScore
	queryBean.PublishedDateBegin = PublishedDateBegin
	queryBean.PublishedDateEnd = PublishedDateEnd
	queryBean.ModifiedDateBegin = ModifiedDateBegin
	queryBean.ModifiedDateEnd = ModifiedDateEnd

	jsonData, err := json.Marshal(queryBean)
	if err != nil {
		fmt.Println("Fatal error:", err.Error())
		return
	}

	fmt.Println("Request to export cve list...")

	httpURL := getServerUrl() + "/cve/cve/export"
	req, err := http.NewRequest(http.MethodPost, httpURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Fatal error:", err.Error())
		return
	}

	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Fatal error: HTTP request failed:", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		outFileName := CVEFilePath
		if outFileName == "" {
			outFileName = infoProject.ProjectName + "-CVE-List.xlsx"
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
		fmt.Println("Export CVE response status:", resp.Status)
	}
}

// 导出CycloneDX SBOM and VEX Report
func executeCyclonedxExport(projectId uint64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	//检查文件格式
	if CVEFilePath != "" {
		if !strings.HasSuffix(CVEFilePath, ".json") && !strings.HasSuffix(CVEFilePath, ".JSON") {
			fmt.Println("invalid file, must end with .json or .JSON")
			return
		}
	}

	err1 := printCVEScanProcessBar(projectId)
	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}

	infoProject, err := getInfoProject(projectId)
	if err != nil {
		fmt.Println("Fatal error:", err.Error())
		return
	}

	var queryBean CVECyclonedxBean
	queryBean.ProjectId = projectId
	queryBean.SuffixType = ".JSON"

	jsonData, err := json.Marshal(queryBean)
	if err != nil {
		fmt.Println("Fatal error:", err.Error())
		return
	}
	fmt.Println("Request to export CycloneDX SBOM and VEX Report...")

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
	if err != nil {
		fmt.Println("Fatal error: HTTP request failed:", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		outFileName := CVEFilePath
		if outFileName == "" {
			outFileName = infoProject.ProjectName + "-CycloneDX.json"
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
		fmt.Println("Export CycloneDX response status:", resp.Status)
	}
}

func detailCVE(projectId uint64, cveId string) {
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

	var queryBean CVEQueryBean
	queryBean.ProjectId = projectId
	queryBean.FuzzyQuery = cveId
	//queryBean.Severity = *severityFlag
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

		var tableDataInfo CVETableDataInfo
		json.Unmarshal(content, &tableDataInfo)
		if tableDataInfo.Code != 200 {
			fmt.Println(tableDataInfo.Msg)
			return
		}

		if tableDataInfo.Total > 0 {
			infoCVE := tableDataInfo.Rows[0]
			infoProject, err2 := getInfoProject(projectId)
			if err2 != nil {
				return
			}

			detailBean, err2 := getCVEDetail(infoCVE.Id, projectId)
			if needFormatOut() {
				formatOutCVE(infoProject.CvssScoreType, detailBean)
			}
		} else {
			fmt.Printf("No %s under this project.", cveId)
		}
	} else {
		fmt.Println("Query CVE response status:", resp.Status)
	}
}

func getCVEDetail(id uint64, projectId uint64) (CVEDetailBean, error) {
	var bean CVEDetailBean
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		return bean, tokenError
	}

	url := getServerUrl() + "/cve/cve/" + strconv.FormatUint(id, 10) + "?projectId=" + strconv.FormatUint(projectId, 10)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err.Error())
		return bean, err
	}

	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
			return bean, err
		}

		outJsonData(content)

		var detailDataInfo CVEDetailDataInfo
		json.Unmarshal(content, &detailDataInfo)
		if detailDataInfo.Code != 200 {
			fmt.Println("Response error:", detailDataInfo.Msg)
			return bean, errors.New(detailDataInfo.Msg)
		}

		bean = detailDataInfo.Data
		return bean, nil
	} else {
		fmt.Println("Get CVE detail response status:", resp.Status)

		return bean, fmt.Errorf("Get detail response status: %q", resp.Status)
	}
}

func formatOutCVE(cvssType string, bean CVEDetailBean) {
	if reflect.DeepEqual(bean, CVEDetailBean{}) {
		fmt.Println("The specified CVE does not exist.")
		return
	}

	solution := "[×]"
	if strings.EqualFold(bean.HasSolution, "yes") {
		solution = "[√]"
	}
	fmt.Println("Vulnerability Details:" + bean.CveId + "/")

	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("%26v: %-56s %28v\n", "url", "https://nvd.nist.gov/vuln/detail/"+bean.CveId, solution+"Solution ")
	fmt.Printf("%26v: %-22s %28v: %-24s\n", "Published", formatDateTime(bean.NvdPublishedDate), "Last Modified", formatDateTime(bean.NvdModifiedDate))

	if bean.CvssType == "CVSS 3" {
		fmt.Println("CVSS3.x Scores and Vulnerability Types──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
		fmt.Printf("%26v: %-22s %28v: %-48s\n", "CVSS SCORE", fmt.Sprintf("%2.1f", bean.CvssScore), "VECTOR", strings.ToUpper(bean.VectorString))

		fmt.Printf("EXPLOITABILITY %3v──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────\n", bean.ExploitabilityScore)
		fmt.Printf("%26v: %-22s %28v: %-22s %26v: %-22s\n", "ATTACK VECTOR (AV)", bean.AttackVector, "ATTACK COMPLEXITY (AC)", bean.AttackComplexity, "PRIVILEGES REQUIRED (PR)", bean.PrivilegesRequired)
		fmt.Printf("%26v: %-22s %28v: %-22s\n", "USER INTERACTION (UI)", bean.UserInteraction, "SCOPE(S)", bean.Scope)

		fmt.Printf("IMPACT %3v──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────\n", bean.ImpactScore)
		fmt.Printf("%26v: %-22s %28v: %-22s %26v: %-22s\n", "CONFIDENTIALITY IMPACT (C)", bean.ConfidentialityImpact, "INTEGRITY IMPACT (I)", bean.IntegrityImpact, "AVAILABILITY IMPACT (A)", bean.AvailabilityImpact)
	} else {
		fmt.Println("CVSS2.x Scores and Vulnerability Types──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
		fmt.Printf("%26v: %-22s %28v: %-48s\n", "CVSS SCORE", fmt.Sprintf("%2.1f", bean.CvssV2), "VECTOR", bean.VectorStringV2)

		fmt.Printf("EXPLOITABILITY %3v──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────\n", bean.ExploitabilitiyScoreV2)
		fmt.Printf("%26v: %-22s %28v: %-22s %26v: %-22s\n", "ACCESS VECTOR (AV)", bean.AccessVectorV2, "ACCESS COMPLEXITY (AC)", bean.AccessComplexityV2, "AUTHENTICATION (AU)", bean.AuthenticationV2)

		fmt.Printf("IMPACT %3v──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────\n", bean.ImpactScoreV2)
		fmt.Printf("%26v: %-22s %28v: %-22s %26v: %-22s\n", "CONFIDENTIALITY IMPACT (C)", bean.ConfidentialityImpactV2, "INTEGRITY IMPACT (I)", bean.IntegrityImpactV2, "AVAILABILITY IMPACT (A)", bean.AvailabilityImpactV2)
	}
	fmt.Println("WEAKNESS ENUMERATION────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("%26v: %-84s \n", "CWE ID", bean.CweId)
	if bean.JiraResolvedCve != nil && len(bean.JiraResolvedCve) > 0 {
		length := len(bean.JiraResolvedCve)
		fmt.Println("Wind River Solutions:")
		fmt.Println("┌────────────────────────────────────────┬───────────────────────────┐")
		fmt.Println("│      Operating System Distribution     │       Fixed Projects      │")
		fmt.Println("├────────────────────────────────────────┼───────────────────────────┤")
		for index, infoJiraSolution := range bean.JiraResolvedCve {
			fmt.Printf("│ %-39s│          %5d            │\n", infoJiraSolution.FullName, infoJiraSolution.FixProjects)
			if index < length-1 {
				fmt.Println("├────────────────────────────────────────┼───────────────────────────┤")
			} else {
				fmt.Println("└────────────────────────────────────────┴───────────────────────────┘")
			}
		}
	}

	var links []InfoLink
	json.Unmarshal([]byte(bean.Link), &links)

	if links != nil && len(links) > 0 {
		length := len(links)
		fmt.Println("References to Advisories, Solutions and Tools:")
		fmt.Println("┌────────────────────────────────────────────────────────────────────────────────────────┬─────────────────────────────────────────────────────────────────┐")
		fmt.Println("│                                        Hyperlink                                       │                              Resource                           │")
		fmt.Println("├────────────────────────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────┤")
		var line, linkLines, tagsLines int
		for index, infoLink := range links {
			tags := ""
			for index := 0; index < len(infoLink.Tags); index++ {
				tags += "[" + infoLink.Tags[index] + "] "
			}

			linkLines = getLineCount(len(infoLink.Url), 87)
			tagsLines = getLineCount(len(tags), 64)

			line = max(linkLines, tagsLines)

			if line > 1 {
				for i := 0; i < line; i++ {
					fmt.Printf("│ %-87s│ %-64s│\n", substring(infoLink.Url, i*87, (i+1)*87), substring(tags, i*64, (i+1)*64))
				}
			} else {
				fmt.Printf("│ %-87s│ %-64s│\n", abbreviate(infoLink.Url, 87), abbreviate(tags, 64))
			}

			if index < length-1 {
				fmt.Println("├────────────────────────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────┤")
			} else {
				fmt.Println("└────────────────────────────────────────────────────────────────────────────────────────┴─────────────────────────────────────────────────────────────────┘")
			}
		}
	}
}

func formatOutCVEs(tableDataInfo CVETableDataInfo) {
	if tableDataInfo.Total > 0 {
		length := len(tableDataInfo.Rows)
		if length > 0 {
			fmt.Println("┌────────────┬──────────────────┬────────────────────────────────────────────┬───────────┬────────────┬───────────────────┬───────────────────┬────────────┐")
			fmt.Println("│  SEVERITY  │      CVE ID      │                     PACKAGE                │ CVSS SCORE│ WR SOLUTION│      PUBLISHED    │      MODIFIED     │    STATUS  │")
			fmt.Println("├────────────┼──────────────────┼────────────────────────────────────────────┼───────────┼────────────┼───────────────────┼───────────────────┼────────────┤")
		} else {
			return
		}
		var line, cveIdLines, packageLines int
		for index, infoCVE := range tableDataInfo.Rows {
			cveIdLines = getLineCount(len(infoCVE.PackageStr), 44)
			packageLines = getLineCount(len(infoCVE.CveId), 17)
			line = max(cveIdLines, packageLines)

			if line > 1 {
				for i := 0; i < line; i++ {
					fmt.Printf("│ %-11s│ %-17s│ %-43s│ %-10s│ %-11s│ %-18s│ %-18s│ %-11s│\n", substring(infoCVE.Severity, i*11, (i+1)*11), substring(infoCVE.CveId, i*17, (i+1)*17), substring(infoCVE.PackageStr, i*43, (i+1)*43), substring(fmt.Sprintf(" %-10.1f", infoCVE.CvssScore), i*10, (i+1)*10), substring(infoCVE.HasSolution, i*11, (i+1)*11), substring(formatDateTime(infoCVE.NvdPublishedDate), i*18, (i+1)*18), substring(formatDateTime(infoCVE.NvdModifiedDate), i*18, (i+1)*18), substring(infoCVE.CveStatus, i*11, (i+1)*11))
				}
			} else {
				fmt.Printf("│ %-11s│ %-17s│ %-43s│ %-10.1f│ %-11s│ %-18s│ %-18s│ %-11s│\n", infoCVE.Severity, infoCVE.CveId, abbreviate(infoCVE.PackageStr, 43), infoCVE.CvssScore, infoCVE.HasSolution, formatDateTime(infoCVE.NvdPublishedDate), formatDateTime(infoCVE.NvdModifiedDate), infoCVE.CveStatus)
			}
			if index < length-1 {
				fmt.Println("├────────────┼──────────────────┼────────────────────────────────────────────┼───────────┼────────────┼───────────────────┼───────────────────┼────────────┤")
			} else {
				fmt.Println("└────────────┴──────────────────┴────────────────────────────────────────────┴───────────┴────────────┴───────────────────┴───────────────────┴────────────┘")
			}
		}

		fmt.Println(" ")
	}
}

type ScoreComparator struct {
	Comparator string  `json:"comparator"`
	Score      float64 `json:"score"`
}

type CVEQueryBean struct {
	ProjectId          uint64            `json:"projectId"`
	ManifestId         uint64            `json:"manifestId"`
	FuzzyQuery         string            `json:"fuzzyQuery"`
	Severity           []string          `json:"severity"`
	CvssScore          []ScoreComparator `json:"cvssScore"`
	Status             []string          `json:"status"`
	Packages           []string          `json:"packages"`
	Cveids             []string          `json:"cveids"`
	PackageGroups      []string          `json:"packageGroups"`
	PublishedDateBegin string            `json:"publishedDateBegin"`
	PublishedDateEnd   string            `json:"publishedDateEnd"`
	ModifiedDateBegin  string            `json:"modifiedDateBegin"`
	ModifiedDateEnd    string            `json:"modifiedDateEnd"`
	HasSolution        string            `json:"hasSolution"`
}

type CVECyclonedxBean struct {
	ProjectId  uint64 `json:"projectId"`
	SuffixType string `json:"suffixType"`
}

type CVEStatusUpdateBean struct {
	ProjectId   uint64 `json:"projectId"`
	CveId       string `json:"cveId"`
	PackageName string `json:"packageName"`
	NewStatus   string `json:"actionType"`
}

type CVETableDataInfo struct {
	Code  int       `json:"code"`
	Msg   string    `json:"msg"`
	Total int       `json:"total"`
	Rows  []InfoCVE `json:"rows"`
}

type InfoCVE struct {
	Id                        uint64  `json:"id"`
	ProjectCveId              uint64  `json:"projectCveId"`
	ProjectId                 uint64  `json:"projectId"`
	CveId                     string  `json:"cveId"`
	CvssType                  string  `json:"cvssType"`
	PackageStr                string  `json:"packageStr"`
	PackageName               string  `json:"packageName"`
	PackageVersion            string  `json:"packageVersion"`
	PackageType               string  `json:"packageType"`
	Severity                  string  `json:"severity"`
	CvssScore                 float64 `json:"cvssScore"`
	HasSolution               string  `json:"hasSolution"`
	CvssV2                    float64 `json:"cvssV2"`
	CvssV3                    float64 `json:"cvssV3"`
	SeverityV2                string  `json:"severityV2"`
	ImpactScoreV2             float64 `json:"impactScoreV2"`
	SeverityV3                string  `json:"severityV3"`
	ImpactScore               float64 `json:"impactScore"`
	NvdPublishedDate          string  `json:"nvdPublishedDate"`
	NvdModifiedDate           string  `json:"nvdModifiedDate"`
	CveStatus                 string  `json:"cveStatus"`
	PackageGroup              string  `json:"packageGroup"`
	AttackVector              string  `json:"attackVector"`
	AttackComplexity          string  `json:"attackComplexity"`
	PrivilegesRequired        string  `json:"privilegesRequired"`
	UserInteraction           string  `json:"userInteraction"`
	Scope                     string  `json:"scope"`
	ConfidentialityImpact     string  `json:"confidentialityImpact"`
	IntegrityImpact           string  `json:"integrityImpact"`
	AvailabilityImpact        string  `json:"availabilityImpact"`
	FirstSaveTime             string  `json:"firstSaveTime"`
	ConfirmFixTime            string  `json:"confirmFixTime"`
	ResolvedTime              string  `json:"resolvedTime"`
	AllowlistFlag             string  `json:"allowlistFlag"`
	IgnoredFlag               string  `json:"ignoredFlag"`
	StarFlag                  string  `json:"starFlag"`
	Link                      string  `json:"link"`
	Description               string  `json:"description"`
	PatchList                 string  `json:"patchList"`
	FixAvailable              string  `json:"fixAvailable"`
	VersionV2                 float64 `json:"versionV2"`
	AccessVectorV2            string  `json:"accessVectorV2"`
	VectorStringV2            string  `json:"vectorStringV2"`
	AuthenticationV2          string  `json:"authenticationV2"`
	IntegrityImpactV2         string  `json:"integrityImpactV2"`
	AccessComplexityV2        string  `json:"accessComplexityV2"`
	AvailabilityImpactV2      string  `json:"availabilityImpactV2"`
	ConfidentialityImpactV2   string  `json:"confidentialityImpactV2"`
	AcinsuInfoV2              string  `json:"acinsuInfoV2"`
	ObtainAllPrivilegeV2      string  `json:"obtainAllPrivilegeV2"`
	ExploitabilitiyScoreV2    float64 `json:"exploitabilitiyScoreV2"`
	ObtainUserPrivilegeV2     string  `json:"obtainUserPrivilegeV2"`
	ObtainOtherPrivilegeV2    string  `json:"obtainOtherPrivilegeV2"`
	UserInteractionRequiredV2 string  `json:"userInteractionRequiredV2"`
	VectorString              string  `json:"vectorString"`
	Version                   float64 `json:"version"`
	ExploitabilityScore       float64 `json:"exploitabilityScore"`
	SubVersion                string  `json:"subVersion"`
	AckDate                   string  `json:"ackDate"`
	CweId                     string  `json:"cweId"`
	FileId                    uint64  `json:"fileId"`
	ManifestId                uint64  `json:"manifestId"`
	PackageNum                int     `json:"packageNum"`
	JiraResolvedCve           string  `json:"jiraResolvedCve"`
}

type CVEDetailDataInfo struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data CVEDetailBean `json:"data"`
}

type CVEDetailBean struct {
	ID                        uint64             `json:"id"`
	ProjectCveId              uint64             `json:"projectCveId"`
	ProjectId                 uint64             `json:"projectId"`
	CveId                     string             `json:"cveId"`
	CvssType                  string             `json:"cvssType"`
	PackageStr                string             `json:"packageStr"`
	PackageName               string             `json:"packageName"`
	PackageNameAlias          string             `json:"packageNameAlias"`
	PackageVersion            string             `json:"packageVersion"`
	PackageType               string             `json:"packageType"`
	Severity                  string             `json:"severity"`
	CvssScore                 float64            `json:"cvssScore"`
	HasSolution               string             `json:"hasSolution"`
	CvssV2                    float64            `json:"cvssV2"`
	CvssV3                    float64            `json:"cvssV3"`
	SeverityV2                string             `json:"severityV2"`
	ImpactScoreV2             float64            `json:"impactScoreV2"`
	SeverityV3                string             `json:"severityV3"`
	ImpactScore               float64            `json:"impactScore"`
	NvdPublishedDate          string             `json:"nvdPublishedDate"`
	NvdModifiedDate           string             `json:"nvdModifiedDate"`
	CveStatus                 string             `json:"cveStatus"`
	PackageGroup              string             `json:"packageGroup"`
	AttackVector              string             `json:"attackVector"`
	AttackComplexity          string             `json:"attackComplexity"`
	PrivilegesRequired        string             `json:"privilegesRequired"`
	UserInteraction           string             `json:"userInteraction"`
	Scope                     string             `json:"scope"`
	ConfidentialityImpact     string             `json:"confidentialityImpact"`
	IntegrityImpact           string             `json:"integrityImpact"`
	AvailabilityImpact        string             `json:"availabilityImpact"`
	FirstSaveTime             time.Time          `json:"firstSaveTime"`
	ConfirmFixTime            time.Time          `json:"confirmFixTime"`
	ResolvedTime              time.Time          `json:"resolvedTime"`
	AllowlistFlag             string             `json:"allowlistFlag"`
	IgnoredFlag               string             `json:"ignoredFlag"`
	StarFlag                  string             `json:"starFlag"`
	Link                      string             `json:"link"`
	Description               string             `json:"description"`
	PatchList                 string             `json:"patchList"`
	FixAvailable              string             `json:"fixAvailable"`
	VersionV2                 float64            `json:"versionV2"`
	AccessVectorV2            string             `json:"accessVectorV2"`
	VectorStringV2            string             `json:"vectorStringV2"`
	AuthenticationV2          string             `json:"authenticationV2"`
	IntegrityImpactV2         string             `json:"integrityImpactV2"`
	AccessComplexityV2        string             `json:"accessComplexityV2"`
	AvailabilityImpactV2      string             `json:"availabilityImpactV2"`
	ConfidentialityImpactV2   string             `json:"confidentialityImpactV2"`
	AcinsuInfoV2              string             `json:"acinsuInfoV2"`
	ObtainAllPrivilegeV2      string             `json:"obtainAllPrivilegeV2"`
	ExploitabilitiyScoreV2    float64            `json:"exploitabilitiyScoreV2"`
	ObtainUserPrivilegeV2     string             `json:"obtainUserPrivilegeV2"`
	ObtainOtherPrivilegeV2    string             `json:"obtainOtherPrivilegeV2"`
	UserInteractionRequiredV2 string             `json:"userInteractionRequiredV2"`
	VectorString              string             `json:"vectorString"`
	Version                   float64            `json:"version"`
	ExploitabilityScore       float64            `json:"exploitabilityScore"`
	SubVersion                string             `json:"subVersion"`
	AckDate                   string             `json:"ackDate"`
	CweId                     string             `json:"cweId"`
	FileId                    uint64             `json:"fileId"`
	ManifestId                uint64             `json:"manifestId"`
	PackageNum                int                `json:"packageNum"`
	JiraResolvedCve           []InfoJiraSolution `json:"jiraResolvedCve"`
}

type InfoJiraSolution struct {
	ID          uint64 `json:"id"`
	CveId       string `json:"cveId"`
	ServiceType string `json:"serviceType"`
	DistroName  string `json:"distroName"`
	FullName    string `json:"fullName"`
	FixProjects int    `json:"fixProjects"`
}

type InfoLink struct {
	Url       string   `json:"url"`
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	Refsource string   `json:"refsource"`
}

func init() {
	//cveCmd.Flags().StringVarP(&ExportFileName, "outFile", "e", "", "Out File Name")
	rootCmd.AddCommand(cveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
