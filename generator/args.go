package main

import (
	"fmt"
	"strings"
)

// Arg holds the name, type and notnull of an argument to a function§
type Arg struct {
	name     string
	t        string
	required bool
}

// Args is a list of arguments with print functionality
type Args []Arg

func (args Args) String() string {
	var sum string
	for _, a := range args {
		sum += a.String()
	}
	if len(sum) > 0 {
		return fmt.Sprintf("args *struct{ %s }", sum)
	}
	return ""
}

func (a Arg) String() string {
	return fmt.Sprint(strings.ToUpper(a.name[:1]), a.name[1:], " ", convertType(a.t))
}
