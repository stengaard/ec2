package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/cli"
)

var lsUsage = `list instances in your ec2 account

Search against the EC2 API for any and all matches to your search query
`

var defaultHeaders = []string{"InstanceId", "T:Name", "DNSName"}

var lsCliCmd = cli.Command{
	Name:   "ls",
	Usage:  "list instances ec2",
	Action: cmdLs,
	Flags: []cli.Flag{
		cli.StringSliceFlag{
			Name:  "H,headers",
			Usage: "Columns to display",
			Value: (*cli.StringSlice)(&defaultHeaders),
		},
		cli.StringSliceFlag{
			Name:  "x,add-headers",
			Usage: "Additional columns to show for each matched instance",
			Value: &cli.StringSlice{},
		},
		cli.BoolFlag{
			Name:  "n,no-headers",
			Usage: "Don't display column headers",
		},
	},
}

func cmdLs(ctx *cli.Context) {
	client, err := connectEc2(ctx)
	if err != nil {
		exit(err.Error())
	}

	args := ctx.Args()

	start := time.Now()
	inst := getInstances(client, args...)

	var (
		headers = expandStringSlice(ctx.StringSlice("headers"))
		xtra    = expandStringSlice(ctx.StringSlice("add-headers"))
	)

	printInstances(inst, append(headers, xtra...), !ctx.Bool("no-headers"))

	fmt.Fprintf(os.Stderr, "Found %d hosts in %0.1fsec\n", len(inst), time.Now().Sub(start).Seconds())

}

func expandStringSlice(s []string) []string {

	var exp func(a, b []string) []string
	exp = func(ok, toexpand []string) []string {
		if len(toexpand) == 0 {
			return ok
		}
		var elem string
		elem, toexpand = toexpand[0], toexpand[1:]
		elems := strings.Split(elem, ",")
		return exp(append(ok, elems[0]), append(elems[1:], toexpand...))
	}

	return exp([]string{}, s)
}
