package main

import "fmt"

// Func holds the signature for a function
type Func struct {
	name string
	recv Field
	args Args
	ret  Args //TODO: should be multiple
	body string
}

func newFunc(fname string, recv Field, args Args, ret Args) (f Func) {
	f.name = toPublic(fname)
	f.recv = recv
	f.args = args
	f.ret = ret
	return
}

func (f Func) String() string {
	return fmt.Sprintf("\nfunc (%s) %s(%s) (%s) {\n\t%s\n}\n", f.recv.String(), f.name, f.args.StringAsArgument(), f.ret.StringAsReturns(), f.body)
}

// Compare is used for comparing functions signatures
func (f Func) Compare() string {
	return fmt.Sprintf("func (%s) %s(%s)%s{}", f.recv.String(), f.name, f.args.StringAsArgument(), f.ret.StringAsReturns())
}
