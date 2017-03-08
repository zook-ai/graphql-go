package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zook-ai/graphql-go/internal/schema"
)

// [] support multiple entrypoints
// [] support interfaces
// [] support input Objects
// [] support enums
// [X] Don't overwrite if the file exists.
var (
	schemaString string
	stub         *os.File
	w            *bufio.Writer
	newFile      bool
)

//This is meant to generate golang stubs from a .graphql file
func main() {
	parseArguments()
	w = bufio.NewWriter(stub)
	defer stub.Close()

	s := schema.New()
	if err := s.Parse(schemaString); err != nil {
		panic(fmt.Sprintf("Problems parsing %s:\n\t %s", os.Args[1], err))
	}

	resolver := newResolver(s.EntryPointNames["query"], false)
	if newFile {
		writeDefault(resolver)
	}

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
	if !exists(out) {
		newFile = true
		stub, err = os.Create(out)
		if err != nil {
			fmt.Printf("Creation of file %s went badly: %s\n", out, err)
			os.Exit(1)
		}
		return
	}
	stub, err = os.OpenFile(out, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Opening file %s went badly: %s\n ", out, err)
		os.Exit(1)
	}
}

func writeDefault(r *Resolver) {

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
	schema = graphql.MustParseSchema(string(b), &` + r.name + `{})
}
func main() {}

`)

}

func exists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}
