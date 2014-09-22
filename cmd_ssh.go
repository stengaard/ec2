package main

import (
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"github.com/goamz/goamz/ec2"

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

func cmdSsh(ctx *cli.Context) {
	client, err := connectEc2(ctx)
	if err != nil {
		exit(err.Error())
	}

	args := []string(ctx.Args())
	if len(args) < 1 {
		exit("please supply at least one instance name")
	}

	username := getUser(ctx)

	post, pre := argsToSelector(args)
	inst := getInstances(client, post, pre)
	if len(inst) == 0 {
		exit("No running instances matched")
	}

	printInstances(inst, defaultHeaders, true)
	_, err = exec.LookPath("ssh")

	// ssh command is installed - lets use that
	if err == nil {
		for i := range inst {
			cmd := exec.Command("ssh", inst[i].DNSName)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			err := cmd.Run()
			if err != nil {
				logger.Println(err)
			}
		}
	} else {
		logger.Println("using internal ssh implementation")
		onAll(username, inst, func(inst *ec2.Instance, sesh *ssh.Session) error {
			sesh.Stdout = os.Stdout
			sesh.Stderr = os.Stderr
			sesh.Stdin = os.Stdin
			err := sesh.Shell()
			if err != nil {
				return err
			}

			sesh.Wait()

			return nil
		})
	}

}

type runresult struct {
	*ec2.Instance
	err error
}

func onAll(username string, instances []*ec2.Instance, fun func(*ec2.Instance, *ssh.Session) error) {

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
	res := make([]error, len(instances))
	start := time.Now()
	for i := range instances {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			res[idx] = runOn(conf, agnt, instances[idx], fun)
		}(i)
	}

	wg.Wait()
	dur := time.Now().Sub(start)

	errs := 0
	for _, err := range res {
		if err != nil {
			errs++
		}
	}

	logger.Printf("connected to %d (%d fails) in %s", len(instances), errs, dur)
}

func runOn(conf *ssh.ClientConfig, agnt agent.Agent, instance *ec2.Instance, fun func(*ec2.Instance, *ssh.Session) error) error {

	client, err := ssh.Dial("tcp", instance.DNSName+":22", conf)
	if err != nil {
		return err
	}

	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	return fun(instance, sess)
}
