package exec

import (
	"context"
	"reflect"

	"github.com/zook-ai/graphql-go/errors"
	"github.com/zook-ai/graphql-go/internal/query"
	"github.com/zook-ai/graphql-go/internal/schema"
	"github.com/zook-ai/graphql-go/introspection"
)

var schemaExec iExec
var typeExec iExec

func init() {
	b := newExecBuilder(schema.Meta)

	if err := b.assignExec(&schemaExec, schema.Meta.Types["__Schema"], reflect.TypeOf(&introspection.Schema{})); err != nil {
		panic(err)
	}

	if err := b.assignExec(&typeExec, schema.Meta.Types["__Type"], reflect.TypeOf(&introspection.Type{})); err != nil {
		panic(err)
	}

	if err := b.finish(); err != nil {
		panic(err)
	}
}

func IntrospectSchema(s *schema.Schema) (interface{}, error) {
	r := &request{
		schema:  s,
		doc:     introspectionQuery,
		limiter: make(semaphore, 10),
	}
	return introspectSchema(context.Background(), r, introspectionQuery.Operations["IntrospectionQuery"].SelSet), nil
}

func introspectSchema(ctx context.Context, r *request, selSet *query.SelectionSet) interface{} {
	return schemaExec.exec(ctx, r, selSet, reflect.ValueOf(introspection.WrapSchema(r.schema)), false)
}

func introspectType(ctx context.Context, r *request, name string, selSet *query.SelectionSet) interface{} {
	t, ok := r.schema.Types[name]
	if !ok {
		return nil
	}
	return typeExec.exec(ctx, r, selSet, reflect.ValueOf(introspection.WrapType(t)), false)
}

var introspectionQuery *query.Document

func init() {
	var err *errors.QueryError
	introspectionQuery, err = query.Parse(introspectionQuerySrc, schema.Meta.Resolve)
	if err != nil {
		panic(err)
	}
}

var introspectionQuerySrc = `
  query IntrospectionQuery {
    __schema {
      queryType { name }
      mutationType { name }
      subscriptionType { name }
      types {
        ...FullType
      }
      directives {
        name
        description
        locations
        args {
          ...InputValue
        }
      }
    }
  }
  fragment FullType on __Type {
    kind
    name
    description
    fields(includeDeprecated: true) {
      name
      description
      args {
        ...InputValue
      }
      type {
        ...TypeRef
      }
      isDeprecated
      deprecationReason
    }
    inputFields {
      ...InputValue
    }
    interfaces {
      ...TypeRef
    }
    enumValues(includeDeprecated: true) {
      name
      description
      isDeprecated
      deprecationReason
    }
    possibleTypes {
      ...TypeRef
    }
  }
  fragment InputValue on __InputValue {
    name
    description
    type { ...TypeRef }
    defaultValue
  }
  fragment TypeRef on __Type {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
                ofType {
                  kind
                  name
                }
              }
            }
          }
        }
      }
    }
  }
`
