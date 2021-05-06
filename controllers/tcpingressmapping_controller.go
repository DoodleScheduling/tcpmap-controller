/*


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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1beta1 "github.com/DoodleScheduling/k8stcpmap-controller/api/v1beta1"
)

const (
	serviceIndex = ".metadata.service"
)

var (
	ErrPortNotFound = errors.New("Port not found")
)

type TCPIngressMappingReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	TCPConfigMap    string
	FrontendService string
}

type TCPIngressMappingReconcilerOptions struct {
	MaxConcurrentReconciles int
}

// SetupWithManager adding controllers
func (r *TCPIngressMappingReconciler) SetupWithManager(mgr ctrl.Manager, opts TCPIngressMappingReconcilerOptions) error {
	// Index the Reqeusttcpmaps by the Service references they point at
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &v1beta1.TCPIngressMapping{}, serviceIndex,
		func(o client.Object) []string {
			vb := o.(*v1beta1.TCPIngressMapping)
			r.Log.Info(fmt.Sprintf("%s/%s", vb.GetNamespace(), vb.Spec.BackendService.Name))
			return []string{
				fmt.Sprintf("%s/%s", vb.GetNamespace(), vb.Spec.BackendService.Name),
			}
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.TCPIngressMapping{}).
		Watches(
			&source.Kind{Type: &v1.Service{}},
			handler.EnqueueRequestsFromMapFunc(r.requestsForServiceChange),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: opts.MaxConcurrentReconciles}).
		Complete(r)
}

func (r *TCPIngressMappingReconciler) requestsForServiceChange(o client.Object) []reconcile.Request {
	s, ok := o.(*v1.Service)
	if !ok {
		panic(fmt.Sprintf("expected a Service, got %T", o))
	}

	ctx := context.Background()
	var list v1beta1.TCPIngressMappingList
	if err := r.List(ctx, &list, client.MatchingFields{
		serviceIndex: objectKey(s).String(),
	}); err != nil {
		return nil
	}

	var reqs []reconcile.Request
	for _, i := range list.Items {
		r.Log.Info("referenced service from a TCPIngressMapping changed detected, reconcile TCPIngressMapping", "namespace", i.GetNamespace(), "name", i.GetName())
		reqs = append(reqs, reconcile.Request{NamespacedName: objectKey(&i)})
	}

	return reqs
}

// +kubebuilder:rbac:groups=networking.infra.doodle.com,resources=TCPIngressMappings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.infra.doodle.com,resources=TCPIngressMappings/status,verbs=get;update;patch

// Reconcile TCPIngressMappings
func (r *TCPIngressMappingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("Namespace", req.Namespace, "Name", req.NamespacedName)
	logger.Info("reconciling TCPIngressMapping")

	// Fetch the TCPIngressMapping instance
	tcpmap := v1beta1.TCPIngressMapping{}

	err := r.Client.Get(ctx, req.NamespacedName, &tcpmap)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	tcpmap, result, reconcileErr := r.reconcile(ctx, tcpmap, logger)

	// Update status after reconciliation.
	if err = r.patchStatus(ctx, &tcpmap); err != nil {
		log.Error(err, "unable to update status after reconciliation")
		return ctrl.Result{Requeue: true}, err
	}

	return result, reconcileErr
}

func (r *TCPIngressMappingReconciler) reconcile(ctx context.Context, tcpmap v1beta1.TCPIngressMapping, logger logr.Logger) (v1beta1.TCPIngressMapping, ctrl.Result, error) {
	// Lookup backend service
	backendService := v1.Service{}
	backendNS := tcpmap.GetNamespace()
	if tcpmap.Spec.BackendService.Namespace != "" {
		backendNS = tcpmap.Spec.BackendService.Namespace
	}

	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: backendNS,
		Name:      tcpmap.Spec.BackendService.Name,
	}, &backendService)

	if err != nil {
		msg := "Service not found"
		r.Recorder.Event(&tcpmap, "Normal", "info", msg)
		return v1beta1.TCPIngressMappingNotReady(tcpmap, v1beta1.BackendServiceNotFoundReason, msg), ctrl.Result{}, nil
	}

	// Lookup frontend service
	frontendService := v1.Service{}
	ns := tcpmap.GetNamespace()
	name := ""

	if r.FrontendService == "" && tcpmap.Spec.FrontendService == nil {
		msg := "Neither a frontendService nor a default one have been specified"
		r.Recorder.Event(&tcpmap, "Normal", "info", msg)
		return v1beta1.TCPIngressMappingNotReady(tcpmap, v1beta1.FrontendServiceNotFoundReason, msg), ctrl.Result{}, nil
	} else {
		if r.FrontendService == "" {
			name = tcpmap.Spec.FrontendService.Name
			if tcpmap.Spec.FrontendService.Namespace != "" {
				ns = tcpmap.Spec.FrontendService.Namespace
			}
		} else {
			parts := strings.Split(r.FrontendService, "/")
			if len(parts) == 1 {
				name = parts[0]
			} else {
				ns = parts[0]
				name = parts[1]
			}
		}
	}

	err = r.Client.Get(ctx, client.ObjectKey{
		Namespace: ns,
		Name:      name,
	}, &frontendService)

	if err != nil {
		msg := "Service not found"
		r.Recorder.Event(&tcpmap, "Normal", "info", msg)
		return v1beta1.TCPIngressMappingNotReady(tcpmap, v1beta1.FrontendServiceNotFoundReason, msg), ctrl.Result{}, nil
	}

	// Lookup configmap
	cm := v1.ConfigMap{}
	ns = tcpmap.GetNamespace()
	name = ""

	if r.TCPConfigMap == "" && tcpmap.Spec.TCPConfigMap == nil {
		msg := "Neither a ConfigMap nor a default one have been specified"
		r.Recorder.Event(&tcpmap, "Normal", "info", msg)
		return v1beta1.TCPIngressMappingNotReady(tcpmap, v1beta1.TCPConfigMapNotFoundReason, msg), ctrl.Result{}, nil
	} else {
		if r.TCPConfigMap == "" {
			name = tcpmap.Spec.TCPConfigMap.Name
			if tcpmap.Spec.TCPConfigMap.Namespace != "" {
				ns = tcpmap.Spec.BackendService.Namespace
			}
		} else {
			parts := strings.Split(r.TCPConfigMap, "/")
			if len(parts) == 1 {
				name = parts[0]
			} else {
				ns = parts[0]
				name = parts[1]
			}
		}
	}

	err = r.Client.Get(ctx, client.ObjectKey{
		Namespace: ns,
		Name:      name,
	}, &cm)

	if err != nil {
		msg := "ConfigMap not found"
		r.Recorder.Event(&tcpmap, "Normal", "info", msg)
		return v1beta1.TCPIngressMappingNotReady(tcpmap, v1beta1.TCPConfigMapNotFoundReason, msg), ctrl.Result{}, nil
	}

	var ports []int32
	for _, p := range frontendService.Spec.Ports {
		ports = append(ports, p.Port)
	}

	for k := range cm.Data {
		p, err := strconv.Atoi(k)
		if err == nil {
			ports = append(ports, int32(p))
		}
	}

	logger.Info("use port pool", "ports", ports, "elected-port", tcpmap.Status.ElectedPort)

	if tcpmap.Status.ElectedPort == 0 {
		electedPort := findPort(ports)
		logger.Info("elected free port", "port", electedPort)

		port, err := getBackendPort(backendService, tcpmap.Spec.BackendService.Port)
		if err != nil {
			msg := "Backend port not found"
			r.Recorder.Event(&tcpmap, "Normal", "info", msg)
			return v1beta1.TCPIngressMappingNotReady(tcpmap, v1beta1.BackendPortNotFoundReason, msg), ctrl.Result{}, nil
		}

		frontendService.Spec.Ports = append(frontendService.Spec.Ports, v1.ServicePort{
			Name:       fmt.Sprintf("%s-%s", backendNS, tcpmap.Spec.BackendService.Name),
			Port:       electedPort,
			TargetPort: intstr.FromInt(int(electedPort)),
			Protocol:   v1.ProtocolTCP,
		})

		if err := r.patchService(ctx, &frontendService); err != nil {
			msg := "Failed to add port to the fronted service"
			r.Recorder.Event(&tcpmap, "Normal", "info", msg)
			return v1beta1.TCPIngressMappingNotReady(tcpmap, v1beta1.FailedRegisterFrontendPortReason, msg), ctrl.Result{}, nil
		}

		cm.Data[strconv.Itoa(int(electedPort))] = fmt.Sprintf(
			"%s/%s:%d:PROXY", backendNS, tcpmap.Spec.BackendService.Name, port,
		)

		if err := r.patchConfigMap(ctx, &cm); err != nil {
			msg := "Failed to add port to the tcp configmap"
			r.Recorder.Event(&tcpmap, "Normal", "info", msg)
			return v1beta1.TCPIngressMappingNotReady(tcpmap, v1beta1.FailedRegisterConfigMapPortReason, msg), ctrl.Result{}, nil
		}

		tcpmap.Status.ElectedPort = electedPort
		msg := "Port mapping successfully registered"
		r.Recorder.Event(&tcpmap, "Normal", "info", msg)
		return v1beta1.TCPIngressMappingReady(tcpmap, v1beta1.PortReadyReason, msg), ctrl.Result{}, err
	} else {
		//TODO
	}

	return tcpmap, ctrl.Result{}, nil
}

func getBackendPort(svc v1.Service, port intstr.IntOrString) (int32, error) {
	for _, v := range svc.Spec.Ports {
		if v.Name == port.String() {
			return v.Port, nil
		}
		if int(v.Port) == port.IntValue() {
			return v.Port, nil
		}
	}

	return 0, ErrPortNotFound
}

func findPort(ports []int32) int32 {
OUTER:
	for i := int32(1025); i <= int32(65535); i++ {
		for _, e := range ports {
			if e == i {
				continue OUTER
			}
		}

		return i
	}

	return 0
}

func (r *TCPIngressMappingReconciler) patchConfigMap(ctx context.Context, cm *v1.ConfigMap) error {
	key := client.ObjectKeyFromObject(cm)
	latest := &v1.ConfigMap{}
	if err := r.Client.Get(ctx, key, latest); err != nil {
		return err
	}

	return r.Client.Patch(ctx, cm, client.MergeFrom(latest))
}

func (r *TCPIngressMappingReconciler) patchService(ctx context.Context, svc *v1.Service) error {
	key := client.ObjectKeyFromObject(svc)
	latest := &v1.Service{}
	if err := r.Client.Get(ctx, key, latest); err != nil {
		return err
	}

	return r.Client.Patch(ctx, svc, client.MergeFrom(latest))
}

func (r *TCPIngressMappingReconciler) patchStatus(ctx context.Context, tcpmap *v1beta1.TCPIngressMapping) error {
	key := client.ObjectKeyFromObject(tcpmap)
	latest := &v1beta1.TCPIngressMapping{}
	if err := r.Client.Get(ctx, key, latest); err != nil {
		return err
	}

	return r.Client.Status().Patch(ctx, tcpmap, client.MergeFrom(latest))
}

// objectKey returns client.ObjectKey for the object.
func objectKey(object metav1.Object) client.ObjectKey {
	return client.ObjectKey{
		Namespace: object.GetNamespace(),
		Name:      object.GetName(),
	}
}
