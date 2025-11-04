## sls-scan

Wind River Studio Security Scanner cli client

### Synopsis

Wind River Studio Security Scanner is a professional-grade security vulnerability scanner, specifically curated to meet the unique needs of embedded systems.
sls-scan is the command line interface client of the Wind River Studio Security Scanner system.

### Examples

```
  Access to the Wind River Studio Security Scanner system is subject to permission control. 
  Before using the sls-scan, it is necessary to use the user set command to set a token for authentication.
  
  Use the following syntax to run the sls-scan tool:
  sls-scan [command] [<command-arguments>] [command-options]

  For command output, changing the value of the --outputFormat parameter can change the output format of the result. for example:
  sls-scan --outputFormat OnlyJson   Display the results in JSON format
  sls-scan --outputFormat OnlyTable  Display the results in a table
  sls-scan --outputFormat OnlyAll    Both JSON and table format data are displayed

  Examples of user-related commands:
  sls-scan user set -t"Your user token"   set user token
  sls-scan user get                       get login user information
    
  Examples of group-related commands:
  sls-scan group create   -n "Your Group Name" -d "Description"                     create new group
  sls-scan group update   -g <GroupID> -n "New Group Name"  -d "New Description"    update group
  sls-scan group list                                                               view all groups 
  sls-scan group detail   -g <GroupID>                                              view details of one group
  sls-scan group delete   -g <GroupID>                                              delete one group that the current user has administrative privileges on
  sls-scan group members  -g <GroupID>                                              list all the members under one group
  sls-scan group projects -g <GroupID>                                              list all the projects under one group
	
  Examples of project-related commands:
  sls-scan project create -n "Your Project Name" -f "path/yourproject/sbomfile"                 create a new project
  sls-scan project create -n "Your Project Name" -f "path/yourproject/sbomfile" -g <GroupID>    create a new project under one group
  sls-scan project update -p <ProjectID> -n " New Project Name" -d "New Description"            update project information
  sls-scan project list                                                                         list all the projects
  sls-scan project detail -p <ProjectID>                                                        view details of one project
  sls-scan project delete -p <ProjectID>                                                        delete one project 
  sls-scan project rescan -p <ProjectID>                                                        rescan one project 
  sls-scan project cancel -p <ProjectID>                                                        cancel scanning of the specified project
  sls-scan project export -p <ProjectID>                                                        export SBOM file under one project
	
  Examples of package-related commands:
  sls-scan package query  -p <ProjectID> -n "<PakcageName>"       view details of one package thru pakcage name under one project
  sls-scan package detail -p <ProjectID> -m <PakcageID>           view all CVEs under cetain package (thru. package id) under one project
  sls-scan package detail -p <ProjectID> -n "<PakcageName>"       view all CVEs under cetain package (thru. package name) under one project
  sls-scan package export -p <ProjectID> -n "<PakcageName>"       export package file thru package name under one project
	
  Examples of CVE-related commands:
  sls-scan cve query  -p <ProjectID> -c <CVE-ID>                                     query CVE thru. CVE-ID under one project
  sls-scan cve query  -p <ProjectID> -z "<FuzzyQuery>"                               query CVE thru. fuzzy query conditions under one project
  sls-scan cve export -p <ProjectID> -c <CVE-ID>        -o "path/outfilename.xlsx"   export cve information thru. CVE-ID under one project
  sls-scan cve export -p <ProjectID> -z "<FuzzyQuery>"  -o "path/outfilename.xlsx"   export cve information thru. fuzzy query conditions under one project
  sls-scan cve cyclonedxExport -p <ProjectID> -o "path/out/filename.json"            export CycloneDX SBOM and VEX Report
  sls-scan cve detail -p <ProjectID> -c <CVE-ID>                                     view cve detail
	
```

### Options

```
  -h, --help                  Help for sls-scan
      --outputFormat string   Configure the output format of the result(Optional values include OnlyTable, onlyJson and All)
```

### SEE ALSO

* [sls-scan cve](sls-scan_cve.md)	 - Operation commands related to the project's CVE
* [sls-scan group](sls-scan_group.md)	 - Operation commands related to the group
* [sls-scan package](sls-scan_package.md)	 - Operation commands related to the project's package
* [sls-scan project](sls-scan_project.md)	 - Operation commands related to the project
* [sls-scan user](sls-scan_user.md)	 - Operation commands related to the user
* [sls-scan version](sls-scan_version.md)	 - Provides the version information of sls-scan

###### Auto generated by spf13/cobra on 19-Dec-2023
