package main

import (
	"fmt"

	"github.com/zook-ai/graphql-go/internal/schema"
)

// Interface holds the graphQL interface types
type Interface struct {
	name          string
	methods       methods
	implementedBy []string
}

type method struct {
	name  string
	typpe string
	args  Args
}

type methods []method

func newInterface(t *schema.Interface) *Interface {
	var i Interface
	i.name = toPrivate(t.Name)
	for _, fieldName := range t.FieldOrder {
		field := t.Fields[fieldName]
		i.methods.add(newMethod(field))
	}
	return &i
}

func (i Interface) String() string {
	return fmt.Sprintf("\ntype %s interface {\n%s}\n", i.name, i.methods.String())
}

func newMethod(f *schema.Field) (m method) {
	//need to create args
	m.name = f.Name
	m.typpe = translate(f.Type.String())
	return
}

func (m method) String() string {
	return "\t" + toPublic(m.name) + "() \t" + m.typpe + "\n"
}

func (ms *methods) add(method method) {
	*ms = append(*ms, method)
}

func (ms *methods) String() (out string) {
	for _, m := range *ms {
		out += m.String()
	}
	return
}
