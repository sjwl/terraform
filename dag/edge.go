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
	switch rvS.Kind() {
	case reflect.Struct:
		hashS, _ := hashstructure.Hash(rvS, nil)
		hashT, _ := hashstructure.Hash(rvS, nil)
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
