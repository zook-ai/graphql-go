package main

import "fmt"

// Func holds the signature for a function
type Func struct {
	name string
	recv Field
	args Args
	ret  Field
}

func newFunc(fname string, recv Field, args Args, ret Field) (f Func) {
	f.name = toPublic(fname)
	f.recv = recv
	f.args = args
	f.ret = ret
	return
}

func (f Func) String() string {
	return fmt.Sprintf("func (%s) %s(%s)%s{\n\t%s\n}", f.recv.String(), f.name, f.args.StringAsArgument(), f.ret.String(), defaultRet(f.ret.typpe))
}
