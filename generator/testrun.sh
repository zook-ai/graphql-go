cd generator
rm testfiles/b.go
go run *.go testfiles/schema.gql testfiles/b.go
cd testfiles/
go run b.go
cd ..