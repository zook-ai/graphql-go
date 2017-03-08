package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zook-ai/graphql-go/internal/schema"
)

// [] support mutation
// [] support subscription
// [] support interfaces
// [X] support input Objects
// [X] support enums
// [X] support arrays
// [] support unions
// [X] Don't overwrite if the file exists.
// [] parse output files and only add missing methods
type existMap map[string]interface{}

var (
	schemaString string
	stub         *os.File
	w            *bufio.Writer
	newFile      bool
	enums        existMap
	inputs       map[string]*InputObject
)

func (e existMap) has(key string) bool {
	_, ok := e[key]
	return ok
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
	var tmp Resolver
	resolver := &tmp
	resolver.name = resolverName(s.EntryPointNames["query"])
	if newFile {
		writeDefault(resolver)
	}

	// Finding inputs and enums
	inputs = make(map[string]*InputObject)
	enums = make(existMap)
	for _, t := range s.Types {
		switch t := t.(type) {
		case *schema.Enum:
			enums[t.Name] = true
		case *schema.InputObject:
			inputs[t.Name] = newInputObject(t)
		}
	}

	// Going through objects and creating resolvers
	for _, o := range s.Objects {
		r := newResolver(o)
		w.WriteString(r.Struct())
		for _, f := range r.funcs {
			w.WriteString(f)
		}
	}

	// Go through input types and create structs
	for _, t := range inputs {
		w.WriteString(t.Struct())
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
