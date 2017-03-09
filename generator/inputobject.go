package main

import (
	"fmt"
	"strings"

	"github.com/zook-ai/graphql-go/internal/schema"
)

// InputObject holds the name of a resolver
type InputObject struct {
	name   string
	fields []Field
}

// Field should be replaced with Arg
type Field struct {
	name  string
	typpe string
}

func newInputObject(t *schema.InputObject) *InputObject {
	var r InputObject
	for _, fieldName := range t.FieldOrder {
		field := t.Fields[fieldName]
		r.addField(field.Name, field.Type.String())
	}
	r.name = toPrivate(t.Name + "Input")
	return &r
}

func (i *InputObject) String() string {
	return fmt.Sprintf("\ntype %s struct {\n%s}\n", i.name, i.args())
}

func (i *InputObject) addField(name, typpe string) {
	//TODO translate type
	name = strings.ToUpper(name[:1]) + name[1:]
	i.fields = append(i.fields, Field{name, translate(typpe)})
}

func (i *InputObject) args() (args string) {
	for _, field := range i.fields {
		args += "\t" + field.name + "\t" + field.typpe + "\n"
	}
	return
}

func (f Field) String() string {
	return f.name + " " + f.typpe
}
