package main

import (
	"fmt"
	"strings"
)

// Resolver holds the name of a resolver
type Resolver struct {
	name     string
	required bool
}

func newResolver(typeName string, required bool) *Resolver {
	var r Resolver
	if len(typeName) > 0 {
		r.name = strings.ToLower(typeName[:1]) + typeName[1:] + "Resolver"
	} else {
		r.name = "resolver"
	}
	r.required = required
	return &r
}

func (r *Resolver) getName() (f string) {
	return "*" + r.name
}

func (r *Resolver) funcName(name, returnType string, required bool, args Args) string {
	pName := strings.ToUpper(name[:1]) + name[1:]
	ret := convertType(returnType)
	defaultRet := defaultRet(ret)
	return fmt.Sprintf("\nfunc (r %s) %s(%s) %s {\n\t%s\n}\n", r.getName(), pName, args.String(), ret, defaultRet)
}

func (r *Resolver) structString() string {
	return fmt.Sprintf("\ntype %s struct{}\n", r.name)
}

func convertType(t string) (real string) {
	nomatch := false
	required := t[len(t)-1:] == "!"
	if required {
		t = t[:len(t)-1]
	}

	if _, contains := enums[t]; contains {
		t = "String"
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
		real = newResolver(t, required).getName()
		nomatch = true
	}

	if !nomatch && !required {
		real = "*" + real
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
