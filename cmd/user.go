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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	USER_TOKEN string
)

// userCmd represents the token command
var userCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"u"},
	Short:   "Operation commands related to the user",
	Long: `Operation commands related to the user, Subcommands include set and get.
• The set subcommand is used to set the user token and obtain operation permissions for other commands. 
  The user token is generated by Wind River Studio Security Scanner. The user needs to open https://studio.windriver.com/scan/ and login, then open the user profile page. Apply to generate a user token and copy the user token. At last, run the set subcommand to configure the user token.

• The get subcommand is used to obtain the user information of the current login user.`,
	Example:          formatExample(userExample),
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand := "set"
		if len(args) > 0 {
			subCommand = args[0]
		}

		if subCommand != "set" && subCommand != "get" {
			fmt.Printf("Unknown subcommand: %s", subCommand)
			return
		}

		executeUserSubCommand(subCommand)
	},
}

func executeUserSubCommand(subCommand string) {
	if subCommand == "set" {
		token := USER_TOKEN
		successFlag, _ := setUserToken(token, true)
		if successFlag {
			getUserInfo()
		} else {
			fmt.Printf("Failed to set user token.")
		}

	} else if subCommand == "get" {
		getUserInfo()
	} else {
		fmt.Printf("Unrecognized subcommand: %s", subCommand)
	}
}

func getUserInfo() {
	cookieStr, tokenError := checkUserToken()
	if tokenError != nil {
		fmt.Println(tokenError.Error())
		return
	}

	httpURL := getServerUrl() + "/system/user/userInfo"
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
			var getUserResult GetUserResult
			json.Unmarshal(content, &getUserResult)
			if getUserResult.Code != 200 {
				fmt.Println(getUserResult.Msg)
				return
			}

			formatOutUser(getUserResult.Data)
		}

	} else {
		fmt.Println("Get user info response status:", resp.Status)
	}
}

func setUserToken(token string, isChange bool) (bool, string) {
	count := 3
	cookie := ""
	for i := 1; i <= count; i++ {
		for token == "" {
			fmt.Print("Please enter your token:")
			fmt.Scanln(&token)
			token = strings.TrimSpace(token)
			isChange = true
		}

		tempCookieStr, err := validate(token)
		if err != nil {
			fmt.Println(err.Error())
			if strings.Contains(err.Error(), "An runtime exception has occurred") {
				fmt.Println(err.Error())
				return false, cookie
			}

			isChange = false
			token = ""

			if i == 3 {
				viper.Set("user.token", token)
				viper.Set("result.output.format", RESULT_OUT_FORMAT)
				viper.WriteConfig()
				//viper.Set("Cookie", "")
				cookie = ""
				return false, cookie
			} else {
				fmt.Println("The user token entered is incorrect.")
			}

			//viper.Set("Cookie", "")
			cookie = ""
		} else {
			//viper.Set("Cookie", tempCookieStr)
			cookie = tempCookieStr
			if isChange {
				//fmt.Println("The user token verification passed.")
				viper.Set("user.token", token)
				viper.Set("result.output.format", RESULT_OUT_FORMAT)
				viper.WriteConfig()
				//os.Setenv("thor.user.token", token)
				//fmt.Println("Set user token to env.")
				//getUserInfo()
			}

			return true, cookie
		}
	}

	return false, cookie
}

func validate(token string) (string, error) {
	url := getServerUrl() + "/apikey/verify?apiKey=" + token
	req, err := http.NewRequest("GET", url, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == 200 {
		content, _ := io.ReadAll(resp.Body)

		var ajaxResult AjaxResult
		json.Unmarshal(content, &ajaxResult)
		if ajaxResult.Code != 200 {
			return "", fmt.Errorf(ajaxResult.Msg)
		}
		return resp.Header.Get("Set-Cookie"), nil
	} else {
		return "", fmt.Errorf("Response code is %s ", strconv.Itoa(resp.StatusCode))
	}
}

func formatOutUser(sysUser InfoSysUser) {
	if reflect.DeepEqual(sysUser, InfoSysUser{}) {
		fmt.Println("Get user info is null.")
		return
	}

	fmt.Println("The login user information is as follows:")
	fmt.Printf("%10v: %-80s \n", "User Name", sysUser.UserName)
	fmt.Printf("%10v: %-80s \n", "Nick Name", sysUser.NickName)
	fmt.Printf("%10v: %-80s \n", "EMail", sysUser.Email)
}

type GetUserResult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data InfoSysUser `json:"data"`
}

type InfoSysUser struct {
	UserId   uint64 `json:"userId"`
	UserName string `json:"userName"`
	NickName string `json:"nickName"`
	Email    string `json:"email"`
	CvesNum  string `json:"cvesNum"`
}

func init() {
	rootCmd.AddCommand(userCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tokenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tokenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
