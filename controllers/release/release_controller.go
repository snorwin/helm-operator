/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package release

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/snorwin/helm-operator/pkg/mapper"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	helmv1 "github.com/snorwin/helm-operator/api/v1"
)

const (
	// HELM_DRIVER environment variable to configure the storage backends (can be set to one of the values: configmap, secret, memory)
	envHelmDriver = "HELM_DRIVER"

	// Finalizer to uninstall the Helm Release
	finalizerName = "release.finalizers.helm.snorwin.io"
)

// Reconciler reconciles a Helm Release object
type Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	mapper mapper.Mapper
}

// +kubebuilder:rbac:groups=helm.snorwin.io,resources=releases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=helm.snorwin.io,resources=releases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=helm.snorwin.io,resources=releases/finalizers,verbs=update

// SetupWithManager register the Release Reconciler to the Manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&helmv1.Release{}).
		Watches(&source.Kind{Type: &helmv1.Chart{}}, handler.EnqueueRequestsFromMapFunc(r.mapper.ReleaseMapFunc)).
		Watches(&source.Kind{Type: &helmv1.Values{}}, handler.EnqueueRequestsFromMapFunc(r.mapper.ReleaseMapFunc)).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Ignore status updates in which case metadata.Generation does not change
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			},
		}).
		Complete(r)
}

// Reconcile installs, upgrades or uninstalls the helm chart in the Release
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("release", req.NamespacedName)

	obj := helmv1.Release{}
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		if apierrors.IsNotFound(err) {
			// return and don't requeue
			return reconcile.Result{}, nil
		}
		// error reading the object - requeue the request
		return reconcile.Result{}, err
	}

	// create a helm client
	helm := new(action.Configuration)
	err := helm.Init(cli.New().RESTClientGetter(), req.Namespace, os.Getenv(envHelmDriver), func(format string, v ...interface{}) {
		logger.V(4).Info(fmt.Sprintf(format, v...))
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	// remove all dependencies form mapper
	ref := helmv1.ObjectReference{
		APIVersion: obj.GroupVersionKind().GroupVersion().String(),
		Kind:       obj.Kind,
		Name:       obj.Name,
		Namespace:  obj.Namespace,
	}
	r.mapper.Graph.RemoveAllDependenciesFor(ref)

	// handle finalizer during deletion
	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		if contains(obj.ObjectMeta.Finalizers, finalizerName) {
			// uninstall helm chart
			uninstall := action.NewUninstall(helm)
			if _, err := uninstall.Run(req.Name); err != nil {
				return reconcile.Result{}, err
			}
			obj.ObjectMeta.Finalizers = remove(obj.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, &obj); err != nil {
				return ctrl.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// add finalizer
	if !contains(obj.ObjectMeta.Finalizers, finalizerName) {
		obj.ObjectMeta.Finalizers = add(obj.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(ctx, &obj); err != nil {
			return ctrl.Result{}, err
		}
	}

	// load helm chart and update dependency
	chart, err := r.loadHelmChart(ctx, obj.Spec.ChartRef)
	if err != nil {
		return reconcile.Result{}, err
	}
	r.mapper.Graph.AddDependency(ref, obj.Spec.ChartRef)

	// load values and update dependencies
	values := chart.Values
	if values == nil {
		values = chartutil.Values{}
	}
	for i := range obj.Spec.ValuesRefs {
		tmp, err := r.loadHelmValues(ctx, obj.Spec.ValuesRefs[i])
		if err != nil {
			return reconcile.Result{}, err
		}
		// merge values
		for k, v := range tmp {
			values[k] = v
		}
		r.mapper.Graph.AddDependency(ref, obj.Spec.ValuesRefs[i])
	}

	// upgrade or install helm chart
	response, err := action.NewStatus(helm).Run(req.Name)
	if err == driver.ErrReleaseNotFound {
		install := action.NewInstall(helm)
		install.ReleaseName = req.Name
		install.Namespace = req.Namespace
		response, err = install.Run(chart, values)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err == nil {
		response, err = action.NewUpgrade(helm).Run(req.Name, chart, values)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		return reconcile.Result{}, err
	}

	// update status
	if response != nil {
		obj.Status = helmv1.ReleaseStatus{
			FirstDeployedTime: metav1.NewTime(response.Info.FirstDeployed.Time),
			LastDeployedTime:  metav1.NewTime(response.Info.LastDeployed.Time),
			Description:       response.Info.Description,
			Status:            response.Info.Status.String(),
			Notes:             response.Info.Notes,
			Version:           response.Version,
		}
	}

	if err = r.Status().Update(ctx, &obj); err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// loadHelmChart get referenced Helm Chart object and create a chart.Chart from it
func (r *Reconciler) loadHelmChart(ctx context.Context, ref helmv1.ObjectReference) (*chart.Chart, error) {
	if ref.APIVersion != helmv1.GroupVersion.String() || ref.Kind != "Chart" {
		return nil, errors.New("invalid APIVersion and Kind for Chart reference")
	}

	obj := helmv1.Chart{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &obj); err != nil {
		return nil, err
	}

	return loader.LoadFiles(obj.Spec.Files.Slice())
}

// loadHelmValues get the referenced Helm Values object and create a chartutil.Values from it
func (r *Reconciler) loadHelmValues(ctx context.Context, ref helmv1.ObjectReference) (chartutil.Values, error) {
	switch ref.APIVersion {
	case helmv1.GroupVersion.String():
		switch ref.Kind {
		case "Values":
			obj := helmv1.Values{}
			if err := r.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &obj); err != nil {
				return nil, err
			}
			return chartutil.ReadValues([]byte(obj.Spec.File.Data))
		default:
			return chartutil.Values{}, errors.New("invalid APIVersion and Kind for Values reference")
		}
	default:
		return chartutil.Values{}, errors.New("invalid APIVersion and Kind for Values reference")
	}
}

// contains check if a string in a []string exists
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// remove a string from a []string if it exist
func remove(slice []string, str string) []string {
	for i, v := range slice {
		if v == str {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// add a string to a []string if it not exist
func add(slice []string, str string) []string {
	for _, v := range slice {
		if v == str {
			return slice
		}
	}
	return append(slice, str)
}
