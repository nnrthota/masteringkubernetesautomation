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
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	hostingv1alpha1 "github.com/wordpress-inc/wordpress-operator/api/v1alpha1"
)

// WordPressReconciler reconciles a WordPress object
type WordPressReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hosting.wordpress.com,resources=wordpresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hosting.wordpress.com,resources=wordpresses/status,verbs=get;update;patch

func (r *WordPressReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	reqLogger := r.Log.WithValues("wordpress", req.NamespacedName)

	// Fetch the WordPress instance
	wordpress := &hostingv1alpha1.WordPress{}
	err := r.Get(ctx, req.NamespacedName, wordpress)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("WordPress resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get WordPress")
		return reconcile.Result{}, err
	}

	err = r.processSecret(wordpress, reqLogger)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	err = r.processPVC(wordpress, reqLogger)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	err = r.processService(wordpress, reqLogger)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	err = r.processDep(wordpress, reqLogger)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	// Update the WordPress status with the pod names
	// List the pods for this WordPress's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(wordpress.Namespace),
		client.MatchingLabels(labelsForWordPress(wordpress.Name)),
	}
	if err = r.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "WordPress.Namespace", wordpress.Namespace, "WordPress.Name", wordpress.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, wordpress.Status.Nodes) {
		wordpress.Status.Nodes = podNames
		err := r.Status().Update(context.TODO(), wordpress)
		if err != nil {
			reqLogger.Error(err, "Failed to update WordPress status")
			return reconcile.Result{}, err
		}
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *WordPressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hostingv1alpha1.WordPress{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}

func (r *WordPressReconciler) processSecret(w *hostingv1alpha1.WordPress, reqLogger logr.Logger) error {
	secret := r.secretForWordPress(w)
	// Set FabricOrderer instance as the owner and controller
	if err := controllerutil.SetControllerReference(w, &secret, r.Scheme); err != nil {
		return err
	}
	currentSecret := &corev1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, currentSecret)
	if err != nil && errors.IsNotFound(err) {
		//Secret not exists
		reqLogger.Info("Creating a new secret", "Namespace", secret.Namespace, "Name", secret.Name)
		err = r.Create(context.TODO(), &secret)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Updating secrets
	eq := reflect.DeepEqual(secret.Data, currentSecret.Data)
	if !eq {
		reqLogger.Info("Updating secret", "Namespace", secret.Namespace, "Name", secret.Name)
		err = r.Update(context.TODO(), &secret)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *WordPressReconciler) processService(w *hostingv1alpha1.WordPress, reqLogger logr.Logger) error {
	service := r.serviceForWordPress(w)
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

func (r *WordPressReconciler) processPVC(w *hostingv1alpha1.WordPress, reqLogger logr.Logger) error {
	pvc := r.pvcForWordPress(w)

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

func (r *WordPressReconciler) processDep(w *hostingv1alpha1.WordPress, reqLogger logr.Logger) error {
	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: w.Name, Namespace: w.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForWordPress(w)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
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

func (r *WordPressReconciler) serviceForWordPress(w *hostingv1alpha1.WordPress) corev1.Service {
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: w.Namespace,
			Name:      "wordpress",
		},
		Spec: apiv1.ServiceSpec{
			Type: "ClusterIP",
			Selector: map[string]string{
				"app":  "wordpress",
				"tier": "frontend",
			},
			Ports: []apiv1.ServicePort{
				{
					Name:     "web",
					Protocol: apiv1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						IntVal: 80,
					},
				},
			},
		},
	}
}

func (r *WordPressReconciler) secretForWordPress(w *hostingv1alpha1.WordPress) corev1.Secret {
	return corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-pass",
			Namespace: w.Namespace,
		},
		Data: map[string][]byte{
			"password": []byte("yff37dqi893kdu"),
		},
	}
}

func (r *WordPressReconciler) pvcForWordPress(m *hostingv1alpha1.WordPress) corev1.PersistentVolumeClaim {
	return corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wp-pv-claim",
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

//deploymentForWordPress returns a WordPress Deployment object
func (r *WordPressReconciler) deploymentForWordPress(m *hostingv1alpha1.WordPress) *appsv1.Deployment {
	ls := labelsForWordPress(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
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
							Name:  "wordpress",
							Image: "wordpress:4.8-apache",
							Env: []apiv1.EnvVar{
								{
									Name:  "WORDPRESS_DB_HOST",
									Value: m.Spec.WordPressDBHost,
								},
								{
									Name: "WORDPRESS_DB_PASSWORD",
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
									Name:          "wordpress",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "wordpress-persistent-storage",
									MountPath: "/var/www/html",
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "wordpress-persistent-storage",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: "wp-pv-claim",
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
func labelsForWordPress(name string) map[string]string {
	return map[string]string{
		"app":  "wordpress",
		"tier": "mysql",
	}
}
