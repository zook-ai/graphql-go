package schema

import (
	"fmt"
	"strings"
	"text/scanner"

	"github.com/zook-ai/graphql-go/errors"
	"github.com/zook-ai/graphql-go/internal/common"
	"github.com/zook-ai/graphql-go/internal/lexer"
)

type Schema struct {
	EntryPoints map[string]NamedType
	Types       map[string]NamedType

	EntryPointNames map[string]string
	Objects         []*Object
	Unions          []*Union
}

func (s *Schema) Resolve(name string) common.Type {
	return s.Types[name]
}

type NamedType interface {
	common.Type
	TypeName() string
	Description() string
}

type Scalar struct {
	Name string
	Desc string
}

type Object struct {
	Name       string
	Interfaces []*Interface
	Fields     map[string]*Field
	FieldOrder []string
	Desc       string

	interfaceNames []string
}

type Interface struct {
	Name          string
	PossibleTypes []*Object
	Fields        map[string]*Field
	FieldOrder    []string
	Desc          string
}

type Union struct {
	Name          string
	PossibleTypes []*Object
	Desc          string

	typeNames []string
}

type Enum struct {
	Name   string
	Values []*EnumValue
	Desc   string
}

type EnumValue struct {
	Name string
	Desc string
}

type InputObject struct {
	Name string
	Desc string
	common.InputMap
}

func (*Scalar) Kind() string      { return "SCALAR" }
func (*Object) Kind() string      { return "OBJECT" }
func (*Interface) Kind() string   { return "INTERFACE" }
func (*Union) Kind() string       { return "UNION" }
func (*Enum) Kind() string        { return "ENUM" }
func (*InputObject) Kind() string { return "INPUT_OBJECT" }

func (t *Scalar) String() string      { return t.Name }
func (t *Object) String() string      { return t.Name }
func (t *Interface) String() string   { return t.Name }
func (t *Union) String() string       { return t.Name }
func (t *Enum) String() string        { return t.Name }
func (t *InputObject) String() string { return t.Name }

func (t *Scalar) TypeName() string      { return t.Name }
func (t *Object) TypeName() string      { return t.Name }
func (t *Interface) TypeName() string   { return t.Name }
func (t *Union) TypeName() string       { return t.Name }
func (t *Enum) TypeName() string        { return t.Name }
func (t *InputObject) TypeName() string { return t.Name }

func (t *Scalar) Description() string      { return t.Desc }
func (t *Object) Description() string      { return t.Desc }
func (t *Interface) Description() string   { return t.Desc }
func (t *Union) Description() string       { return t.Desc }
func (t *Enum) Description() string        { return t.Desc }
func (t *InputObject) Description() string { return t.Desc }

type Field struct {
	Name string
	Args common.InputMap
	Type common.Type
	Desc string
}

func New() *Schema {
	s := &Schema{
		EntryPointNames: make(map[string]string),
		Types:           make(map[string]NamedType),
	}
	for n, t := range Meta.Types {
		s.Types[n] = t
	}
	return s
}

func (s *Schema) Parse(schemaString string) error {
	sc := &scanner.Scanner{
		Mode: scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanStrings,
	}
	sc.Init(strings.NewReader(schemaString))

	l := lexer.New(sc)
	err := l.CatchSyntaxError(func() {
		parseSchema(s, l)
	})
	if err != nil {
		return err
	}

	for _, t := range s.Types {
		if err := resolveNamedType(s, t); err != nil {
			return err
		}
	}

	s.EntryPoints = make(map[string]NamedType)
	for key, name := range s.EntryPointNames {
		t, ok := s.Types[name]
		if !ok {
			if !ok {
				return errors.Errorf("type %q not found", name)
			}
		}
		s.EntryPoints[key] = t
	}

	for _, obj := range s.Objects {
		obj.Interfaces = make([]*Interface, len(obj.interfaceNames))
		for i, intfName := range obj.interfaceNames {
			t, ok := s.Types[intfName]
			if !ok {
				return errors.Errorf("interface %q not found", intfName)
			}
			intf, ok := t.(*Interface)
			if !ok {
				return errors.Errorf("type %q is not an interface", intfName)
			}
			obj.Interfaces[i] = intf
			intf.PossibleTypes = append(intf.PossibleTypes, obj)
		}
	}

	for _, union := range s.Unions {
		union.PossibleTypes = make([]*Object, len(union.typeNames))
		for i, name := range union.typeNames {
			t, ok := s.Types[name]
			if !ok {
				return errors.Errorf("object type %q not found", name)
			}
			obj, ok := t.(*Object)
			if !ok {
				return errors.Errorf("type %q is not an object", name)
			}
			union.PossibleTypes[i] = obj
		}
	}

	return nil
}

func resolveNamedType(s *Schema, t NamedType) error {
	switch t := t.(type) {
	case *Object:
		for _, f := range t.Fields {
			if err := resolveField(s, f); err != nil {
				return err
			}
		}
	case *Interface:
		for _, f := range t.Fields {
			if err := resolveField(s, f); err != nil {
				return err
			}
		}
	case *InputObject:
		if err := resolveInputObject(s, &t.InputMap); err != nil {
			return err
		}
	}
	return nil
}

func resolveField(s *Schema, f *Field) error {
	t, err := common.ResolveType(f.Type, s.Resolve)
	if err != nil {
		return err
	}
	f.Type = t
	return resolveInputObject(s, &f.Args)
}

func resolveInputObject(s *Schema, io *common.InputMap) error {
	for _, f := range io.Fields {
		t, err := common.ResolveType(f.Type, s.Resolve)
		if err != nil {
			return err
		}
		f.Type = t
	}
	return nil
}

func parseSchema(s *Schema, l *lexer.Lexer) {
	for l.Peek() != scanner.EOF {
		desc := l.DescComment()
		switch x := l.ConsumeIdent(); x {
		case "schema":
			l.ConsumeToken('{')
			for l.Peek() != '}' {
				name := l.ConsumeIdent()
				l.ConsumeToken(':')
				typ := l.ConsumeIdent()
				s.EntryPointNames[name] = typ
			}
			l.ConsumeToken('}')
		case "type":
			obj := parseObjectDecl(l)
			obj.Desc = desc
			s.Types[obj.Name] = obj
			s.Objects = append(s.Objects, obj)
		case "interface":
			intf := parseInterfaceDecl(l)
			intf.Desc = desc
			s.Types[intf.Name] = intf
		case "union":
			union := parseUnionDecl(l)
			union.Desc = desc
			s.Types[union.Name] = union
			s.Unions = append(s.Unions, union)
		case "enum":
			enum := parseEnumDecl(l)
			enum.Desc = desc
			s.Types[enum.Name] = enum
		case "input":
			input := parseInputDecl(l)
			input.Desc = desc
			s.Types[input.Name] = input
		case "scalar":
			name := l.ConsumeIdent()
			s.Types[name] = &Scalar{Name: name, Desc: desc}
		default:
			l.SyntaxError(fmt.Sprintf(`unexpected %q, expecting "schema", "type", "enum", "interface", "union", "input" or "scalar"`, x))
		}
	}
}

func parseObjectDecl(l *lexer.Lexer) *Object {
	o := &Object{}
	o.Name = l.ConsumeIdent()
	if l.Peek() == scanner.Ident {
		l.ConsumeKeyword("implements")
		for {
			o.interfaceNames = append(o.interfaceNames, l.ConsumeIdent())
			if l.Peek() == '{' {
				break
			}
		}
	}
	l.ConsumeToken('{')
	o.Fields, o.FieldOrder = parseFields(l)
	l.ConsumeToken('}')
	return o
}

func parseInterfaceDecl(l *lexer.Lexer) *Interface {
	i := &Interface{}
	i.Name = l.ConsumeIdent()
	l.ConsumeToken('{')
	i.Fields, i.FieldOrder = parseFields(l)
	l.ConsumeToken('}')
	return i
}

func parseUnionDecl(l *lexer.Lexer) *Union {
	union := &Union{}
	union.Name = l.ConsumeIdent()
	l.ConsumeToken('=')
	union.typeNames = []string{l.ConsumeIdent()}
	for l.Peek() == '|' {
		l.ConsumeToken('|')
		union.typeNames = append(union.typeNames, l.ConsumeIdent())
	}
	return union
}

func parseInputDecl(l *lexer.Lexer) *InputObject {
	i := &InputObject{}
	i.Fields = make(map[string]*common.InputValue)
	i.Name = l.ConsumeIdent()
	l.ConsumeToken('{')
	for l.Peek() != '}' {
		v := common.ParseInputValue(l)
		i.Fields[v.Name] = v
		i.FieldOrder = append(i.FieldOrder, v.Name)
	}
	l.ConsumeToken('}')
	return i
}

func parseEnumDecl(l *lexer.Lexer) *Enum {
	enum := &Enum{}
	enum.Name = l.ConsumeIdent()
	l.ConsumeToken('{')
	for l.Peek() != '}' {
		v := &EnumValue{}
		v.Desc = l.DescComment()
		v.Name = l.ConsumeIdent()
		enum.Values = append(enum.Values, v)
	}
	l.ConsumeToken('}')
	return enum
}

func parseFields(l *lexer.Lexer) (map[string]*Field, []string) {
	fields := make(map[string]*Field)
	var fieldOrder []string
	for l.Peek() != '}' {
		f := &Field{}
		f.Desc = l.DescComment()
		f.Name = l.ConsumeIdent()
		if l.Peek() == '(' {
			f.Args.Fields = make(map[string]*common.InputValue)
			l.ConsumeToken('(')
			for l.Peek() != ')' {
				v := common.ParseInputValue(l)
				f.Args.Fields[v.Name] = v
				f.Args.FieldOrder = append(f.Args.FieldOrder, v.Name)
			}
			l.ConsumeToken(')')
		}
		l.ConsumeToken(':')
		f.Type = common.ParseType(l)
		fields[f.Name] = f
		fieldOrder = append(fieldOrder, f.Name)
	}
	return fields, fieldOrder
}
