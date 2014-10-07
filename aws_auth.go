package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
)

func awsAuthFromFile(f string) (aws.Auth, error) {
	a := aws.Auth{}

	buf, err := ioutil.ReadFile(f)
	if err != nil {
		return a, err
	}

	lines := strings.Split(string(buf), "\n")
	if len(lines) < 2 {
		return a, fmt.Errorf("malformed awsfile")
	}

	a.AccessKey = lines[0]
	a.SecretKey = lines[1]
	return a, nil

}

func awsAuth(ctx *cli.Context) (aws.Auth, error) {
	auth, err := awsAuthFromFile(ctx.GlobalString("awsfile"))
	if err == nil {
		return auth, nil
	}
	auth, err = awsAuthFromFile(os.Getenv("AWS_FILE"))
	if err == nil {
		return auth, nil
	}
	auth = aws.Auth{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_KEY_ID"),
	}
	if auth.AccessKey != "" && auth.SecretKey != "" {
		return auth, nil
	}

	auth, err = awsAuthFromFile(os.ExpandEnv("${HOME}/.awssecret"))
	if err == nil {
		return auth, nil
	}

	return aws.Auth{}, fmt.Errorf("could not find AWS credentials")

}

func connectEc2(ctx *cli.Context) (*ec2.EC2, error) {
	auth, err := awsAuth(ctx)
	if err != nil {
		return nil, err
	}

	region := ctx.GlobalString("region")
	reg, ok := aws.Regions[region]
	if !ok {
		reg, ok = aws.Regions[os.Getenv("AWS_REGION")]
		if !ok {
			return nil, fmt.Errorf("could not find AWS region")
		}
	}

	return ec2.New(auth, reg), nil

}
