## sls-scan group projects

List all projects for the specified group

### Synopsis

List all projects for the specified group that the current user is a member of.

```
sls-scan group projects [flags]
```

### Examples

```
 sls-scan group projects -g <GroupID>      list all the projects under one group
```

### Options

```
  -g, --groupId int   Group ID (default -1)
```

### Options inherited from parent commands

```
  -h, --help                  Help for sls-scan
      --outputFormat string   Configure the output format of the result(Optional values include OnlyTable, onlyJson and All)
```

### SEE ALSO

* [sls-scan group](sls-scan_group.md)	 - Operation commands related to the group

###### Auto generated by spf13/cobra on 19-Dec-2023
