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
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	APP_NAME = "sls-scan"
	VERSION  = "1.0.0"

	longDesc = `The {{.appName}} is the command line interface of the Wind River Studio Security Scanner system.
Wind River Studio Security Scanner is a professional-grade security vulnerability scanner, specifically curated to meet the unique needs of embedded systems.`

	usageExample = `  Use the following syntax to run the {{.appName}}:
  {{.appName}} [command] [<command-arguments>] [command-options]

`

	projectExample = `  {{.appName}} project cancel -p <ProjectID>                                             cancel scanning of the specified project
  {{.appName}} project create -n "Project Name" -f "path/sbomfile"                       create a new project
  {{.appName}} project create -n "Project Name" -f "path/sbomfile" -g <GroupID>          create a new project under one group
  {{.appName}} project delete -p <ProjectID>                                             delete one project  
  {{.appName}} project detail -p <ProjectID>                                             view details of one project
  {{.appName}} project export -p <ProjectID>                                             export SBOM file under one project 
  {{.appName}} project list  
  {{.appName}} project rescan -p <ProjectID>                                             rescan one project
  {{.appName}} project update -p <ProjectID> -n "New Project Name" -d "New Description"  update project information`

	packageExample = `  {{.appName}} package detail -p <ProjectID> -m <PakcageID>                           view all CVEs under cetain package (thru. package id) under one project
  {{.appName}} package detail -p <ProjectID> -n "<PakcageName>"                       view all CVEs under cetain package (thru. package name) under one project
  {{.appName}} package export -p <ProjectID> -n "<PakcageName>"                       export package file thru. package name under one project
  {{.appName}} package query  -p <ProjectID> -n "<PakcageName>"                       view details of one package thru pakcage name under one project`

	cveExample = `  {{.appName}} cve detail -p <ProjectID> -c <CVE-ID>                                      view cve detail
  {{.appName}} cve export -p <ProjectID> -z "<FuzzyQuery>"  -o "path/outfilename.xlsx"    export cve information thru. fuzzy query conditions under one project
  {{.appName}} cve query  -p <ProjectID> -z "<FuzzyQuery>"                                query CVE thru. fuzzy query conditions under one project
  {{.appName}} cve update -p <ProjectID> -c <cveId> -a <newStatus> -N <packageName>       Update the status of CVE.`

	groupExample = `  {{.appName}} group create   -n "Your Group Name" -d "Description"                      create new group
  {{.appName}} group delete   -g <GroupID>                                               delete one group that the current user has administrative privileges on
  {{.appName}} group detail   -g <GroupID>                                               view details of one group
  {{.appName}} group list                                                                view all groups 
  {{.appName}} group members  -g <GroupID>                                               list all the members under one group
  {{.appName}} group projects -g <GroupID>                                               list all the projects under one group
  {{.appName}} group update   -g <GroupID> -n "New Group Name"  -d "New Description"     update group`

	userExample = `  {{.appName}} user get                       get login user information  
  {{.appName}} user set -t"User Token"        set user token`

	formatOutExample = `  For command output, changing the value of the --outputFormat parameter can change the output format of the result. for example:
  {{.appName}} --outputFormat OnlyJson   Display the results in JSON format
  {{.appName}} --outputFormat OnlyTable  Display the results in a table
  {{.appName}} --outputFormat OnlyAll    Both JSON and table format data are displayed`

	rootExample = usageExample + `  Before using the {{.appName}}, it is necessary to use the user set command to set a token for authentication. The user token is generated by Wind River Studio Security Scanner.

  Examples of user-related commands:
` + userExample + `

  Examples of group-related commands:
` + groupExample + `

  Examples of project-related commands:
` + projectExample + `

  Examples of package-related commands:
` + packageExample + `

  Examples of CVE-related commands:
` + cveExample + `

` + formatOutExample
)

var (
	RESULT_OUT_FORMAT_ONLY_TABLE = "ONLYTABLE"
	RESULT_OUT_FORMAT_ONLY_JSON  = "ONLYJSON"
	RESULT_OUT_FORMAT_ALL        = "ALL"
	RESULT_OUT_FORMAT            = RESULT_OUT_FORMAT_ONLY_TABLE
	MarkdownDocs                 bool
	cfgFile                      = "sls-scan.yaml"

	nameFlag = pflag.StringP("name", "n", "", "Project, Package or Group name")

	projectIdFlag = pflag.Uint64P("projectId", "p", 0, "Project ID")

	sbomFileFlag = pflag.StringP("SBOMFile", "f", "", "Project SBOM File")

	descriptionFlag = pflag.StringP("description", "d", "", "Description")

	packageIdFlag   = pflag.Uint64P("packageId", "m", 0, "Package ID")
	outFileNameFlag = pflag.StringP("outFile", "o", "", "Out File Name")

	cveIdFlag              = pflag.StringP("cveId", "c", "", "CVE ID")
	fuzzyQueryFlag         = pflag.StringP("fuzzyQuery", "z", "", "Fuzzy Query")
	publishedDateBeginFlag = pflag.StringP("publishedDateBegin", "B", "", "Published Date Begin")
	publishedDateEndFlag   = pflag.StringP("publishedDateEnd", "E", "", "Published Date End")
	scoreComparatorFlag    = pflag.StringArrayP("scoreComparator", "s", []string{}, "Comparato for CVSS score")
	modifiedDateBegin      = pflag.StringP("modifiedDateBegin", "M", "", "Modified Date Begin")
	modifiedDateEnd        = pflag.StringP("modifiedDateEnd", "D", "", "Modified Date End")
	status                 = pflag.StringArrayP("status", "u", []string{}, "Cve Status")
	severity               = pflag.StringArrayP("severity", "v", []string{}, "Cve Severity")
	newStatusFlag          = pflag.StringP("newStatus", "a", "", "New Cve Status")
	packageNameFlag        = pflag.StringP("packageName", "N", "", "PackageName")

	groupIdFlag   = pflag.Int64P("groupId", "g", -1, "Group ID")
	userTokenFlag = pflag.StringP("userToken", "t", "", "User Token")

	sbomFormatFlag   = pflag.StringP("sbomFormat", "b", "SPDX", "Export SBOM File format, Two formats available: SPDX and CycloneDX")
	outputFormatFlag = pflag.String("outputFormat", "", "Configure the output format of the result(Optional values include OnlyTable, onlyJson and All)")

	helpFlag         = pflag.BoolP("help", "h", false, "Help for "+APP_NAME)
	markdownDocsFlag = pflag.BoolP("markdownDocs", "", false, "gen Markdown docs")
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              APP_NAME,
	Short:            "Wind River Studio Security Scanner cli client",
	Long:             formatExample(longDesc),
	Example:          formatExample(rootExample),
	TraverseChildren: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	sortCommands(rootCmd.Commands())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	if err != nil {
		rootCmd.SilenceUsage = false
		rootCmd.SilenceErrors = false

		c := color.New(color.FgHiRed)
		c.Println(err)
	}
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	//rootCmd.PersistentFlags().StringVarP(&RESULT_OUT_FORMAT, "outputFormat", "r", "OnlyTable", "Result output format")
	//viper.BindPFlag("result.output.format", rootCmd.PersistentFlags().Lookup("outputFormat"))
	//cobra.OnInitialize(initConfig)
	initConfig()

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	pflag.CommandLine.SetNormalizeFunc(wordSepNormalizeFunc)

	pflag.CommandLine.MarkHidden("name")
	pflag.CommandLine.MarkHidden("SBOMFile")
	pflag.CommandLine.MarkHidden("cveId")
	pflag.CommandLine.MarkHidden("description")

	pflag.CommandLine.MarkHidden("status")
	pflag.CommandLine.MarkHidden("severity")
	pflag.CommandLine.MarkHidden("modifiedDateBegin")
	pflag.CommandLine.MarkHidden("modifiedDateEnd")
	pflag.CommandLine.MarkHidden("scoreComparator")
	pflag.CommandLine.MarkHidden("publishedDateEnd")
	pflag.CommandLine.MarkHidden("publishedDateBegin")
	pflag.CommandLine.MarkHidden("fuzzyQuery")
	pflag.CommandLine.MarkHidden("outFile")
	pflag.CommandLine.MarkHidden("packageId")
	pflag.CommandLine.MarkHidden("projectId")
	pflag.CommandLine.MarkHidden("sbomFormat")
	pflag.CommandLine.MarkHidden("userToken")
	pflag.CommandLine.MarkHidden("groupId")
	pflag.CommandLine.MarkHidden("markdownDocs")
	pflag.CommandLine.MarkHidden("newStatus")
	pflag.CommandLine.MarkHidden("packageName")
	//rootCmd.SetVersionTemplate(customTemplate)
	//rootCmd.CompletionOptions.DisableDefaultCmd = true
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sls-scan.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	pflag.Parse()

	var formatFromConfig = viper.GetString("result.output.format")
	var formatFromFlag = strings.ToUpper(*outputFormatFlag)
	if formatFromConfig != "" {
		RESULT_OUT_FORMAT = formatFromConfig
	}
	if formatFromFlag != "" {
		if formatFromFlag != RESULT_OUT_FORMAT_ONLY_JSON && formatFromFlag != RESULT_OUT_FORMAT_ONLY_TABLE && formatFromFlag != RESULT_OUT_FORMAT_ALL {
			fmt.Println("Paramter outputFormat error, outputFormat value is OnlyTable, onlyJson or All.")
		} else {
			RESULT_OUT_FORMAT = formatFromFlag
			viper.Set("result.output.format", RESULT_OUT_FORMAT)
			viper.WriteConfig()
		}
	}

	MarkdownDocs = *markdownDocsFlag
}

// Sort commands by their Use field
func sortCommands(commands []*cobra.Command) {
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Use < commands[j].Use
	})
}

func GenDocs() {
	if MarkdownDocs {
		if err := doc.GenMarkdownTree(rootCmd, "./docs/md"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func wordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	from := []string{"-", "_"}
	to := "."
	for _, sep := range from {
		name = strings.Replace(name, sep, to, -1)
	}
	return pflag.NormalizedName(name)
}

func getServerUrl() string {
	return "https://studio.windriver.com/scan/api"
	//return "http://thor.mls.wrs.com/jxu6/api"
	//return "http://localhost:4200/api"
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("SLS-SCAN")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %v , err:%s \n", cfgFile, err))
	}

	//s := viper.AllSettings()
	//bs, _ := yaml.Marshal(s)

	//log.Printf("config settings  %s \r\n", string(bs))

	// Monitor configuration file changes in real time
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		//fmt.Printf("config file changed %s \n", in.Name)
	})

	//os.Setenv("SLS_CLI_USER_TOKEN", "test")

	//viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	//iper.AutomaticEnv()
	//viper.SetEnvPrefix("sls.cli.") // Sets the prefix of the environment variable to be read

	//viper.BindEnv("user.token", "SLS_CLI_USER_TOKEN")
	//fmt.Println(viper.Get("SLS_CLI_USER_TOKEN")) //通过.访问
	//viper.Set("SLS_CLI_USER_TOKEN", "QQQQQQQ")
	//viper.WriteConfig()
}

func checkUserToken() (string, error) {
	token := strings.TrimSpace(viper.GetString("user.token"))
	flag, cookie := setUserToken(token, false)
	if flag {
		return cookie, nil
	} else {
		return "", fmt.Errorf("User token verification failed.")
	}
}

func formatExample(tmpl string) string {
	return Tprintf(tmpl, map[string]interface{}{
		"appName": APP_NAME,
	})
}

// Tprintf renders a string from a given template string and field values
func Tprintf(tmpl string, data map[string]interface{}) string {
	t := template.Must(template.New("").Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		return ""
	}
	return buf.String()
}

func outJsonData(content []byte) {
	if RESULT_OUT_FORMAT == "ONLYJSON" || RESULT_OUT_FORMAT == "ALL" {
		fmt.Println(string(content))
	}
}

func needFormatOut() bool {
	return RESULT_OUT_FORMAT != "ONLYJSON"
}

func abbreviate(inTxt string, maxlen int) string {
	if len(inTxt) <= maxlen {
		return inTxt
	}
	newText := inTxt[0 : maxlen-3]
	newText = newText + "..."

	return newText
}

func getLineCount(strLen int, lineNum int) int {
	lineCount := strLen / lineNum
	if strLen%lineNum > 0 {
		lineCount++
	}

	return lineCount
}

func max(num1, num2 int) int {
	var result int

	if num1 > num2 {
		result = num1
	} else {
		result = num2
	}

	return result
}

func substring(str string, start, end int) string {
	var strLen = len(str)
	if start >= strLen {
		return ""
	}

	if end > strLen {
		return strings.TrimSpace(str[start:strLen])
	}

	return strings.TrimSpace(str[start:end])
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func formatDateTime(dateTime string) string {
	if strings.TrimSpace(dateTime) == "" {
		return ""
	}

	dateTime = strings.Replace(dateTime, "T", " ", 1)
	dateTime = strings.Replace(dateTime, "Z", " ", 1)
	if len(dateTime) > 19 {
		dateTime = dateTime[:19]
	}

	return dateTime
}
