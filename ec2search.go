package main

import (
	"fmt"
	"strings"

	"github.com/goamz/goamz/ec2"
)

// indentend lines are skipped as
// default filter
var fieldString = `architecture
availability-zone
 block-device-mapping.attach-time
architecture
availability-zone
 block-device-mapping.attach-time
 block-device-mapping.delete-on-termination
 block-device-mapping.device-name
 block-device-mapping.status
 block-device-mapping.volume-id
 client-token
dns-name
group-id
group-name
 iam-instance-profile.arn
image-id
instance-id
instance-lifecycle
instance-state-code
instance-state-name
instance-type
instance.group-id
instance.group-name
ip-address
kernel-id
key-name
 launch-index
 launch-time
 monitoring-state
 owner-id
 placement-group-name
platform
private-dns-name
private-ip-address
 product-code
 product-code.type
 ramdisk-id
 reason
 requester-id
 reservation-id
 root-device-name
 root-device-type
 source-dest-check
 spot-instance-request-id
state-reason-code
state-reason-message
subnet-id
tag-key
tag-value
virtualization-type
vpc-id
 hypervisor
 network-interface.description
 network-interface.subnet-id
 network-interface.vpc-id
 network-interface.network-interface.id
 network-interface.owner-id
 network-interface.availability-zone
 network-interface.requester-id
 network-interface.requester-managed
 network-interface.status
 network-interface.mac-address
 network-interface-private-dns-name
 network-interface.source-destination-check
 network-interface.group-id
 network-interface.group-name
 network-interface.attachment.attachment-id
 network-interface.attachment.instance-id
 network-interface.attachment.instance-owner-id
 network-interface.addresses.private-ip-address
 network-interface.attachment.device-index
 network-interface.attachment.status
 network-interface.attachment.attach-time
 network-interface.attachment.delete-on-termination
 network-interface.addresses.primary
 network-interface.addresses.association.public-ip
 network-interface.addresses.association.ip-owner-id
 association.public-ip
 association.ip-owner-id
 association.allocation-id
 association.association-id
`
var filterFields []string

func init() {
	for _, s := range strings.Split(fieldString, "\n") {
		if s == "" || s[0] == ' ' {
			continue
		}
		filterFields = append(filterFields, s)
	}
}

func asEc2Filter(raw map[string][]string) *ec2.Filter {
	f := ec2.NewFilter()
	for key, value := range raw {
		f.Add(key, value...)
	}
	return f
}

func searchEc2(client *ec2.EC2, searchTerm map[string][]string) <-chan *ec2.Instance {
	c := make(chan *ec2.Instance)

	var f *ec2.Filter
	if len(searchTerm) == 0 {
		f = ec2.NewFilter()
		f.Add("instance-state-name", "running")
	} else {
		f = asEc2Filter(searchTerm)
	}

	go func(filter *ec2.Filter) {
		resp, err := client.DescribeInstances(nil, filter)
		if err != nil {
			logger.Println(err)
			return
		}
		for _, res := range resp.Reservations {
			for _, inst := range res.Instances {
				c <- &inst
			}
		}
		close(c)
	}(f)

	return c
}

func getInstances(client *ec2.EC2, postfilter []string, filter map[string][]string) []*ec2.Instance {
	instMap := map[string]*ec2.Instance{}
	for i := range searchEc2(client, filter) {
		include := false
		if len(postfilter) == 0 {
			include = true
		}
		for _, p := range postfilter {
			if matches(i, p) {
				include = true
				break
			}
		}

		if include {
			instMap[i.InstanceId] = i
		}
	}

	inst := make([]*ec2.Instance, len(instMap))
	j := 0
	for _, i := range instMap {
		inst[j] = i
		j++
	}

	return inst
}

func matches(i *ec2.Instance, m string) bool {
	matchs := func(in ...string) bool {
		for _, b := range in {
			if strings.Contains(strings.ToLower(b), strings.ToLower(m)) {
				return true
			}
		}
		return false
	}

	if matchs(i.InstanceId, i.ImageId, i.DNSName, i.IPAddress, i.AvailabilityZone,
		i.InstanceType, i.KeyName, i.PrivateDNSName, i.PrivateIPAddress,
		i.SubnetId, i.State.Name) {
		return true
	}

	gs := make([]string, len(i.SecurityGroups)*2)
	for n := 0; n < len(i.SecurityGroups)*2; n += 2 {
		g := i.SecurityGroups[n/2]
		gs[n] = g.Id
		gs[n+1] = g.Name
	}
	if matchs(gs...) {
		return true
	}

	tagvals := make([]string, len(i.Tags))
	for n := 0; n < len(i.Tags); n++ {
		tagvals[n] = i.Tags[n].Value
	}
	fmt.Println(tagvals)

	return matchs(tagvals...)

}

type instance struct {
	ID string `aws:"instance-id"`
}
