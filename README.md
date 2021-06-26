[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)
[![GitHub release](https://img.shields.io/github/v/release/furkanthehuman/bssh)](https://GitHub.com/Naereen/StrapDown.js/releases/)
[![GitHub release](https://img.shields.io/github/workflow/status/furkanthehuman/bssh/bssh-goreleaser)](https://GitHub.com/Naereen/StrapDown.js/releases/)

<img src="logo.png" alt="logo"
	title="logo" width="150" height="150" />
# Bssh 
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Bssh](#Bssh)
  - [Usage](#usage)
    - [Adding Connections](#adding-connections-to-the-bucket)
    - [Connecting](#connecting)
    - [Parallel Code Execution](#parallel-code-execution)
  - [Installation](#installation)
    - [go-get](#installing-with-go-get)
    - [From The Source ](#installing-From-The-Source)
    - [Install From Binary](#Install-From-Binary)
  - [Notes](#notes)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


Bssh is an ssh bucket for categorizing and automating ssh connections. Also, with parallel command execution and connection checks(pings) over categories (namespaces).

![example gif](index.gif)
## Usage

### Adding Connections To The Bucket
Start by adding an ssh connection    
`bssh add --addr root@example.com  -n cluster --key .ssh/id_rsa --alias exmaple_conn`           
or     
`bssh add --addr root@example.com  -n cluster --password <SOME PASSWORD> --alias exmaple_conn`

`-n <namepsace>` flag enables you to categorize the connections.

### Connecting 
By typing `bssh` you can now see the connection and connect to it. Also, by using `bssh -n` you can filter by namespace You can add more connections to same namespace.

### Parallel Code Execution
Let's say you have 30 different servers in you cluster and want check if the time is synchronized. 
run:    
`bssh run -n date`    
select the namespace you want to check.   
Or you might want which host is up or down.     
`bssh ping -n`

for more commands check `bssh --help`



## Installation

### Installing With go-get

If you have go installed you can use go-get to install bssh.    
`go get https://github.com/FurkanTheHuman/bssh`

### Installing From The Source
You can install the project by cloning this repository. Of course you need go installed.         
`git clone https://github.com/FurkanTheHuman/bssh`    
`cd bssh`    
`go build .`    
`sudo cp bssh /usr/local/bin/bssh`     

### Install From Binary
You can directly install the executable from GitHub release. Choose the version that is compatible with you system from [here](https://github.com/FurkanTheHuman/bssh/releases)     
`wget -O bssh https://github.com/FurkanTheHuman/bssh/releases/download/v0.1.5/bssh_0.1.5_linux_amd64.tar.gz`     
`chmod +x bssh`     
`sudo cp bssh /usr/local/bin/bssh`   
