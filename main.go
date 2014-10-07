// ec2 is simple AWS EC2 utility to list and connect to
// EC2 instances based.
//
// See http://github.com/stengaard/ec2
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"time"

	"github.com/codegangsta/cli"
	"github.com/goamz/goamz/aws"
)

var logger = log.New(os.Stderr, "", 0)
var usage = `swiss army tool knife for AWS EC2

ec2 can list and connect to instances. It searches the
EC2 API to fetch addresses of the instances.

ec2 looks for credentials in this order:
  * --awsfile on the command line
  * AWS_FILE environment variable
  * AWS_ACCESS_KEY_ID and AWS_SECRET_KEY_ID env vars
  * ~/.awssecret
In case the ec2 looks in a file the format is a two line
file with first line being the AWS access key id and the
second being the AWS secret key id.
`

func main() {

	app := cli.NewApp()
	app.Name = "ec2"
	app.Version = "0.9"
	app.Author = "Brian Stengaard"
	app.Email = "brian+ec2@stengaard.eu"
	app.Usage = usage
	app.Compiled = time.Now()
	app.Commands = []cli.Command{
		lsCliCmd,
		sshCliCmd,
		runCliCmd,
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "awsfile,a",
			Value: "$HOME/.awssecret",
			Usage: ".awssecret file - ",
		},
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "quiet down, please",
		},
		cli.StringFlag{
			Name:  "region,r",
			Usage: "Which AWS region to use (env var: AWS_REGION)",
			Value: aws.USEast.Name,
		},
	}

	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("quiet") {
			logger = log.New(ioutil.Discard, "", 0)
		}

		return nil
	}

	app.Run(os.Args)

}

func exit(s string) {
	if '\n' != s[len(s)-1] {
		s = s + "\n"
	}
	fmt.Fprintf(os.Stderr, s)
	os.Exit(1)
}

func exitErr(err error) {
	exit(err.Error())
}

func exitf(msg string, args ...interface{}) {
	exit(fmt.Sprintf(msg, args))
}

func getUser(ctx *cli.Context) string {
	username := ctx.String("user")
	if username == "" {
		u, err := user.Current()
		if err != nil {
			exitf("username not set and could not be fetched: %s", err)
		}
		username = u.Username
	}
	return username
}
