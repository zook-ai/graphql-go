package main

import (
	"strings"

	"github.com/zook-ai/graphql-go/internal/schema"
)

// InputObject holds the name of a resolver
type InputObject struct {
	s *Struct
}

// Field should be replaced with Arg
type Field struct {
	name  string
	typpe string
}

func newInputObject(t *schema.InputObject) *InputObject {
	var r InputObject
	r.s = &Struct{}
	for _, fieldName := range t.FieldOrder {
		field := t.Fields[fieldName]
		r.addField(field.Name, field.Type.String())
	}
	r.s.name = toPrivate(t.Name + "Input")
	return &r
}

func (i *InputObject) String() string {
	if exists.hasStruct(i.s) {
		return ""
	}
	return i.s.String()
}

func (i *InputObject) addField(name, typpe string) {
	//TODO translate type
	name = strings.ToUpper(name[:1]) + name[1:]
	i.s.fields = append(i.s.fields, Field{name, translate(typpe)})
}

func (i *InputObject) args() (args string) {
	for _, field := range i.s.fields {
		args += "\t" + field.name + "\t" + field.typpe + "\n"
	}
	return
}

func (f Field) String() string {
	return f.name + " " + f.typpe
}
