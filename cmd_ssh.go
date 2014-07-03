package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"sync"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"launchpad.net/goamz/ec2"

	"github.com/codegangsta/cli"
)

var (
	sshUsage = `SSH to an EC2 instance`
	sshDesc  = `SSH into an EC2 instance

Examples
   $ ec2 ssh i-abcdefgh
Looks up via the instance ID.

Lookup via Name tag
   $ ec2 ssh prod-www

If the lookup returns several hosts ec2 connects to them
sequentially.
`

	sshCliCmd = cli.Command{
		Name:        "ssh",
		Usage:       sshUsage,
		Description: sshDesc,
		Action:      cmdSsh,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "user,u",
				Usage: "username for ssh login",
			},
		},
	}
)

func init() {
	// hack username in there
	uf := sshCliCmd.Flags[0].(cli.StringFlag)
	u, _ := user.Current()
	uf.Value = u.Username
}

func cmdSsh(ctx *cli.Context) {
	client, err := connectEc2(ctx)
	if err != nil {
		exit(err.Error())
	}

	args := ctx.Args()
	if len(args) < 1 {
		exit("please supply at least one instance name")
	}

	inst := getInstances(client, args[0])
	if len(inst) == 0 {
		exit("No instances matched")
	}

	printInstances(inst, defaultHeaders, true)
	for i := range inst {
		cmd := exec.Command("ssh", inst[i].DNSName)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Run()
	}

}

func onAll(username string, instances []ec2.Instance, fun func(ec2.Instance, *ssh.Session)) {

	wg := &sync.WaitGroup{}

	agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		logger.Println(err)
		return
	}
	defer agentConn.Close()

	agnt := agent.NewClient(agentConn)

	conf := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.PublicKeysCallback(agnt.Signers)},
	}

	for i := range instances {
		wg.Add(1)
		go func(inst ec2.Instance, fun func(ec2.Instance, *ssh.Session)) {
			defer wg.Done()
			start := time.Now()
			runOn(conf, agnt, inst, fun)
			fmt.Println(name(inst), time.Now().Sub(start))
		}(instances[i], fun)
	}

	wg.Wait()
}

func runOn(conf *ssh.ClientConfig, agnt agent.Agent, instance ec2.Instance, fun func(ec2.Instance, *ssh.Session)) {

	client, err := ssh.Dial("tcp", instance.DNSName+":22", conf)
	if err != nil {
		logger.Println(err)
		return
	}

	if err != nil {
		logger.Println(err)
	}

	sess, err := client.NewSession()
	if err != nil {
		logger.Println(err)
		return
	}
	defer sess.Close()

	fun(instance, sess)
}
