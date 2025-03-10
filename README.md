# sls-scan

[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://go.dev) 
[![License](https://img.shields.io/github/license/windriver/sls-scan)](License) <br>

# Getting Started

## Overview
`Wind River Studio Security Scanner` is a professional-grade security vulnerability scanner, specifically curated to meet the unique needs of embedded systems. 
The `sls-scan` is the command line interface of the Wind River Studio Security Scanner system.

* Version: 1.0
* Author: Zhiming Wang <zhiming.wang@windriver.com>
* Copyright © 2024 Wind River Systems, Inc.


# Tutorial

## Features
- sls-scan scaffolding based on
[cobra][1], [viper][2], [pflag][3].

[1]: https://github.com/spf13/cobra
[2]: https://github.com/spf13/viper
[3]: https://github.com/spf13/pflag

- The configuration items in the configuration file are unified with the command line parameters, and the command line parameters take precedence(Realized by [cobra][1], [viper][2] and [pflag][3]).

## Quik Start

```sh
$ git clone --depth 1 https://github.com/Wind-River/wr-sls-scan.git
$ cd sls-scan
$ go build
$ sls-scan help
```

## License

This project is under the Apache-2.0 License.
See the [LICENSE](LICENSE) file for the full license text.
