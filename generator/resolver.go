package main

import (
	"fmt"
	"strings"

	"github.com/zook-ai/graphql-go/internal/schema"
)

// Resolver holds the name of a resolver
type Resolver struct {
	name  string
	args  Args
	funcs []string
}

func newResolver(t *schema.Object) *Resolver {
	var r Resolver
	r.name = toPrivate(t.Name + "Resolver")
	for _, fname := range t.FieldOrder {
		f := t.Fields[fname]
		var args Args
		for _, argName := range f.Args.FieldOrder {
			args = append(args, argFromInputValue(f.Args.Fields[argName]))
		}
		r.funcs = append(r.funcs, r.funcName(fname, f.Type.String(), args))
	}
	for _, i := range t.Interfaces {
		is := interfaces[i.Name]
		is.implementedBy = append(is.implementedBy, t.Name)
		interfaces[i.Name] = is
	}
	return &r
}

func (r *Resolver) getName() (f string) {
	return "*" + r.name
}

func (r *Resolver) funcName(name, returnType string, args Args) string {
	pName := strings.ToUpper(name[:1]) + name[1:]
	ret := translate(returnType)
	defaultRet := defaultRet(ret)
	return fmt.Sprintf("\nfunc (r %s) %s(%s) %s {\n\t%s\n}\n", r.getName(), pName, args.StringAsArgument(), ret, defaultRet)
}

// Struct echoes the struct of a resolver
func (r *Resolver) Struct() string {
	return fmt.Sprintf("\ntype %s struct{\n%s}\n", r.name, r.args.String())
}
