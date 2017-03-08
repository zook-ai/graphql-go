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
	if len(t.Name) > 0 {
		r.name = strings.ToLower(t.Name[:1]) + t.Name[1:] + "Input"
	} else {
		r.name = "input"
	}
	return &r
}

func (i *InputObject) Struct() string {
	return fmt.Sprintf("type %s struct {\n%s}", i.name, i.args())
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
