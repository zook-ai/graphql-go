package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"bufio"

	"strings"

	"github.com/zook-ai/graphql-go/internal/schema"
)

var (
	schemaString string
	stub         *os.File
	w            *bufio.Writer
	resolver     *Resolver
)

// Resolver holds the name of a resolver
type Resolver struct {
	name     string
	required bool
}

// Arg holds the name, type and notnull of an argument to a function§
type Arg struct {
	name     string
	t        string
	required bool
}

// Args is a list of arguments with print functionality
type Args []Arg

func (args Args) String() string {
	var sum string
	for _, a := range args {
		sum += a.String()
	}
	if len(sum) > 0 {
		return fmt.Sprintf("args *struct{ %s }", sum)
	}
	return ""
}

func (a Arg) String() string {
	return fmt.Sprint(strings.ToUpper(a.name[:1]), a.name[1:], " ", convertType(a.t))
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
	if !r.required {
		f += "*"
	}
	return f + r.name
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
		real = "uint64" // TODO check if this is correct
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
		case "uint64":
			return "return 0"
		default:
			return "return " + t + "{}"
		}
	}
	return ""
}

//This is meant to generate golang stubs from a .graphql file
func main() {
	parseArguments()
	w = bufio.NewWriter(stub)
	defer stub.Close()

	s := schema.New()
	if err := s.Parse(schemaString); err != nil {
		panic(fmt.Sprintf("Problems parsing %s:\n\t %s", os.Args[1], err))
	}

	resolver = newResolver(s.EntryPointNames["query"], false)
	writeDefault()
	for _, o := range s.Objects {
		resolver = newResolver(o.Name, false)
		w.WriteString(resolver.structString())
		for _, fname := range o.FieldOrder {
			f := o.Fields[fname]
			var args Args
			for _, argName := range f.Args.FieldOrder {
				args = append(args, Arg{argName, f.Args.Fields[argName].Type.String(), false})
			}
			w.WriteString(resolver.funcName(fname, f.Type.String(), false, args))
		}
	}

	w.Flush()
}

func parseArguments() {
	if len(os.Args) < 3 {
		fmt.Println("Usage is: ./generator [input.graphql] [output.go]")
		os.Exit(1)
	}
	//Read in file
	in := os.Args[1]
	schemaBytes, err := ioutil.ReadFile(in)
	if err != nil {
		fmt.Printf("Could not read %s: [%s]\n", in, err)
		os.Exit(1)
	}
	schemaString = string(schemaBytes)

	//Open output path
	out := os.Args[2]
	os.Remove(out)
	stub, err = os.Create(out)
	if err != nil {
		fmt.Printf("Creation of file %s went badly: %s\n", out, err)
		os.Exit(1)
	}
}

func writeDefault() {

	w.WriteString(`package main

import (
	"io/ioutil"
	graphql "github.com/neelance/graphql-go"
)

var schema *graphql.Schema

func init() {
	var err error
	b, err := ioutil.ReadFile("schema.gql")
	if err != nil {
		panic(err)
	}
	schema = graphql.MustParseSchema(string(b), &` + resolver.name + `{})
}
func main() {}

`)

}
