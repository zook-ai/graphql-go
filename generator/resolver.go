package main

import (
	"fmt"
	"strings"
)

// Resolver holds the name of a resolver
type Resolver struct {
	name string
}

func newResolver(typeName string) *Resolver {
	var r Resolver
	if len(typeName) > 0 {
		r.name = strings.ToLower(typeName[:1]) + typeName[1:] + "Resolver"
	} else {
		r.name = "resolver"
	}

	return &r
}

func (r *Resolver) getName() (f string) {
	return "*" + r.name
}

func (r *Resolver) funcName(name, returnType string, required bool, args Args) string {
	pName := strings.ToUpper(name[:1]) + name[1:]
	ret := translate(returnType)
	defaultRet := defaultRet(ret)
	return fmt.Sprintf("\nfunc (r %s) %s(%s) %s {\n\t%s\n}\n", r.getName(), pName, args.String(), ret, defaultRet)
}

func (r *Resolver) structString() string {
	return fmt.Sprintf("\ntype %s struct{}\n", r.name)
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
		real = newResolver(t).getName()
	}
	return
}

func defaultRet(t string) (d string) {
	if len(t) > 0 {
		if t[:1] == "*" {
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
