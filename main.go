// ec2 is simple AWS EC2 utility to list and connect to
// EC2 instances based
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"launchpad.net/goamz/aws"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)
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
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "awsfile,a",
			Value: "$HOME/.awssecret",
			Usage: ".awssecret file - ",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "verbose output",
		},
		cli.StringFlag{
			Name:  "region,r",
			Usage: "Which AWS region to use",
			Value: aws.USEast.Name,
		},
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
