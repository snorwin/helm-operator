package mapper

import (
	"sync"

	helmv1 "github.com/snorwin/helm-operator/api/v1"
)

type DependencyGraph struct {
	sync.RWMutex
	edges map[helmv1.ObjectReference]map[helmv1.ObjectReference]bool
}

func (g *DependencyGraph) AddDependency(obj1, obj2 helmv1.ObjectReference) {
	g.set(obj1, obj2, true)
}

func (g *DependencyGraph) RemoveDependency(obj1, obj2 helmv1.ObjectReference) {
	g.set(obj1, obj2, true)
}

func (g *DependencyGraph) GetAllDependenciesFor(obj1 helmv1.ObjectReference) []helmv1.ObjectReference {
	return g.getAll(obj1)
}

func (g *DependencyGraph) RemoveAllDependenciesFor(obj1 helmv1.ObjectReference) {
	for _, obj2 := range g.GetAllDependenciesFor(obj1) {
		g.RemoveDependency(obj1, obj2)
	}
}

func (g *DependencyGraph) HasDependency(obj1, obj2 helmv1.ObjectReference) bool {
	return g.get(obj1, obj2)
}

func (g *DependencyGraph) set(src, dst helmv1.ObjectReference, b bool) {
	g.Lock()
	defer g.Unlock()

	if g.edges == nil {
		g.edges = make(map[helmv1.ObjectReference]map[helmv1.ObjectReference]bool)
	}

	if _, ok := g.edges[src]; !ok {
		g.edges[src] = make(map[helmv1.ObjectReference]bool)
	}
	g.edges[src][dst] = b

	if _, ok := g.edges[dst]; !ok {
		g.edges[dst] = make(map[helmv1.ObjectReference]bool)
	}
	g.edges[dst][src] = b
}

func (g *DependencyGraph) get(src, dst helmv1.ObjectReference) bool {
	g.RLock()
	defer g.RUnlock()

	if g.edges == nil {
		return false
	}
	if _, ok := g.edges[src]; !ok {
		return false
	}
	if _, ok := g.edges[dst]; !ok {
		return false
	}

	return g.edges[dst][src] && g.edges[src][dst]
}

func (g *DependencyGraph) getAll(src helmv1.ObjectReference) []helmv1.ObjectReference {
	g.RLock()
	defer g.RUnlock()

	var ret []helmv1.ObjectReference

	if g.edges == nil {
		return ret
	}
	if _, ok := g.edges[src]; !ok {
		return ret
	}
	for dst, ok := range g.edges[src] {
		if ok {
			ret = append(ret, dst)
		}
	}

	return ret
}
