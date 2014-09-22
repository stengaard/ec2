package main

import "strings"

// argsToSelector splits a string slice in to a list of post fetch filters and a
// filter suitable for EC2 querying.
func argsToSelector(in []string) (post []string, filter map[string][]string) {
	filter = map[string][]string{}
	for _, s := range in {
		if strings.Contains(s, "=") {
			kv := strings.SplitN(s, "=", 2)
			key := kv[0]
			value := kv[1]

			filter[key] = append(filter[key], value)
		} else {
			post = append(post, s)
		}
	}

	return
}
