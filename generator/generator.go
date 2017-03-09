package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/zook-ai/graphql-go/internal/schema"
)

// [] support mutation
// [] support subscription
// [X] support interfaces
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
	interfaces   map[string]*Interface
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
	resolver.name = toPrivate(s.EntryPointNames["query"] + "Resolver")
	if newFile {
		writeDefault(resolver)
	}

	// Finding inputs and enums
	inputs = make(map[string]*InputObject)
	enums = make(existMap)
	interfaces = make(map[string]*Interface)
	for _, t := range s.Types {
		switch t := t.(type) {
		case *schema.Enum:
			enums[t.Name] = true
		case *schema.InputObject:
			inputs[t.Name] = newInputObject(t)
		case *schema.Interface:
			interfaces[t.Name] = newInterface(t)
		}
	}

	// Going through objects and creating resolvers
	// Need to check if we implement an interface
	for _, o := range s.Objects {
		r := newResolver(o)
		w.WriteString(r.Struct())
		for _, f := range r.funcs {
			w.WriteString(f)
		}
	}

	// Go through input types and create structs
	for _, t := range inputs {
		w.WriteString(t.String())
	}

	// interfaces
	for _, i := range interfaces {
		// Print the interface
		w.WriteString(i.String())

		// Create a resolver for interface
		w.WriteString(fmt.Sprintf("\ntype %sResolver struct{\n\t%s\n}\n", i.name, i.name))

		// Create translations to structs
		for _, name := range i.implementedBy {
			w.WriteString(fmt.Sprintf(
				`
				func (r *%sResolver) To%s() (%s, bool) {
				c, ok := r.%s.(%s)
				return c, ok
			}`,
				i.name, name, translate(name), i.name, translate(name)))
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
	existed := exists(out)
	if !existed {
		newFile = true
		stub, err = os.Create(out)
		if err != nil {
			log.Fatalf("Creation of file %s went badly: %s\n", out, err)
		}
		return
	}
	stub, err = os.OpenFile(out, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Opening file %s went badly: %s\n ", out, err)
	}

	if existed {
		parseFile(out)
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

func parseFile(fname string) {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		trimmed := strings.Trim(line, " ")
		if strings.Index(trimmed, "func") == 0 {
			fmt.Println("FUNCTION: ", line)
		} else if strings.Index(trimmed, "type") == 0 {
			fmt.Println("TYPE:", line)
		}
	}
}
