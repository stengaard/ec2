package main

import (
	"fmt"
	"reflect"
	"strings"

	"launchpad.net/goamz/ec2"
)

func printInstances(insts []*ec2.Instance, fields []string, header bool) {
	// get proper field names - throw away the values
	var err error
	fields, _, err = getFields(&ec2.Instance{}, fields)
	if err != nil {
		exit(err.Error())
	}

	// field length of headers
	widthMap := map[string]int{}
	for _, f := range fields {
		widthMap[f] = len(f) + 1
	}
	fieldVals := make([][]string, len(insts))

	// field length of values
	for i := range insts {
		_, vals, err := getFields(insts[i], fields)
		if err != nil {
			exit(err.Error())
		}
		for j, fval := range vals {
			width := len(fval) + 1
			max := widthMap[fields[j]]
			if width > max {
				widthMap[fields[j]] = width
			}
		}
		fieldVals[i] = vals
	}

	// generate template string
	templ := make([]string, len(fields))
	elem := "%%-%ds"
	for i, field := range fields {
		templ[i] = fmt.Sprintf(elem, widthMap[field])
	}
	template := strings.Join(templ, "  ") + "\n"

	// print headers
	out := make([]interface{}, len(fields))
	// hack around not being able to do []string{"foo"}... in fmt.Printf
	for i := range out {
		out[i] = fields[i]
	}
	var line string
	if header {
		line = fmt.Sprintf(template, out...)
		fmt.Print(line)
		buf := []byte(line)
		for i := 0; i < len(buf)-1; i++ {
			buf[i] = '-'
		}
		line = string(buf)
		fmt.Print(line)
	}

	// print instances values
	for i := range fieldVals {
		for j := range out {
			out[j] = fieldVals[i][j]
		}
		fmt.Printf(template, out...)

	}

	if header {
		fmt.Print(line)
	}

}

func name(i ec2.Instance) string {
	for _, t := range i.Tags {
		if t.Key == "Name" {
			return t.Value
		}
	}

	return i.InstanceId
}

// get field names and values for inst, where fields is  either ec2 tag
// name, struct tag name, xml element or ec2 api filter name
func getFields(inst *ec2.Instance, fields []string) ([]string, []string, error) {
	vals := make([]string, len(fields))
	for i, name := range fields {
		field, val, err := getFieldVal(inst, name)
		if err != nil {
			return nil, nil, err
		}
		fields[i] = field
		vals[i] = val

	}

	return fields, vals, nil
}

func dehyphenize(s string) string {
	ss := strings.Split(s, "-")
	o := ss[0]
	for i := 1; i < len(ss); i++ {
		p := ss[i]
		p = strings.ToUpper(p[:1]) + p[1:]
		o += p
	}
	return o
}

func getFieldVal(inst *ec2.Instance, name string) (string, string, error) {
	if len(name) < 2 {
		// no fields or tag can be < 2 chars in the
		// canonical form
		return "", "", fmt.Errorf("no such field %s", name)
	}

	// Get tag value
	isTag, tagval := name[:2], name[2:]
	if "t:" == strings.ToLower(isTag) {
		for _, t := range inst.Tags {
			if t.Key == tagval {
				return "T:" + t.Key, t.Value, nil
			}
		}
		// tags are not necessarily present - not present is not an error
		return name, "-", nil
	}

	//
	structTags := map[string]int{}
	t := reflect.TypeOf(*inst)
	for i := 0; i < t.NumField(); i++ {
		// get bar from `xml:"foo>bo>bar,baz,buz"`
		t := t.Field(i).Tag.Get("xml")
		if t == "" {
			continue
		}

		ts := strings.Split(t, ">")
		t = ts[len(ts)-1]

		ts = strings.Split(t, ",")
		t = ts[0]

		if t == "" {
			continue
		}
		structTags[t] = i

	}

	v := reflect.ValueOf(*inst)
	var field reflect.Value
	fname := ""

	// Use struct names as they are "prettier" by some standard.

	// try the struct tag
	if i, ok := structTags[name]; ok {
		field = v.Field(i)
		fname = t.Field(i).Name
	}
	// try the hyphenized version
	if i, ok := structTags[dehyphenize(name)]; ok {
		field = v.Field(i)
		fname = t.Field(i).Name
	}

	secgroup := map[string]bool{
		"group":          true,
		"groups":         true,
		"securitygroups": true,
	}

	// TODO : alias map based solution
	// map[string]func(inst) string, string
	// hack for  security-groups,
	if _, ok := secgroup[strings.ToLower(name)]; ok {
		groups := []string{}
		for i := range inst.SecurityGroups {
			groups = append(groups, inst.SecurityGroups[i].Name)
		}
		return "SecurityGroups", strings.Join(groups, ", "), nil
	}

	// try the raw field name
	if !field.IsValid() {
		field = v.FieldByName(name)
		fname = name

	}

	// still no luck - bum out
	if !field.IsValid() {
		return "", "", fmt.Errorf("no such field %s", name)
	}
	return fname, field.String(), nil
}
