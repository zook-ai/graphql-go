package main

type crayMap map[string]existMap

var exists crayMap

func (b *crayMap) hasFunc(fu Func) bool {
	c := *b
	if _, ok := c["func"]; !ok {
		return false
	}

	_, ok := c["func"][fu.Compare()]
	if !ok {
		return false
	}
	return true
}

func (b *crayMap) putFunc(fu Func) {
	c := *b
	if _, ok := c["func"]; !ok {
		c["func"] = make(existMap)
	}
	c["func"][fu.Compare()] = true
}

func (b *crayMap) hasStruct(st *Struct) bool {
	c := *b
	if _, ok := c["struct"]; !ok {
		return false
	}
	_, ok := c["struct"][st.String()]
	if !ok {
		return false
	}
	return true
}

func (b *crayMap) putStruct(fu Struct) {
	c := *b
	if _, ok := c["struct"]; !ok {
		c["struct"] = make(existMap)
	}
	c["struct"][fu.String()] = true
}

func (b *crayMap) hasInterface(in *Interface) bool {
	c := *b
	if _, ok := c["interface"]; !ok {
		return false
	}
	_, ok := c["interface"][in.String()]
	if !ok {
		return false
	}
	return true
}

func (b *crayMap) putInterface(fu Interface) {
	c := *b
	if _, ok := c["interface"]; !ok {
		c["interface"] = make(existMap)
	}
	c["interface"][fu.String()] = true
}
