package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"

	"reflect"

	"github.com/zook-ai/graphql-go/internal/schema"
)

// [] support scalars
// [] support mutation
// [] support subscription
// [X] support interfaces
// [X] support input Objects
// [X] support enums
// [X] support arrays
// [X] support multiple inputs
// [] support unions
// [X] Don't overwrite if the file exists.
// [X] parse output files and only add missing methods
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
	} else {
		// return
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
			w.WriteString(f.String())
		}
	}

	// Go through input types and create structs
	for _, t := range inputs {
		w.WriteString(t.String())
	}

	// interfaces
	for _, i := range interfaces {
		// Print the interface
		if !exists.hasInterface(i) {
			w.WriteString(i.String())
		}

		// Create a resolver for interface
		str := Struct{name: i.name + "Resolver", fields: []Field{Field{typpe: i.name}}}
		if !exists.hasStruct(&str) {
			w.WriteString(str.String())
		}

		// Create translations to structs
		for _, name := range i.implementedBy {
			fun := Func{name: "To" + name, recv: Field{"r", "*" + i.name + "Resolver"}}
			fun.body = fmt.Sprintf("c, ok := r.%s.(%s)\n\treturn c, ok", i.name, translate(name))
			fun.ret = Args{Arg{typpe: translate(name)}, Arg{typpe: "bool"}}
			if !exists.hasFunc(fun) {
				w.WriteString(fun.String())
			}
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
	existed := fileExists(out)
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

func fileExists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}

func parseFile(fname string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fname, nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	exists = make(map[string]existMap)

	for _, d := range f.Decls {
		// fmt.Println(reflect.TypeOf(d))
		switch d := d.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					switch t := s.Type.(type) {
					case *ast.StructType:
						var str Struct
						str.name = s.Name.Name
						for _, field := range t.Fields.List {
							var f Field
							f.typpe = getType(field.Type)
							for _, name := range field.Names {
								f.name = name.Name
								str.fields = append(str.fields, f)
							}
							if len(field.Names) == 0 {
								str.fields = append(str.fields, f)
							}
						}
						exists.putStruct(str)

					case *ast.InterfaceType:
						var i Interface
						i.name = s.Name.Name
						for _, m := range t.Methods.List {
							m2 := method{name: m.Names[0].Name}
							switch t := m.Type.(type) {
							case *ast.FuncType:
								for _, param := range t.Params.List {
									arg := Arg{typpe: getType(param.Type)}
									if len(param.Names) > 0 {
										arg.name = param.Names[0].Name
									}
									m2.args = append(m2.args, arg)
								}
								for _, ret := range t.Results.List {
									m2.typpe = getType(ret.Type)
								}
							}
							i.methods = append(i.methods, m2)
						}
						exists.putInterface(i)
					}
				}
			}
		case *ast.FuncDecl:
			var name = d.Name.Name

			var args Args
			if d.Type.Params != nil {
				for _, params := range d.Type.Params.List {
					switch s := params.Type.(type) {
					case *ast.StarExpr:
						switch p := s.X.(type) {
						case *ast.StructType:
							for _, f := range p.Fields.List {
								t := getType(f.Type)
								for _, n := range f.Names {
									args = append(args, Arg{n.Name, t})
								}
							}
						}
					}
				}
			}
			var ret Args
			if d.Type.Results != nil {
				for _, res := range d.Type.Results.List {
					r := Arg{typpe: getType(res.Type)}
					for _, name := range res.Names {
						r.name = name.Name
						ret = append(ret, r)
					}
					if len(res.Names) == 0 {
						ret = append(ret, r)
					}
				}
			}
			var recv Field
			if d.Recv != nil {
				for _, r := range d.Recv.List {
					recv.typpe = getType(r.Type)
					recv.name = r.Names[0].Name
				}
			}
			f := newFunc(name, recv, args, ret)
			exists.putFunc(f)
		}
	}
}

func getType(expr ast.Expr) string {
	switch r := expr.(type) {
	case *ast.StarExpr:
		return "*" + fmt.Sprint(r.X)
	case *ast.ArrayType:
		return "[]" + getType(r.Elt)
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", r.X, r.Sel)
	case *ast.Ident:
		return r.Name
	default:
		fmt.Println("YOU MISSED THIS:", reflect.TypeOf(r))
	}
	return ""
}
