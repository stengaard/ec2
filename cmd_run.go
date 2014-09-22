package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"github.com/codegangsta/cli"
	"github.com/goamz/goamz/ec2"
	"github.com/mgutz/ansi"
)

var (
	runUsage = `Run a command on EC2 instance(s)`
	runDesc  = `Run a command concurrently on a set of EC2 instances

Examples
   $ ec2 run i-abcdefgh ls
Executes the command on i-abcdefgh

Run ls on the host matching tag:Name=prod-www
   $ ec2 run prod-www -- ls

If the lookup returns several hosts ec2 connects to them
concurrently, executes the command and gathers the result.
It display the result of each execution.
`

	runCliCmd = cli.Command{
		Name:            "run",
		Usage:           runUsage,
		Description:     runDesc,
		Action:          cmdRun,
		SkipFlagParsing: true,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "user,u",
				Usage: "username for ssh login",
			},
		},
	}
)

func cmdRun(ctx *cli.Context) {
	client, err := connectEc2(ctx)
	if err != nil {
		exit(err.Error())
	}

	args := ctx.Args()

	var i int
	for i = 0; i < len(args); i++ {
		if args[i] == "--" {
			break
		}
	}

	// -- not found        -- found, but no command after that
	if i == len(args) || i+1 == len(args) {
		exit("no command given. Specify a command as 'run host1 host2 -- /usr/bin/mycmd'")
	}
	cmd := strings.Join(args[i+1:], " ")
	post, pre := argsToSelector(args[:i])

	username := getUser(ctx)
	insts := getInstances(client, post, pre)
	if len(insts) == 0 {
		exit("No running instances matched your query")
	}

	// logger
	logChan := make(chan *logline)
	go func() {
		width := 10
		for _, i := range insts {
			cw := len(name(i)) + 1
			if cw > width {
				width = cw
			}
		}

		template := fmt.Sprintf("%%%ds: %%s", width)

		for line := range logChan {
			if line.err {
				line.line = ansi.Color(line.line, "red")
			}
			logger.Printf(template, name(line.origin), line.line)
		}
	}()

	onAll(username, insts, func(inst *ec2.Instance, sesh *ssh.Session) error {

		log := func(f io.Reader, errDev bool) {
			s := bufio.NewScanner(f)
			for s.Scan() {
				line := s.Text()
				logChan <- &logline{
					origin: inst,
					line:   line,
					err:    errDev,
					Time:   time.Now(),
				}
			}

			if err := s.Err(); err != nil {
				dev := "stdout"
				if errDev {
					dev = "stderr"
				}
				logger.Printf("error reading from %s on %s: %s", dev, name(inst), err)
			}
		}

		stdout, err := sesh.StdoutPipe()
		if err != nil {
			return nil
		}

		stderr, err := sesh.StderrPipe()
		if err != nil {
			return nil
		}

		go log(stdout, false)
		go log(stderr, true)

		return sesh.Run(cmd)

	})
	close(logChan)
}

type logline struct {
	origin *ec2.Instance
	line   string
	err    bool
	time.Time
}
