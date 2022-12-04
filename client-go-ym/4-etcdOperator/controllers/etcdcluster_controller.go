/*
Copyright 2022 cnych.

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
	"strconv"

	"github.com/cnych/etcd-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	etcdv1alpha1 "github.com/cnych/etcd-operator/api/v1alpha1"
)

var (
	EtcdClusterLabelKey       = "etcd.ydzs.io/cluster"
	EtcdClusterCommonLabelKey = "app"
	EtcdDataDirName           = "datadir"
)

// EtcdClusterReconciler reconciles a EtcdCluster object
type EtcdClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=etcd.ydzs.io,resources=etcdclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=etcd.ydzs.io,resources=etcdclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=etcd.ydzs.io,resources=etcdclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EtcdCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *EtcdClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("etcdcluster", req.NamespacedName)

	// TODO(user): your logic here
	// 首先我们获取 EtcdCluster 实例
	var etcdCluster etcdv1alpha1.EtcdCluster
	if err := r.Get(ctx, req.NamespacedName, &etcdCluster); err != nil {
		// EtcdCluster was deleted，Ignore
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 得到 EtcdCluster 过后去创建对应的StatefulSet和Service
	// CreateOrUpdate

	// (就是观察的当前状态和期望的状态进行对比)

	// 调谐，获取到当前的一个状态，然后和我们期望的状态进行对比是不是就可以

	// CreateOrUpdate Service
	var svc corev1.Service
	svc.Name = etcdCluster.Name
	svc.Namespace = etcdCluster.Namespace
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
			//  调谐的函数必须在这里面实现，实际上就是去拼装我们的 Service
			MutateHeadlessSvc(&etcdCluster, &svc)
			return controllerutil.SetControllerReference(&etcdCluster, &svc, r.Scheme)
		})
		log.Info("CreateOrUpdate Result", "Service", or)
		return err
	}); err != nil {
		return ctrl.Result{}, err
	}

	// CreateOrUpdate StatefulSet
	var sts appsv1.StatefulSet
	sts.Name = etcdCluster.Name
	sts.Namespace = etcdCluster.Namespace
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r.Client, &sts, func() error {
			// 调谐必须在这个函数中去实现
			MutateStatefulSet(&etcdCluster, &sts)
			return controllerutil.SetControllerReference(&etcdCluster, &sts, r.Scheme)
		})
		log.Info("CreateOrUpdate Result", "StatefulSet", or)
		return err
	}); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func newContainers(cluster *v1alpha1.EtcdCluster) []corev1.Container {
	return []corev1.Container{
		{
			Name:  "etcd",
			Image: cluster.Spec.Image,
			Ports: []corev1.ContainerPort{
				{
					Name:          "peer",
					ContainerPort: 2380,
				},
				{
					Name:          "client",
					ContainerPort: 2379,
				},
			},
			Env: []corev1.EnvVar{
				{
					Name:  "INITIAL_CLUSTER_SIZE",
					Value: strconv.Itoa(int(*cluster.Spec.Size)),
				},
				{
					Name:  "SET_NAME",
					Value: cluster.Name,
				},
				{
					Name: "POD_IP",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
				},
				{
					Name: "MY_NAMESPACE",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.namespace",
						},
					},
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      EtcdDataDirName,
					MountPath: "/var/run/etcd",
				},
			},
			Command: []string{
				"/bin/sh", "-ec",
				"HOSTNAME=$(hostname)\n\n              ETCDCTL_API=3\n\n              eps() {\n                  EPS=\"\"\n                  for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do\n                      EPS=\"${EPS}${EPS:+,}http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379\"\n                  done\n                  echo ${EPS}\n              }\n\n              member_hash() {\n                  etcdctl member list | grep -w \"$HOSTNAME\" | awk '{ print $1}' | awk -F \",\" '{ print $1}'\n              }\n\n              initial_peers() {\n                  PEERS=\"\"\n                  for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do\n                    PEERS=\"${PEERS}${PEERS:+,}${SET_NAME}-${i}=http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380\"\n                  done\n                  echo ${PEERS}\n              }\n\n              # etcd-SET_ID\n              SET_ID=${HOSTNAME##*-}\n\n              # adding a new member to existing cluster (assuming all initial pods are available)\n              if [ \"${SET_ID}\" -ge ${INITIAL_CLUSTER_SIZE} ]; then\n                  # export ETCDCTL_ENDPOINTS=$(eps)\n                  # member already added?\n\n                  MEMBER_HASH=$(member_hash)\n                  if [ -n \"${MEMBER_HASH}\" ]; then\n                      # the member hash exists but for some reason etcd failed\n                      # as the datadir has not be created, we can remove the member\n                      # and retrieve new hash\n                      echo \"Remove member ${MEMBER_HASH}\"\n                      etcdctl --endpoints=$(eps) member remove ${MEMBER_HASH}\n                  fi\n\n                  echo \"Adding new member\"\n\n                  etcdctl member --endpoints=$(eps) add ${HOSTNAME} --peer-urls=http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380 | grep \"^ETCD_\" > /var/run/etcd/new_member_envs\n\n                  if [ $? -ne 0 ]; then\n                      echo \"member add ${HOSTNAME} error.\"\n                      rm -f /var/run/etcd/new_member_envs\n                      exit 1\n                  fi\n\n                  echo \"==> Loading env vars of existing cluster...\"\n                  sed -ie \"s/^/export /\" /var/run/etcd/new_member_envs\n                  cat /var/run/etcd/new_member_envs\n                  . /var/run/etcd/new_member_envs\n\n                  exec etcd --listen-peer-urls http://${POD_IP}:2380 \\\n                      --listen-client-urls http://${POD_IP}:2379,http://127.0.0.1:2379 \\\n                      --advertise-client-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379 \\\n                      --data-dir /var/run/etcd/default.etcd\n              fi\n\n              for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do\n                  while true; do\n                      echo \"Waiting for ${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local to come up\"\n                      ping -W 1 -c 1 ${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local > /dev/null && break\n                      sleep 1s\n                  done\n              done\n\n              echo \"join member ${HOSTNAME}\"\n              # join member\n              exec etcd --name ${HOSTNAME} \\\n                  --initial-advertise-peer-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380 \\\n                  --listen-peer-urls http://${POD_IP}:2380 \\\n                  --listen-client-urls http://${POD_IP}:2379,http://127.0.0.1:2379 \\\n                  --advertise-client-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379 \\\n                  --initial-cluster-token etcd-cluster-1 \\\n                  --data-dir /var/run/etcd/default.etcd \\\n                  --initial-cluster $(initial_peers) \\\n                  --initial-cluster-state new",
			},
			Lifecycle: &corev1.Lifecycle{
				PreStop: &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"/bin/sh", "-ec",
							"HOSTNAME=$(hostname)\n\n                    member_hash() {\n                        etcdctl member list | grep -w \"$HOSTNAME\" | awk '{ print $1}' | awk -F \",\" '{ print $1}'\n                    }\n\n                    eps() {\n                        EPS=\"\"\n                        for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do\n                            EPS=\"${EPS}${EPS:+,}http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379\"\n                        done\n                        echo ${EPS}\n                    }\n\n                    export ETCDCTL_ENDPOINTS=$(eps)\n                    SET_ID=${HOSTNAME##*-}\n\n                    # Removing member from cluster\n                    if [ \"${SET_ID}\" -ge ${INITIAL_CLUSTER_SIZE} ]; then\n                        echo \"Removing ${HOSTNAME} from etcd cluster\"\n                        etcdctl member remove $(member_hash)\n                        if [ $? -eq 0 ]; then\n                            # Remove everything otherwise the cluster will no longer scale-up\n                            rm -rf /var/run/etcd/*\n                        fi\n                    fi",
						},
					},
				},
			},
		},
	}
}

func MutateHeadlessSvc(cluster *v1alpha1.EtcdCluster, svc *corev1.Service) {
	svc.Labels = map[string]string{
		EtcdClusterCommonLabelKey: "etcd",
	}
	svc.Spec = corev1.ServiceSpec{
		ClusterIP: corev1.ClusterIPNone,
		Selector: map[string]string{
			EtcdClusterLabelKey: cluster.Name,
		},
		Ports: []corev1.ServicePort{
			{
				Name: "peer",
				Port: 2380,
			},
			{
				Name: "client",
				Port: 2379,
			},
		},
	}
}

func MutateStatefulSet(cluster *v1alpha1.EtcdCluster, sts *appsv1.StatefulSet) {
	sts.Labels = map[string]string{
		EtcdClusterCommonLabelKey: "etcd",
	}
	sts.Spec = appsv1.StatefulSetSpec{
		Replicas:    cluster.Spec.Size,
		ServiceName: cluster.Name,
		Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
			EtcdClusterLabelKey: cluster.Name,
		}},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					EtcdClusterLabelKey:       cluster.Name,
					EtcdClusterCommonLabelKey: "etcd",
				},
			},
			Spec: corev1.PodSpec{
				Containers: newContainers(cluster),
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: EtcdDataDirName,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *EtcdClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&etcdv1alpha1.EtcdCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
