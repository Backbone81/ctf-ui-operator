package minio

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

const (
	Image          = "minio/minio:RELEASE.2025-04-22T22-12-26Z"
	tmpVolumeName  = "tmp"
	dataVolumeName = "data"
)

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

type DeploymentReconciler struct {
	utils.DefaultSubReconciler
}

func NewDeploymentReconciler(client client.Client) *DeploymentReconciler {
	return &DeploymentReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
}

func (r *DeploymentReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&appsv1.Deployment{})
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, minio *v1alpha1.Minio) (ctrl.Result, error) {
	currentSpec, err := r.getDeployment(ctx, minio)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredDeploymentSpec(minio)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentSpec == nil {
		return r.reconcileOnCreate(ctx, desiredSpec)
	}
	return r.reconcileOnUpdate(ctx, currentSpec, desiredSpec)
}

func (r *DeploymentReconciler) reconcileOnCreate(ctx context.Context, desiredSpec *appsv1.Deployment) (ctrl.Result, error) {
	if err := r.GetClient().Create(ctx, desiredSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) reconcileOnUpdate(ctx context.Context, currentSpec *appsv1.Deployment, desiredSpec *appsv1.Deployment) (ctrl.Result, error) {
	if equality.Semantic.DeepDerivative(desiredSpec.Spec, currentSpec.Spec) {
		return ctrl.Result{}, nil
	}

	currentSpec.Spec = desiredSpec.Spec
	if err := r.GetClient().Update(ctx, currentSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) getDeployment(ctx context.Context, minio *v1alpha1.Minio) (*appsv1.Deployment, error) {
	var deployment appsv1.Deployment
	if err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(minio), &deployment); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &deployment, nil
}

//nolint:funlen // We want to keep the structure of the yaml manifest.
func (r *DeploymentReconciler) getDesiredDeploymentSpec(minio *v1alpha1.Minio) (*appsv1.Deployment, error) {
	var resources corev1.ResourceRequirements
	if minio.Spec.Resources != nil {
		resources = *minio.Spec.Resources
	}
	emptyDirSize, err := resource.ParseQuantity("128Mi")
	if err != nil {
		return nil, err
	}
	result := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      minio.Name,
			Namespace: minio.Namespace,
			Labels:    minio.GetDesiredLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To[int32](1),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: ptr.To(metav1.LabelSelector{
				MatchLabels: minio.GetDesiredLabels(),
			}),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: minio.GetDesiredLabels(),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: minio.Name,
					SecurityContext: ptr.To(corev1.PodSecurityContext{
						RunAsUser:    ptr.To[int64](65534),
						RunAsNonRoot: ptr.To(true),
					}),
					Containers: []corev1.Container{
						{
							Name:  "minio",
							Image: Image,
							Args: []string{
								"server",
								"--address=:9000",
								"--console-address=:9001",
								"/mnt/data",
							},
							SecurityContext: ptr.To(corev1.SecurityContext{
								ReadOnlyRootFilesystem:   ptr.To(true),
								AllowPrivilegeEscalation: ptr.To(false),
								Privileged:               ptr.To(false),
								Capabilities: ptr.To(corev1.Capabilities{
									Drop: []corev1.Capability{
										"ALL",
									},
								}),
							}),
							Ports: []corev1.ContainerPort{
								{
									Name:          "minio",
									ContainerPort: 9000,
								},
								{
									Name:          "web",
									ContainerPort: 9001,
								},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: ptr.To(corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: minio.Name,
										},
									}),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      dataVolumeName,
									MountPath: "/mnt/data",
								},
								{
									Name:      tmpVolumeName,
									MountPath: "/tmp",
								},
							},
							Resources: resources,
							StartupProbe: ptr.To(corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: ptr.To(corev1.HTTPGetAction{
										Path: "/minio/health/live",
										Port: intstr.FromString("minio"),
									}),
								},
								TimeoutSeconds:   1,  // explicit default value required for DeepDerivative
								PeriodSeconds:    10, // explicit default value required for DeepDerivative
								SuccessThreshold: 1,  // explicit default value required for DeepDerivative
								FailureThreshold: 3,  // explicit default value required for DeepDerivative
							}),
							ReadinessProbe: ptr.To(corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: ptr.To(corev1.HTTPGetAction{
										Path: "/minio/health/live",
										Port: intstr.FromString("minio"),
									}),
								},
								TimeoutSeconds:   1,  // explicit default value required for DeepDerivative
								PeriodSeconds:    10, // explicit default value required for DeepDerivative
								SuccessThreshold: 1,  // explicit default value required for DeepDerivative
								FailureThreshold: 3,  // explicit default value required for DeepDerivative
							}),
							LivenessProbe: ptr.To(corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: ptr.To(corev1.HTTPGetAction{
										Path: "/minio/health/live",
										Port: intstr.FromString("minio"),
									}),
								},
								TimeoutSeconds:   1,  // explicit default value required for DeepDerivative
								PeriodSeconds:    10, // explicit default value required for DeepDerivative
								SuccessThreshold: 1,  // explicit default value required for DeepDerivative
								FailureThreshold: 3,  // explicit default value required for DeepDerivative
							}),
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: dataVolumeName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: ptr.To(corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: minio.Name,
								}),
							},
						},
						{
							Name: tmpVolumeName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: ptr.To(corev1.EmptyDirVolumeSource{
									Medium:    corev1.StorageMediumMemory,
									SizeLimit: ptr.To(emptyDirSize),
								}),
							},
						},
					},
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(minio, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}
