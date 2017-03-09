package main

type crayMap map[string]existMap

var exists crayMap

func (c crayMap) hasFunc(fu Func) bool {
	if _, ok := c["func"]; !ok {
		return false
	}
	_, ok := c["func"][fu.String()]
	if !ok {
		return false
	}
	return true
}
