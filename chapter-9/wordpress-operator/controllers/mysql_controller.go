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
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	hostingv1alpha1 "github.com/wordpress-inc/wordpress-operator/api/v1alpha1"
)

// MYSQLReconciler reconciles a MYSQL object
type MYSQLReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hosting.wordpress.com,resources=MYSQLs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hosting.wordpress.com,resources=MYSQLs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *MYSQLReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	reqLogger := r.Log.WithValues("MYSQL", req.NamespacedName)

	// Fetch the Memcached instance
	MYSQL := &hostingv1alpha1.MYSQL{}
	err := r.Get(ctx, req.NamespacedName, MYSQL)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("MYSQL resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get MYSQL")
		return ctrl.Result{}, err
	}

	err = r.processPVC(MYSQL, reqLogger)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	err = r.processService(MYSQL, reqLogger)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	err = r.processDep(MYSQL, reqLogger)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}
	// Update the WordPress status with the pod names
	// List the pods for this WordPress's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(MYSQL.Namespace),
		client.MatchingLabels(labelsForMYSQL(MYSQL.Name)),
	}
	if err = r.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "MYSQL.Namespace", MYSQL.Namespace, "MYSQL.Name", MYSQL.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, MYSQL.Status.Nodes) {
		MYSQL.Status.Nodes = podNames
		err := r.Status().Update(context.TODO(), MYSQL)
		if err != nil {
			reqLogger.Error(err, "Failed to update MYSQL status")
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *MYSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hostingv1alpha1.MYSQL{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}

func (r *MYSQLReconciler) processService(w *hostingv1alpha1.MYSQL, reqLogger logr.Logger) error {
	service := r.serviceForMYSQL(w)
	// Set FabricOrderer instance as the owner and controller
	if err := controllerutil.SetControllerReference(w, &service, r.Scheme); err != nil {
		return err
	}
	currentService := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, currentService)
	if err != nil && errors.IsNotFound(err) {
		//Secret not exists
		reqLogger.Info("Creating a new service", "Namespace", service.Namespace, "Name", service.Name)
		err = r.Create(context.TODO(), &service)
		if err != nil {
			return err
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return err
	}
	return nil
}

func (r *MYSQLReconciler) processPVC(w *hostingv1alpha1.MYSQL, reqLogger logr.Logger) error {
	pvc := r.pvcForMYSQL(w)

	if err := controllerutil.SetControllerReference(w, &pvc, r.Scheme); err != nil {
		return err
	}

	currentPVC := &corev1.PersistentVolumeClaim{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, currentPVC)
	if err != nil && errors.IsNotFound(err) {
		//Secret not exists
		reqLogger.Info("Creating a new pvc", "Namespace", pvc.Namespace, "Name", pvc.Name)
		err = r.Create(context.TODO(), &pvc)
		if err != nil {
			return err
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get PVC")
		return err
	}
	return nil
}

func (r *MYSQLReconciler) processDep(w *hostingv1alpha1.MYSQL, reqLogger logr.Logger) error {
	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: "mysql", Namespace: w.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForMYSQL(w)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", "mysql")
			return err
		}
		// Deployment created successfully - return and requeue
		return nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return err
	}

	// Ensure the deployment size is the same as the spec
	size := w.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return err
		}
		// Spec updated - return and requeue
		return nil
	}
	return nil
}

func (r *MYSQLReconciler) serviceForMYSQL(w *hostingv1alpha1.MYSQL) corev1.Service {
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: w.Namespace,
			Name:      "wordpress-mysql",
		},
		Spec: apiv1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"app":  "wordpress",
				"tier": "mysql",
			},
			Ports: []apiv1.ServicePort{
				{
					Name:     "mysql",
					Protocol: apiv1.ProtocolTCP,
					Port:     3306,
				},
			},
		},
	}
}

func (r *MYSQLReconciler) pvcForMYSQL(m *hostingv1alpha1.MYSQL) corev1.PersistentVolumeClaim {
	return corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-pv-claim",
			Namespace: m.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: m.Spec.DataVolumeSize,
				},
			},
		},
	}
}

//deploymentForMYSQL returns a MYSQL Deployment object
func (r *MYSQLReconciler) deploymentForMYSQL(m *hostingv1alpha1.MYSQL) *appsv1.Deployment {
	ls := labelsForMYSQL(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql",
			Namespace: m.Namespace,
			Labels:    map[string]string{"app": "wordpress"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Strategy: appsv1.DeploymentStrategy{Type: "Recreate"},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "mysql",
							Image: "mysql:5.6",
							Env: []apiv1.EnvVar{
								{
									Name: "MYSQL_ROOT_PASSWORD",
									ValueFrom: &apiv1.EnvVarSource{
										SecretKeyRef: &apiv1.SecretKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "mysql-pass",
											},
											Key: "password",
										},
									},
								},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "mysql",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 3306,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "mysql-persistent-storage",
									MountPath: "/var/lib/MYSQL",
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "mysql-persistent-storage",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: "mysql-pv-claim",
								},
							},
						},
					},
				},
			},
		},
	}
	// Set WordPress instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsForWordPress returns the labels for selecting the resources
// belonging to the given WordPress CR name.
func labelsForMYSQL(name string) map[string]string {
	return map[string]string{
		"app":  "wordpress",
		"tier": "mysql",
	}
}
