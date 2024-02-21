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
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	GroupID          int64
	GroupName        string
	GroupDescription string
)

// groupCmd represents the group command
var groupCmd = &cobra.Command{
	Use:     "group",
	Aliases: []string{"g", "gr"},
	Short:   "Operation commands related to the group",
	Long:    `Project user group, used for sharing projects among users, There are functions such as user group management, inviting users to join group, and creating projects for group. Subcommands include create, update, delete, list, detail, members and projects.`,
	Example: formatExample(groupExample),
	Run: func(cmd *cobra.Command, args []string) {
		subCommand := "list"
		if len(args) > 0 {
			subCommand = strings.ToLower(args[0])
		}

		if subCommand != "create" && subCommand != "update" && subCommand != "list" && subCommand != "detail" && subCommand != "delete" && subCommand != "projects" && subCommand != "members" && subCommand != "project" && subCommand != "member" {
			fmt.Printf("Unknown subcommand: %s", subCommand)
			return
		}

		executeGroupSubCommand(subCommand)
	},
}

func executeGroupSubCommand(subCommand string) {
	if subCommand == "list" {
		list()
		return
	} else if subCommand == "create" {
		createGroup()
		return
	}

	groupId := GroupID
	for groupId < 0 {
		fmt.Print("Please enter group ID:")
		fmt.Scanln(&groupId)
	}

	if subCommand == "detail" {
		detail(groupId)
	} else if subCommand == "delete" {
		delete(groupId)
	} else if subCommand == "update" {
		update(groupId)
	} else if subCommand == "members" || subCommand == "member" {
		members(groupId)
	} else if subCommand == "projects" || subCommand == "project" {
		projects(groupId)
	} else {
		fmt.Printf("Unrecognized subcommand: %s", subCommand)
	}
}

func createGroup() {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	groupName := strings.TrimSpace(GroupName) //*projectNameFlag
	for groupName == "" {
		fmt.Print("Please enter group name:")
		fmt.Scanln(&groupName)
		groupName = strings.TrimSpace(groupName)
	}

	description := GroupDescription

	var createGroupParam CreateGroupParam
	createGroupParam.GroupName = groupName
	createGroupParam.Description = description
	jsonData, err := json.Marshal(createGroupParam)
	if err != nil {
		fmt.Println("Fatal error: ", err.Error())
		return
	}

	//fmt.Println("Request to create group...")

	httpURL := getServerUrl() + "/group"
	req, err := http.NewRequest(http.MethodPost, httpURL, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
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
			var groupDetailResult GroupDetailResult
			json.Unmarshal(content, &groupDetailResult)
			if groupDetailResult.Code != 200 {
				fmt.Println(groupDetailResult.Msg)
				return
			}
			formatOutGroup(groupDetailResult.Data)
		}
	} else {
		fmt.Println("Create group response status:", resp.Status)
	}
}

func update(groupId int64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	groupName := strings.TrimSpace(GroupName)
	for groupName == "" {
		fmt.Print("Please enter group name:")
		fmt.Scanln(&groupName)
		groupName = strings.TrimSpace(groupName)
	}

	description := GroupDescription

	var updateGroupParam UpdateGroupParam
	updateGroupParam.GroupId = groupId
	updateGroupParam.GroupName = groupName
	updateGroupParam.Description = description

	jsonData, err := json.Marshal(updateGroupParam)
	if err != nil {
		fmt.Println("Fatal error: ", err.Error())
		return
	}

	httpURL := getServerUrl() + "/group"
	req, err := http.NewRequest(http.MethodPut, httpURL, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
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

func detail(groupId int64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	httpURL := getServerUrl() + "/group/list/" + strconv.FormatInt(groupId, 10)
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

		outJsonData(content)

		if needFormatOut() {
			var groupDetailResult GroupDetailResult
			json.Unmarshal(content, &groupDetailResult)
			if groupDetailResult.Code != 200 {
				fmt.Println(groupDetailResult.Msg)
				return
			}

			formatOutGroup(groupDetailResult.Data)
		}
	} else {
		fmt.Println("Get group detail response status: ", resp.Status)
	}
}

func delete(groupId int64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	fmt.Println("Request to delete the group...")

	url := getServerUrl() + "/group/" + strconv.FormatInt(groupId, 10)
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
				fmt.Println("Delete group error:" + jaxResult.Msg)
			} else {
				fmt.Println(jaxResult.Msg)
			}
		}
		return
	} else {
		fmt.Println("Delete group response status:" + resp.Status)
		return
	}
}

func list() {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	var queryBean InfoGroup
	if GroupID < 0 {
		queryBean.GroupId = -1
	} else {
		queryBean.GroupId = GroupID
	}

	queryBean.GroupName = GroupName

	jsonData, err := json.Marshal(queryBean)
	if err != nil {
		fmt.Println("Fatal error: ", err.Error())
		return
	}

	httpURL := getServerUrl() + "/group/list"
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
			var tableDataInfo GroupTableDataInfo
			json.Unmarshal(content, &tableDataInfo)
			if tableDataInfo.Code != 200 {
				fmt.Println(tableDataInfo.Msg)
				return
			}
			if tableDataInfo.Total > 0 {
				fmt.Println("Group List:")
				formatOutGroups(tableDataInfo.Rows)
			} else {
				fmt.Println("No groups.")
			}
		}
	} else {
		fmt.Println("List group response status: ", resp.Status)
	}
}

func members(groupId int64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	httpURL := getServerUrl() + "/group/groupUser/list?groupId=" + strconv.FormatInt(groupId, 10)
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

		outJsonData(content)

		if needFormatOut() {
			var tableDataInfo GroupUserTableDataInfo
			json.Unmarshal(content, &tableDataInfo)
			if tableDataInfo.Code != 200 {
				fmt.Println(tableDataInfo.Msg)
				return
			}
			if tableDataInfo.Total > 0 {
				fmt.Println("Group Member List:")
				formatOutGroupUsers(tableDataInfo.Rows)
			} else {
				fmt.Println("No group members.")
			}
		}
	} else {
		fmt.Println("List group member response status: ", resp.Status)
	}
}

func projects(groupId int64) {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	httpURL := getServerUrl() + "/project/project/statistic?groupId=" + strconv.FormatInt(groupId, 10)
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

		outJsonData(content)

		if needFormatOut() {
			var tableDataInfo ProjectTableDataInfo
			json.Unmarshal(content, &tableDataInfo)
			if tableDataInfo.Code != 200 {
				fmt.Println(tableDataInfo.Msg)
				return
			}
			if tableDataInfo.Total > 0 {
				fmt.Println("Group Project List:")
				formatOutProjects(tableDataInfo.Rows)
			} else {
				fmt.Println("No group projects.")
			}
		}
	} else {
		fmt.Println("List project response status: ", resp.Status)
	}
}

func formatOutGroup(infoGroup InfoGroup) {
	if reflect.DeepEqual(infoGroup, InfoGroup{}) {
		fmt.Println("The specified group does not exist.")
		return
	}

	fmt.Printf("%12v: %-64s %15s: %-36d\n", "Group Name", infoGroup.GroupName, "Group ID", infoGroup.GroupId)
	fmt.Printf("%12v: %-142s\n", "Description", infoGroup.Description)
	fmt.Printf("%12v: %-64s %15s: %-36s\n", "Created By", infoGroup.CreateBy, "Created Date", formatDateTime(infoGroup.CreateTime))
}

func formatOutGroups(infoGroups []InfoGroup) {
	length := len(infoGroups)
	if length > 0 {
		fmt.Println("┌────────────┬────────────────────────────────────────────────┬──────────────────┬────────────────────┐")
		fmt.Println("│  GROUP ID  │                   GROUP NAME                   │       ROLE       │    TOTAL PROJECTS  │")
		fmt.Println("├────────────┼────────────────────────────────────────────────┼──────────────────┼────────────────────┤")
		var line int
		for index, infoGroupVo := range infoGroups {
			line = getLineCount(len(infoGroupVo.GroupName), 47)

			if line > 1 {
				for i := 0; i < line; i++ {
					fmt.Printf("│ %-11s│ %-47s│ %-17s│ %-19s│\n", substring(strconv.FormatInt(infoGroupVo.GroupId, 10), i*11, (i+1)*11), substring(infoGroupVo.GroupName, i*47, (i+1)*47), substring(infoGroupVo.InfoGroupUser.RoleName, i*17, (i+1)*17), substring(strconv.Itoa(len(infoGroupVo.ProjectList)), i*19, (i+1)*19))
				}
			} else {
				fmt.Printf("│ %-11d│ %-47s│ %-17s│ %-19d│\n", infoGroupVo.GroupId, abbreviate(infoGroupVo.GroupName, 47), infoGroupVo.InfoGroupUser.RoleName, len(infoGroupVo.ProjectList))
			}

			if index < length-1 {
				fmt.Println("├────────────┼────────────────────────────────────────────────┼──────────────────┼────────────────────┤")
			} else {
				fmt.Println("└────────────┴────────────────────────────────────────────────┴──────────────────┴────────────────────┘")
			}

		}

		fmt.Println(" ")
	}
}

func formatOutGroupUsers(infoGroupUsers []InfoGroupUser) {
	length := len(infoGroupUsers)
	if length > 0 {
		fmt.Println("┌────────────────────────────────────┬────────────────────────────────────────────────┬───────────────────┐")
		fmt.Println("│              NICK NAME             │                     EMAIL                      │      MAX ROLE     │")
		fmt.Println("├────────────────────────────────────┼────────────────────────────────────────────────┼───────────────────┤")
		var line, nameLines, emailLines int
		for index, vo := range infoGroupUsers {
			nameLines = getLineCount(len(vo.NickName), 35)
			emailLines = getLineCount(len(vo.Email), 47)
			line = max(nameLines, emailLines)

			if line > 1 {
				for i := 0; i < line; i++ {
					fmt.Printf("│ %-35s│ %-47s│ %-18s│\n", substring(vo.NickName, i*35, (i+1)*35), substring(vo.Email, i*47, (i+1)*47), substring(vo.RoleName, i*18, (i+1)*18))
				}
			} else {
				fmt.Printf("│ %-35s│ %-47s│ %-18s│\n", abbreviate(vo.NickName, 35), abbreviate(vo.Email, 47), abbreviate(vo.RoleName, 18))
			}

			if index < length-1 {
				fmt.Println("├────────────────────────────────────┼────────────────────────────────────────────────┼───────────────────┤")
			} else {
				fmt.Println("└────────────────────────────────────┴────────────────────────────────────────────────┴───────────────────┘")
			}
		}
		fmt.Println(" ")
	}
}

type GroupUserTableDataInfo struct {
	Code  int             `json:"code"`
	Msg   string          `json:"msg"`
	Total int             `json:"total"`
	Rows  []InfoGroupUser `json:"rows"`
}

type GroupDetailResult struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data InfoGroup `json:"data"`
}

type CreateGroupParam struct {
	GroupName   string `json:"groupName"`
	Description string `json:"description"`
}

type UpdateGroupParam struct {
	GroupId     int64  `json:"groupId"`
	GroupName   string `json:"groupName"`
	Description string `json:"description"`
}

func init() {
	rootCmd.AddCommand(groupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// groupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// groupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
