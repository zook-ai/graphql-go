package main

import (
	"fmt"
	"strings"

	"github.com/zook-ai/graphql-go/internal/common"
	"github.com/zook-ai/graphql-go/internal/schema"
)

// Arg holds the name, type and notnull of an argument to a function§
type Arg struct {
	name string
	t    string
}

// Args is a list of arguments with print functionality
type Args []Arg

func argFromField(field *schema.Field) (a Arg) {
	a.name = field.Name
	a.t = field.Type.String()
	return
}

func argFromInputValue(in *common.InputValue) (a Arg) {
	a.name = in.Name
	a.t = translate(in.Type.String())
	return
}

func (args *Args) String() (sum string) {
	for _, a := range *args {
		sum += a.String() + " "
	}
	return
}

// StringAsArgument returns
func (args *Args) StringAsArgument() string {
	sum := args.String()
	if len(sum) > 0 {
		return fmt.Sprintf("args *struct{ %s }", sum)
	}
	return ""
}

func (args *Args) add(a Arg) {
	*args = append(*args, a)
}

func (a Arg) String() string {
	if len(a.name) > 0 {
		return fmt.Sprint(strings.ToUpper(a.name[:1]), a.name[1:], " ", a.t)
	}
	return translate(a.t)
}

func toPrivate(n string) (out string) {
	if len(n) > 0 {
		out = strings.ToLower(n[:1]) + n[1:]
	}
	return
}

func toPublic(n string) (out string) {
	if len(n) > 0 {
		out = strings.ToUpper(n[:1]) + n[1:]
	}
	return
}

func translate(qtype string) (gotype string) {
	required := qtype[len(qtype)-1:] == "!"
	if required {
		qtype = qtype[:len(qtype)-1]
	}
	gotype = convertType(qtype)
	if gotype[:1] != "*" && !required {
		gotype = "*" + gotype
	}
	return
}

func convertType(t string) (real string) {

	if t[:1] == "[" && t[len(t)-1:] == "]" {
		return "[]" + translate(t[1:len(t)-1])
	}

	if enums.has(t) {
		return "string"
	}

	if i, ok := inputs[t]; ok {
		return i.name
	}

	switch t {
	case "Int":
		real = "int"
	case "String":
		real = "string"
	case "Boolean":
		real = "bool"
	case "Float":
		real = "float32"
	case "ID":
		real = "graphql.ID"
	default:
		real = "*" + toPrivate(t+"Resolver")
	}
	return
}

func defaultRet(t string) (d string) {
	if len(t) > 0 {
		if t[:1] == "*" || t[:2] == "[]" {
			return "return nil"
		}
		switch t {
		case "int":
			return "return 0"
		case "string":
			return "return \"\""
		case "boolean":
			return "return false"
		case "float32":
			return "return 0"
		case "graphql.ID":
			return "return \"\""
		default:
			return "return &" + t + "{}"
		}
	}
	return ""
}
