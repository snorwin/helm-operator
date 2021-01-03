package mapper

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	helmv1 "github.com/snorwin/helm-operator/api/v1"
)

type Mapper struct {
	Graph DependencyGraph
}

func (m *Mapper) ReleaseMapFunc(obj client.Object) []reconcile.Request {
	var ret []reconcile.Request

	ref := helmv1.ObjectReference{
		APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
		Namespace:  obj.GetNamespace(),
		Name:       obj.GetName(),
	}

	for _, dependency := range m.Graph.GetAllDependenciesFor(ref) {
		if dependency.APIVersion != helmv1.GroupVersion.String() || dependency.Kind != "Release" {
			continue
		}
		ret = append(ret, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: dependency.Namespace,
				Name:      dependency.Name,
			},
		})
	}

	return ret
}
