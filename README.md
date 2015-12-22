# Golang ssh examples

This demonstrates how to execute commands and HTTP GET remotely with golang ssh. But this is not the simple case only. It assumes you are using your local machine to go to a target machine via a proxy. The goal is to execute the commands on the target as well as HTTP GET on its behalf. It is a machine you wouldn't have access to without the proxy.

## Running

Thats how I can run it. It assumes that the user names are the same on both hosts. The proxy needs to be accessible from your machine and the target only from the proxy:

```
$ go run main.go -proxy 54.0.0.1 -target 10.0.0.1 -user hans
```
