ec2
===

ec2 is helper tool to get info on those pesky EC2 instances.

```
ec2 -h
NAME:
   ec2 - swiss army tool knife for AWS EC2

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


USAGE:
   ec2 [global options] command [command options] [arguments...]

VERSION:
   0.9

COMMANDS:
   ls		list instances ec2
   ssh		ssh to instance
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --awsfile, -a '$HOME/.awssecret'	.awssecret file -
   --verbose				verbose output
   --region, -r 'us-east-1'		Which AWS region to use
   --version, -v			print the version
   --help, -h				show help
```

Requirements
============

 - `go` (`1.2` should do fine).
 - an `ssh-agent` running if you wanna do any sort of SSH stuff.
 - patience.

Features
========

Reasonably fast (bounded by the EC2 API mostly)
```
$ ec2 ls > /dev/null
InstanceId   T:Name                     DNSName
------------------------------------------------------------------------------------
------------------------------------------------------------------------------------
Found 250 hosts in 1.3sec
```
Searches concurrently across several selectors if some is given
```
$ ec2 ls m1.small > /dev/null
InstanceId   T:Name                 DNSName
-------------------------------------------------------------------------------
-------------------------------------------------------------------------------
Found 29 hosts in 1.3sec
```


Upcoming stuffs
===============
 - `ec2 run` that concurrently runs commands across filtered hosts
