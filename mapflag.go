package main

import (
	"fmt"
	"strings"
)

type MapFlag map[string][]string

func (m MapFlag) String() string {
	s := make([]string, len(m))
	for k, v := range m {
		s = append(s, fmt.Sprintf("%s=%s", k, strings.Join(v, "|")))
	}
	return strings.Join(s, " ")
}
