package main

import (
	"fmt"
	"strings"
)

// stringListValue flag args can be supplied as -a av,fb or -a av -a fb (equivalent)
type stringListValue []string

func (s stringListValue) String() string {
	return strings.Join(s, ",")
}

func (ss *stringListValue) Set(s string) error {
	*ss = append(*ss, strings.Split(s, ",")...)
	return nil
}

type stringMapValue map[string][]string

func (f stringMapValue) String() string {
	s := []string{}
	for k, v := range f {
		s = append(s, fmt.Sprintf("%s=%s", k, strings.Join(v, "|")))
	}
	return strings.Join(s, " & ")
}

func (f stringMapValue) Set(s string) error {
	for _, elem := range strings.Split(s, ",") {
		l := strings.SplitN(elem, "=", 2)
		if len(l) == 1 {
			return fmt.Errorf("filters should be off the form 'foo=bar'")
		}
		f[l[0]] = append(f[l[0]], l[1])
	}
	return nil
}
