package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Dialer func(network, addr string) (net.Conn, error)

func sshAgent() agent.Agent {
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return agent.NewClient(sock)
}

func sshClientConfig(agent agent.Agent, user string) *ssh.ClientConfig {
	signers, err := agent.Signers()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signers...)},
		HostKeyCallback: nil,
	}
	return config
}

func tunneler(client *ssh.Client) Dialer {
	return func(network, addr string) (net.Conn, error) {
		return client.Dial(network, addr)
	}
}

func proxiedExec(cmd string, client *ssh.Client, sshagent agent.Agent) ([]byte, error) {
	session, err := client.NewSession()
	if err != nil {
		return []byte{}, err
	}
	defer session.Close()

	err = agent.RequestAgentForwarding(session)
	if err != nil {
		return nil, err
	}

	err = agent.ForwardToAgent(client, sshagent)
	if err != nil {
		return nil, err
	}

	return session.CombinedOutput(cmd)
}

func proxiedHttpClient(client *ssh.Client) http.Client {
	return http.Client{Transport: &http.Transport{Dial: tunneler(client)}}
}

func proxyClient(proxy, user string, sshagent agent.Agent) *ssh.Client {
	config := sshClientConfig(sshagent, user)
	client, err := ssh.Dial("tcp", proxy+":22", config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return client
}

func targetClient(target, user string, client *ssh.Client, sshagent agent.Agent) *ssh.Client {
	config := sshClientConfig(sshagent, user)
	conn, err := client.Dial("tcp", target+":22")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	nextConn, sshChan, req, err := ssh.NewClientConn(conn, target, config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return ssh.NewClient(nextConn, sshChan, req)
}

func remoteExecHostname(proxy, target, user string) {
	sshagent := sshAgent()
	client := proxyClient(proxy, user, sshagent)
	out, err := proxiedExec("hostname", client, sshagent)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(out))

	nextClient := targetClient(target, user, client, sshagent)
	out, err = proxiedExec("hostname", nextClient, sshagent)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println(strings.Trim(string(out), " \n"))
	}
}

func tunneledHttpGet(proxy, target, user string) {
	sshagent := sshAgent()
	client := proxyClient(proxy, user, sshagent)
	c := http.Client{Transport: &http.Transport{Dial: tunneler(client)}}

	resp, err := c.Get("http://what-is-my-ip.net/?text")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(body))

	nextClient := targetClient(target, user, client, sshagent)
	c = http.Client{Transport: &http.Transport{Dial: tunneler(nextClient)}}

	resp, err = c.Get("http://what-is-my-ip.net/?text")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func main() {
	var proxy, target, user string
	flag.StringVar(&proxy, "proxy", "", "Provide proxy machine")
	flag.StringVar(&target, "target", "", "Provide target machine")
	flag.StringVar(&user, "user", "", "Provide user")
	flag.Parse()
	remoteExecHostname(proxy, target, user)
	tunneledHttpGet(proxy, target, user)
}
