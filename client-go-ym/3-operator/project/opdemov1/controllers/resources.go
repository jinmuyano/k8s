/*
 * @Date: 2022-11-22 11:04:46
 * @LastEditors: zhangwenqiang
 * @LastEditTime: 2022-11-23 11:05:35
 * @FilePath: /opdemo/controllers/resources.go
 */
package controllers

import (
	// "github.com/cnych/opdemo/v2/v1beta1"
	"opdemov2/api/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewDeploy(app *v1beta1.AppService) *appsv1.Deployment {
	labels := map[string]string{"myapp": app.Name}
	selector := &metav1.LabelSelector{
		MatchLabels: labels,
	}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            app.Name,
			Namespace:       app.Namespace,
			OwnerReferences: makeOwnerReference(app),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: app.Spec.Size,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: newContainers(app),
				},
			},
			Selector: selector,
		},
	}
}

func makeOwnerReference(app *v1beta1.AppService) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(app, schema.GroupVersionKind{
			Kind:    v1beta1.Kind,
			Group:   v1beta1.GroupVersion.Group,
			Version: v1beta1.GroupVersion.Version,
		}),
	}
}

func newContainers(app *v1beta1.AppService) []corev1.Container {
	var containerPorts []corev1.ContainerPort
	for _, svcPort := range app.Spec.Ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: svcPort.TargetPort.IntVal,
		})
	}
	return []corev1.Container{{
		Name:      app.Name,
		Image:     app.Spec.Image,
		Resources: app.Spec.Resources,
		Env:       app.Spec.Envs,
		Ports:     containerPorts,
	},
	}
}

func NewService(app *v1beta1.AppService) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            app.Name,
			Namespace:       app.Namespace,
			OwnerReferences: makeOwnerReference(app),
		},
		Spec: corev1.ServiceSpec{
			Ports: app.Spec.Ports,
			Type:  corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				"myapp": app.Name,
			},
		},
	}
}
