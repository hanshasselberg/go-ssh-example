# Golang ssh examples

This demonstrates how to execute commands and HTTP GET remotely with golang ssh. But this is not the simple case only. It assumes you are using your local machine to go to a target machine via a proxy. The goal is to execute the commands on the target as well as HTTP GET on its behalf. It is a machine you wouldn't have access to without the proxy.

## Running

Thats how I can run it. It assumes that the user names are the same on both hosts. Target in this case is a private ip in some cloud. The proxy is reachable from your machine:

```
$ go run main.go -proxy 54.247.84.60 -target 10.12.8.7 -user hans
```
