package main

import (
	"fmt"

	"github.com/zook-ai/graphql-go/internal/schema"
)

// Resolver holds the name of a resolver
type Resolver struct {
	name  string
	args  Args
	s     Struct
	funcs []Func
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
		fun := r.newFunc(fname, f.Type.String(), args)
		if !newFile && exists.hasFunc(fun) {
			continue
		}
		r.funcs = append(r.funcs, fun)
	}
	for _, i := range t.Interfaces {
		is := interfaces[i.Name]
		is.implementedBy = append(is.implementedBy, t.Name)
		interfaces[i.Name] = is
	}
	r.s.name = r.name
	return &r
}

func (r *Resolver) getName() (f string) {
	return "*" + r.name
}

func (r *Resolver) newFunc(name, returnType string, args Args) Func {
	return newFunc(name, Field{name: "r", typpe: "*" + r.name}, args, Field{typpe: translate(returnType)})
}

func (r *Resolver) funcName(name, returnType string, args Args) string {
	pName := toPublic(name)
	ret := translate(returnType)
	defaultRet := defaultRet(ret)
	return fmt.Sprintf("\nfunc (r %s) %s(%s) %s {\n\t%s\n}\n", r.getName(), pName, args.StringAsArgument(), ret, defaultRet)
}

// Struct echoes the struct of a resolver
func (r *Resolver) Struct() string {
	if !newFile && exists["struct"].has(r.s.String()) {
		return ""
	}
	return fmt.Sprintf("\ntype %s struct{\n%s}\n", r.name, r.args.String())
}
