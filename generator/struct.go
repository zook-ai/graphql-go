package main

import "fmt"

type Struct struct {
	name   string
	fields []Field
}

func (s Struct) String() string {
	var fields string
	for _, f := range s.fields {
		fields += f.String() + "\n"
	}
	return fmt.Sprintf("type %s struct{\n\t%s\n}", s.name, fields)
}
