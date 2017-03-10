package main

import "fmt"

type Struct struct {
	name   string
	fields []Field
}

func (s Struct) String() string {
	var fields string
	for _, f := range s.fields {
		fields += "\t" + f.String() + "\n"
	}
	if len(fields) > 0 {
		fields = "\n" + fields
	}
	return fmt.Sprintf("\ntype %s struct{%s}\n", s.name, fields)
}
