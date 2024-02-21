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
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	ProjectID   uint64
	ProjectName string
	SBOMFile    string
	Description string
	SBOMFormat  string
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:              "project",
	Aliases:          []string{"p", "proj"},
	Short:            "Operation commands related to the project",
	Long:             `Operation commands related to the project, Subcommands include list, detail, create, update, delete, rescan, cancel and export.`,
	Example:          formatExample(projectExample),
	TraverseChildren: true,
	SilenceUsage:     false,
	SilenceErrors:    false,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand := "list"
		if len(args) > 0 {
			subCommand = strings.ToLower(args[0])
		}

		if subCommand != "list" && subCommand != "create" && subCommand != "detail" && subCommand != "rescan" && subCommand != "scan" && subCommand != "cancel" && subCommand != "delete" {
			fmt.Printf("Unknown subcommand: %s", subCommand)
			return
		}

		executeProjectSubCommand(subCommand)
	},
}

func executeProjectSubCommand(subCommand string) {
	if subCommand == "list" {
		listProject()
		return
	} else if subCommand == "create" {
		createProject()
		return
	}

	projectId := ProjectID
	for projectId <= 0 {
		fmt.Print("Please enter project ID:")
		fmt.Scanln(&projectId)
	}

	if subCommand == "detail" {
		projectDetail(projectId)
	} else if subCommand == "rescan" || subCommand == "scan" {
		scanProject(projectId)
	} else if subCommand == "cancel" {
		cancelProject(projectId)
	} else if subCommand == "delete" {
		deleteProject(projectId)
	} else if subCommand == "update" {
		updateProject(projectId)
	} else if subCommand == "export" {
		exportSbom(projectId)
	} else {
		fmt.Printf("Unrecognized subcommand: %s", subCommand)
	}
}

func exportSbom(projectId uint64) {
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

	fmt.Println("Request to export SBOM file ...")

	infoProject, err := getInfoProject(projectId)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fileFormat := SBOMFormat
	fileFormat = strings.Trim(fileFormat, " ")
	fileFormat = strings.ToUpper(fileFormat)
	fileSuffix := ".spdx.json"
	if strings.Index(fileFormat, "SP") >= 0 {
		fileFormat = ".SPDX.JSON"
		fileSuffix = ".spdx.json"
	} else {
		fileFormat = ".JSON"
		fileSuffix = "-CycloneDX.json"
	}

	var queryBean PackageQueryBean
	queryBean.ProjectId = projectId
	queryBean.SuffixType = fileFormat
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
		fmt.Println("Export SBOM file response status:", resp.Status)
	}
}

func listProject() {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	var queryBean InfoProject
	if ProjectID != 0 {
		queryBean.ProjectId = ProjectID
	}

	queryBean.ProjectName = ProjectName
	if GroupID < 0 {
		queryBean.GroupId = -1
	} else {
		queryBean.GroupId = GroupID
	}

	jsonData, err := json.Marshal(queryBean)
	if err != nil {
		fmt.Println("Fatal error: ", err.Error())
		return
	}

	url := getServerUrl() + "/project/list"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err.Error())
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
			var tableDataInfo ProjectTableDataInfo
			json.Unmarshal(content, &tableDataInfo)

			if tableDataInfo.Code != 200 {
				fmt.Println(tableDataInfo.Msg)
				return
			}

			if tableDataInfo.Total > 0 {
				fmt.Println("Project List:")
				formatOutProjects(tableDataInfo.Rows)
			} else {
				fmt.Println("No projects.")
			}
		}

	} else {
		fmt.Println("List project response status: ", resp.Status)
	}
}

func projectDetail(projectId uint64) {
	err1 := printCVEScanProcessBar(projectId)

	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}

	infoProject, err := getInfoProject(projectId)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if reflect.DeepEqual(infoProject, InfoProject{}) {
		fmt.Println("The project does not exist or has no permissions.")
		return
	} else {
		if needFormatOut() {
			formatOutProject(infoProject)
		}

		getLastestFile(projectId)
		manifestSummary(projectId)
	}
}

func createProject() {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	groupId := GroupID
	if groupId < 0 {
		groupId = 0
	}

	projectName := strings.TrimSpace(ProjectName) //*projectNameFlag
	for projectName == "" {
		fmt.Print("Please enter project name:")
		fmt.Scanln(&projectName)
		projectName = strings.TrimSpace(projectName)
	}

	description := Description
	sbomFile := SBOMFile //*sbomFileFlag
	sbomFile = strings.TrimSpace(sbomFile)
	for sbomFile == "" {
		fmt.Print("Please enter SBOM file path:")
		fmt.Scanln(&sbomFile)
	}

	isExsted := false
	var err error
	for !isExsted {
		isExsted, err = pathExists(sbomFile)
		if err != nil {
			fmt.Printf("%s is invalid file.\n", sbomFile)
			fmt.Print("Please enter valid SBOM file path:")
			fmt.Scanln(&sbomFile)
			isExsted = false
			continue
		} else {
			isExsted = true
		}
	}

	_, fileName := filepath.Split(sbomFile)
	// 打开文件
	f, err := os.Open(sbomFile)
	if err != nil {
		fmt.Printf("Open the file error: " + err.Error())
		return
	}
	defer f.Close()
	// create buffer
	var buffer bytes.Buffer
	//Implemented multipart parsing of MIME through package multipart
	writer := multipart.NewWriter(&buffer)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		panic(err)
	}
	// Copy the file content to the new FormFile
	_, err = io.Copy(part, f)
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	httpURL := getServerUrl() + "/project/upload/"
	req, err := http.NewRequest("POST", httpURL, &buffer)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("upload error")
		panic(err)
	} else {
		uploadContent, err := io.ReadAll(resp.Body)
		if err != nil {
			outJsonData(uploadContent)

			if needFormatOut() {
				fmt.Println(err.Error())
			}

			return
		}

		var uploadResult UploadResult
		json.Unmarshal(uploadContent, &uploadResult)

		var createProjectParam CreateProjectParam
		createProjectParam.ProjectName = projectName
		createProjectParam.ManifestFile = uploadResult.Url
		createProjectParam.Description = description
		createProjectParam.GroupId = groupId

		jsonData, err := json.Marshal(createProjectParam)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		httpURL := getServerUrl() + "/project/project/"
		req, err := http.NewRequest(http.MethodPost, httpURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		req.Header.Set("Cookie", cookieStr)
		req.Header.Add("Content-Length", strconv.Itoa(len(jsonData)))
		req.Header.Add("Content-Type", "application/json;charset=UTF-8")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Println("request create project ...")
		defer resp.Body.Close()

		createContent, err := io.ReadAll(resp.Body)
		outJsonData(createContent)

		if err != nil {
			fmt.Println(err.Error())
			return
		} else {
			var projectDetailResult ProjectDetailResult
			json.Unmarshal(createContent, &projectDetailResult)
			if projectDetailResult.Code != 200 {
				if needFormatOut() {
					fmt.Println(projectDetailResult.Msg)
				}

				return
			}

			if needFormatOut() {
				fmt.Println("Project '" + projectDetailResult.Data.ProjectName + "' is created successfully.")
			}

			scanProject(projectDetailResult.Data.ProjectId)
		}

	}
}

func updateProject(projectId uint64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	fmt.Println("Request to update the project...")

	var updateProjectParam UpdateProjectParam
	updateProjectParam.ProjectId = ProjectID
	updateProjectParam.ProjectName = ProjectName
	updateProjectParam.Description = Description
	updateProjectParam.ManifestFile = "1333333"
	if GroupID < 0 {
		updateProjectParam.GroupId = -1
	} else {
		updateProjectParam.GroupId = GroupID
	}

	jsonData, err := json.Marshal(updateProjectParam)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	url := getServerUrl() + "/project/project"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err.Error())
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
		fmt.Println("Update project response status:" + resp.Status)
		return
	}
}

func scanProject(projectId uint64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	var scanDTO KnowledgebaseScanDTO
	scanDTO.ProjectId = projectId
	jsonData, err := json.Marshal(scanDTO)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	param := string(jsonData)

	fmt.Println("Request to scan the project...")
	httpURL := getServerUrl() + "/knowledgebase/scanall/"
	req, err := http.NewRequest(http.MethodPost, httpURL,
		strings.NewReader(param))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	req.Header.Set("Cookie", cookieStr)
	req.Header.Add("Content-Length", strconv.Itoa(strings.Count(param, "")))
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		var jaxResult AjaxResult
		json.Unmarshal(content, &jaxResult)
		if jaxResult.Code != 200 {
			fmt.Println(jaxResult.Msg)
			return
		}
	}

	err = printCVEScanProcessBar(projectId)

	if err == nil {
		fmt.Println("CVE Scan is completed.")
		return
	} else {
		fmt.Println(err.Error())
		return
	}
}

func cancelProject(projectId uint64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	fmt.Println("Request to cancel the project scanning...")

	url := getServerUrl() + "/knowledgebase/cancelscan/" + strconv.FormatUint(projectId, 10)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println(err.Error())
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
				fmt.Println("Cancel project scanning error:" + jaxResult.Msg)
			} else {
				fmt.Println(jaxResult.Msg)
			}
		}

		return
	} else {
		fmt.Println("Cancel project scanning response status:" + resp.Status)
		return
	}
}

func deleteProject(projectId uint64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	fmt.Println("Request to delete the project...")

	url := getServerUrl() + "/project/project/" + strconv.FormatUint(projectId, 10)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		fmt.Println(err.Error())
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
				fmt.Println("Delete project error:" + jaxResult.Msg)
			} else {
				fmt.Println(jaxResult.Msg)
			}
		}
		return
	} else {
		fmt.Println("Delete project response status:" + resp.Status)
		return
	}
}

func getInfoProject(projectId uint64) (InfoProject, error) {
	var p InfoProject
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		return p, tokenError
	}

	url := getServerUrl() + "/project/project/" + strconv.FormatUint(projectId, 10)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return p, err
	}

	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return p, err
		}

		//outJsonData(content)

		var projectDetailResult ProjectDetailResult
		json.Unmarshal(content, &projectDetailResult)

		if projectDetailResult.Code != 200 {
			return p, errors.New(projectDetailResult.Msg)
		}

		p = projectDetailResult.Data
		return p, nil
	} else {
		fmt.Println("Get project response status:", resp.Status)

		return p, fmt.Errorf("Get project response status: %q", resp.Status)
	}
}

func getProjectCveScanStatus(projectId uint64) (ProjectCVEScanStatus, error) {
	var p ProjectCVEScanStatus
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		return p, tokenError
	}

	url := getServerUrl() + "/knowledgebase/getScanRate/CVE/" + strconv.FormatUint(projectId, 10)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return p, err
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
			return p, err
		}

		var projectCVEScanStatusResult ProjectCVEScanStatusResult
		json.Unmarshal(content, &projectCVEScanStatusResult)
		if projectCVEScanStatusResult.Code != 200 {
			fmt.Println(projectCVEScanStatusResult.Msg)
			return projectCVEScanStatusResult.Data, errors.New(projectCVEScanStatusResult.Msg)
		}

		p = projectCVEScanStatusResult.Data
		return p, nil
	} else {
		fmt.Println("GetProjectCveScanStatus response status:", resp.Status)

		return p, fmt.Errorf("GetProjectCveScanStatus response status: %q", resp.Status)
	}
}

func printCVEScanProcessBar(projectId uint64) error {
	projectCVEScanStatus, err := getProjectCveScanStatus(projectId)
	if err != nil {
		return err
	}

	if projectCVEScanStatus.CveScanStatus == "Running" || projectCVEScanStatus.CveScanStatus == "Waiting" {
		fmt.Println("CVE scan in progress...")
	}

	processRate := 0
	step := 0
	hasBar := false
	for projectCVEScanStatus.CveScanStatus == "Running" || projectCVEScanStatus.CveScanStatus == "Waiting" {
		hasBar = true
		time.Sleep(time.Duration(2) * time.Second)

		projectCVEScanStatus, err = getProjectCveScanStatus(projectId)
		if err != nil {
			return err
		}

		step = projectCVEScanStatus.CveScanRate - processRate
		if step > 0 {
			for i := 1; i < step; i++ {
				fmt.Print("█")
			}
		}

		processRate = projectCVEScanStatus.CveScanRate
	}

	if projectCVEScanStatus.CveScanStatus == "Exception" {
		if hasBar {
			fmt.Println()
		}

		return fmt.Errorf("The first scan of this project is failed, please try to rescan it again.")
	} else {
		if hasBar {
			step = 100 - processRate
			if step > 0 {
				for i := 1; i < step; i++ {
					fmt.Print("█")
				}
			}

			fmt.Println()
		}

		return nil
	}
}

func getLastestFile(projectId uint64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		return
	}

	url := getServerUrl() + "/project/lastestfile/" + strconv.FormatUint(projectId, 10)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err.Error())
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
			var infoManifestFileResult InfoManifestFileResult
			json.Unmarshal(content, &infoManifestFileResult)
			if infoManifestFileResult.Code != 200 {
				fmt.Println(infoManifestFileResult.Msg)
				return
			}

			if !reflect.DeepEqual(infoManifestFileResult.Data, InfoManifestFile{}) {
				formatOutManifestFile(infoManifestFileResult.Data)
			}
		}
	} else {
		fmt.Println("GetLastestFile response status:", resp.Status)
	}
}

func manifestSummary(projectId uint64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		return
	}

	url := getServerUrl() + "/dashboard/manifestSummary/" + strconv.FormatUint(projectId, 10)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err.Error())
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
			var manifestSummaryResult ManifestSummaryResult
			json.Unmarshal(content, &manifestSummaryResult)
			if manifestSummaryResult.Code != 200 {
				fmt.Println(manifestSummaryResult.Msg)
				return
			}

			if !reflect.DeepEqual(manifestSummaryResult.Data, ManifestSummaryStruct{}) {
				formatOutManifestSummary(manifestSummaryResult.Data)
			}
		}
	} else {
		fmt.Println("ManifestSummary response status:", resp.Status)
	}
}

func formatOutProjects(infoProjects []InfoProject) {
	length := len(infoProjects)
	if length > 0 {
		fmt.Println("┌────────────┬────────────────────────────────────────────────────────────────────────────────┬────────────────┬─────────────────────┬─────────────────────┐")
		fmt.Println("│ PROJECT ID │                                   PROJECT NAME                                 │   GROUP NAME   │     LAST SCANNED    │      LAST UPDATED   │")
		fmt.Println("├────────────┼────────────────────────────────────────────────────────────────────────────────┼────────────────┼─────────────────────┼─────────────────────┤")
		var line, nameLines, groupLines int
		for index, infoProject := range infoProjects {
			nameLines = getLineCount(len(infoProject.ProjectName), 79)
			groupLines = getLineCount(len(infoProject.GroupName), 15)

			line = max(nameLines, groupLines)
			if line > 1 {
				for i := 0; i < line; i++ {
					fmt.Printf("│ %-11s│ %-79s│ %-15s│ %-20s│ %-20s│\n", substring(strconv.FormatUint(infoProject.ProjectId, 10), i*11, (i+1)*11), substring(infoProject.ProjectName, i*79, (i+1)*79), substring(infoProject.GroupName, i*15, (i+1)*15), substring(formatDateTime(infoProject.LastScanned), i*20, (i+1)*20), substring(formatDateTime(infoProject.UpdateTime), i*20, (i+1)*20))
				}
			} else {
				fmt.Printf("│ %-11d│ %-79s│ %-15s│ %-20s│ %-20s│\n", infoProject.ProjectId, abbreviate(infoProject.ProjectName, 79), abbreviate(infoProject.GroupName, 15), formatDateTime(infoProject.LastScanned), formatDateTime(infoProject.UpdateTime))
			}

			if index == length-1 {
				fmt.Println("└────────────┴────────────────────────────────────────────────────────────────────────────────┴────────────────┴─────────────────────┴─────────────────────┘")
			} else {
				fmt.Println("├────────────┼────────────────────────────────────────────────────────────────────────────────┼────────────────┼─────────────────────┼─────────────────────┤")
			}
		}

		fmt.Println(" ")
	}
}

func formatOutProject(infoProject InfoProject) {
	if reflect.DeepEqual(infoProject, InfoProject{}) {
		fmt.Println("The specified project does not exist.")
		return
	}
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("%12v: %-64s %15s: %-36s\n", "Project Name", infoProject.ProjectName, "Group Name", infoProject.GroupName)
	fmt.Printf("%12v: %-64s %15s: %-36s\n", "Last Scanned", formatDateTime(infoProject.LastScanned), "Created By", infoProject.CreateBy)
	fmt.Printf("%12v: %-64s %15s: %-36s\n", "Last Updated", formatDateTime(infoProject.UpdateTime), "Created Date", formatDateTime(infoProject.CreateTime))
	fmt.Printf("%12v: %-142s\n", "Description", infoProject.Description)
}

func formatOutManifestFile(infoManifestFile InfoManifestFile) {
	if reflect.DeepEqual(infoManifestFile, InfoManifestFile{}) {
		fmt.Println("The specified manifest file does not exist.")
	} else {
		fmt.Printf("%12v: %-64s %15s: %-36s\n", "Distro Name", infoManifestFile.DistroName, "Distro Version", infoManifestFile.DistroVersion)
	}
}

func formatOutManifestSummary(manifestSummaryStruct ManifestSummaryStruct) {
	if reflect.DeepEqual(manifestSummaryStruct, ManifestSummaryStruct{}) {
		fmt.Println("The specified manifest file does not exist.")
	} else {
		fmt.Println("MANIFEST SUMMARY /")
		fmt.Println("┌─────────────────────────┬─────────────────────────────┬───────────────────────────┬───────────────────────────┬────────────────────────────┐")
		fmt.Printf("│            %-6d       │             %-6d          │            %-6d         │            %-6d         │             %-6d         │\n", manifestSummaryStruct.PackagesNum, manifestSummaryStruct.AllowlistedPackagesNum, manifestSummaryStruct.CvesNum, manifestSummaryStruct.AffectedCvesNum, manifestSummaryStruct.AllowlistedCVEsNum)
		fmt.Printf("│       All Packages      │      Allowlisted Packages   │          All CVEs         │       Vulnerable CVEs     │       Allowlisted CVEs     │\n")
		fmt.Println("└─────────────────────────┴─────────────────────────────┴───────────────────────────┴───────────────────────────┴────────────────────────────┘")
	}
}

type AjaxResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type UploadResult struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	FileName string `json:"fileName"`
	Url      string `json:"url"`
}

type CreateProjectParam struct {
	ProjectName  string `json:"projectName"`
	ManifestFile string `json:"manifestFile"`
	Description  string `json:"description"`
	GroupId      int64  `json:"groupId"`
}

type UpdateProjectParam struct {
	ProjectId    uint64 `json:"projectId"`
	ProjectName  string `json:"projectName"`
	Description  string `json:"description"`
	GroupId      int64  `json:"groupId"`
	ManifestFile string `json:"manifestFile"`
	ScanSetting  string `json:"scanSetting"`
}

type ProjectDetailResult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data InfoProject `json:"data"`
}

type InfoGroup struct {
	GroupId       int64         `json:"groupId"`
	GroupName     string        `json:"groupName"`
	Description   string        `json:"description"`
	CreateBy      string        `json:"createBy"`
	CreateTime    string        `json:"createTime"`
	UpdateBy      string        `json:"updateBy"`
	UpdateTime    string        `json:"updateTime"`
	InfoGroupUser InfoGroupUser `json:"infoGroupUser"`
	ProjectList   []InfoProject `json:"projectList"`
}

type InfoGroupUser struct {
	GroupUserId uint64 `json:"groupUserId"`
	GroupId     uint64 `json:"groupId"`
	UserId      uint64 `json:"userId"`
	Status      string `json:"status"`
	Role        string `json:"role"`
	RoleName    string `json:"roleName"`
	UserName    string `json:"userName"`
	Email       string `json:"email"`
	NickName    string `json:"nickName"`
	GroupName   string `json:"groupName"`
	CreateBy    string `json:"createBy"`
	CreateTime  string `json:"createTime"`
	UpdateBy    string `json:"updateBy"`
	UpdateTime  string `json:"updateTime"`
}

type InfoProject struct {
	ProjectId            uint64                  `json:"projectId"`
	ProjectName          string                  `json:"projectName"`
	GroupName            string                  `json:"groupName"`
	ProductCode          string                  `json:"productCode"`
	LicenseNumber        string                  `json:"licenseNumber"`
	VersionNum           string                  `json:"versionNum"`
	ManifestFile         string                  `json:"manifestFile"`
	SecurityRisk         string                  `json:"securityRisk"`
	Status               string                  `json:"status"`
	HighSeverityCVEs     int                     `json:"highSeverityCVEs"`
	LicenseRisk          string                  `json:"licenseRisk"`
	LastScanned          string                  `json:"lastScanned"`
	HasDistroInfo        string                  `json:"hasDistroInfo"`
	DelFlag              string                  `json:"delFlag"`
	CvssScoreType        string                  `json:"cvssScoreType"`
	ScanSetting          string                  `json:"scanSetting"`
	UserId               uint64                  `json:"userId"`
	Role                 string                  `json:"role"`
	CveScanStatus        string                  `json:"cveScanStatus"`
	CveScanRate          int                     `json:"cveScanRate"`
	CveScanErrorCode     int                     `json:"cveScanErrorCode"`
	CveScanUserAction    int                     `json:"cveScanUserAction"`
	CveScanSuccessNumber int                     `json:"cveScanSuccessNumber"`
	CveScanId            int                     `json:"cveScanId"`
	Description          string                  `json:"description"`
	GroupId              int64                   `json:"groupId"`
	EditPackageNotify    bool                    `json:"editPackageNotify"`
	EditPackageCount     int                     `json:"editPackageCount"`
	EnableDefaultSBOM    bool                    `json:"enableDefaultSBOM"`
	SearchValue          string                  `json:"searchValue"`
	CreateBy             string                  `json:"createBy"`
	CreateTime           string                  `json:"createTime"`
	UpdateBy             string                  `json:"updateBy"`
	UpdateTime           string                  `json:"updateTime"`
	Remark               string                  `json:"remark"`
	SeverityCVEs         SeverityCVEsStruct      `json:"severityCVEs"`
	Vulnerabilities      []VulnerabilitiesStruct `json:"vulnerabilities"`
}

type ProjectCVEScanStatusResult struct {
	Code int                  `json:"code"`
	Msg  string               `json:"msg"`
	Data ProjectCVEScanStatus `json:"data"`
}

type ProjectCVEScanStatus struct {
	CveScanStatus    string `json:"cveScanStatus"`
	CveScanErrorCode int    `json:"cveScanErrorCode"`
	CveScanRate      int    `json:"cveScanRate"`
}

type InfoManifestFileResult struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data InfoManifestFile `json:"data"`
}

type InfoManifestFile struct {
	FileId         uint64 `json:"fileId"`
	ProjectId      uint64 `json:"projectId"`
	FileName       string `json:"fileName"`
	HasDistroInfo  string `json:"hasDistroInfo"`
	FullDistroName string `json:"fullDistroName"`
	DistroName     string `json:"distroName"`
	DistroVersion  string `json:"distroVersion"`
	ProjectLabels  string `json:"projectLabels"`
	ReleaseTime    string `json:"releaseTime"`
	PackageNum     int    `json:"packageNum"`
}

type GroupTableDataInfo struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Total int         `json:"total"`
	Rows  []InfoGroup `json:"rows"`
}

type ProjectTableDataInfo struct {
	Code  int           `json:"code"`
	Msg   string        `json:"msg"`
	Total int           `json:"total"`
	Rows  []InfoProject `json:"rows"`
}

type ManifestSummaryResult struct {
	Code int                   `json:"code"`
	Msg  string                `json:"msg"`
	Data ManifestSummaryStruct `json:"data"`
}

type ManifestSummaryStruct struct {
	ProjectsNum            int `json:"projectsNum"`
	PackagesNum            int `json:"packagesNum"`
	AllowlistedPackagesNum int `json:"allowlistedPackagesNum"`
	CvesNum                int `json:"cvesNum"`
	AllowlistedCVEsNum     int `json:"allowlistedCVEsNum"`
	AffectedCvesNum        int `json:"affectedCvesNum"`
	ProhibitedLicensesNum  int `json:"prohibitedLicensesNum"`
	CisaNum                int `json:"cisaNum"`
}

type SeverityCVEsStruct struct {
	Critical int `json:"Critical"`
	High     int `json:"High"`
	Medium   int `json:"Medium"`
	Low      int `json:"Low"`
	None     int `json:"None"`
}

type VulnerabilitiesStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type KnowledgebaseScanDTO struct {
	ProjectId   uint64   `json:"projectId"`
	ProjectName string   `json:"projectName"`
	ScanFileIds []uint64 `json:"scanFileIds"`
}

func init() {
	rootCmd.AddCommand(projectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// projectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//projectCmd.Flags().BoolP("projectID", "id", false, "project ID")
}
