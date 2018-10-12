package dag

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/hashstructure"
)

// Edge represents an edge in the graph, with a source and target vertex.
type Edge interface {
	Source() Vertex
	Target() Vertex

	Hashable
}

// BasicEdge returns an Edge implementation that simply tracks the source
// and target given as-is.
func BasicEdge(source, target Vertex) Edge {
	return &basicEdge{S: source, T: target}
}

// basicEdge is a basic implementation of Edge that has the source and
// target vertex.
type basicEdge struct {
	S, T Vertex
}

func (e *basicEdge) Hashcode() interface{} {

	rvS := reflect.ValueOf(e.S)
	if rvS.Kind() == reflect.Ptr {
		rvS = rvS.Elem()
	}
	rvT := reflect.ValueOf(e.T)
	if rvT.Kind() == reflect.Ptr {
		rvT = rvT.Elem()
	}
	switch rvS.Kind() {
	case reflect.Struct:
		s := rvS.Interface()
		t := rvS.Interface()
		hashS, _ := hashstructure.Hash(s, nil)
		hashT, _ := hashstructure.Hash(t, nil)
		return fmt.Sprintf("%x-%x", hashS, hashT)
	}
	return fmt.Sprintf("%p-%p", e.S, e.T)
}

func (e *basicEdge) Source() Vertex {
	return e.S
}

func (e *basicEdge) Target() Vertex {
	return e.T
}
