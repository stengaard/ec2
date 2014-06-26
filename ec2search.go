package main

import (
	"strings"
	"sync"

	"launchpad.net/goamz/ec2"
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

func searchEc2(client *ec2.EC2, line ...string) <-chan *ec2.Instance {
	c := make(chan *ec2.Instance)
	filters := []*ec2.Filter{}
	for _, term := range line {
		values := strings.SplitN(term, "=", 2)

		var filterKeys []string
		if len(values) > 1 {
			filterKeys = []string{values[0]}
			term = values[1]
		} else {
			filterKeys = filterFields
		}

		for _, fil := range filterKeys {
			f := ec2.NewFilter()
			f.Add(fil, strings.Split(term, ",")...)
			filters = append(filters, f)
		}
	}

	// if no filters - get all instances
	if len(filters) == 0 {
		filters = []*ec2.Filter{nil}
	}

	wg := &sync.WaitGroup{}
	for _, filter := range filters {
		wg.Add(1)
		go func(filter *ec2.Filter) {
			defer wg.Done()
			resp, err := client.Instances(nil, filter)
			if err != nil {
				logger.Println(err)
			}
			for _, res := range resp.Reservations {
				for _, inst := range res.Instances {
					c <- &inst
				}
			}
		}(filter)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	return c
}

func getInstances(client *ec2.EC2, line ...string) []*ec2.Instance {
	inst := []*ec2.Instance{}
	for i := range searchEc2(client, line...) {
		inst = append(inst, i)
	}
	return inst
}
