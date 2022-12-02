# 初始etcd

0.键值对存储

1.etcd 比较多的应用场景是用于服务注册与发现

```
服务注册与发现(Service Discovery)要解决的是分布式系统中最常见的问题之一，即在同一个分布式集群中的进程或服务如何才能找到对方并建立连接。从本质上说，服务发现就是要了解集群中是否有进程在监听 UDP 或者 TCP 端口，并且通过名字就可以进行查找和链接。

```



2.消息订阅与方法

![image-20221125212522567](/Users/admin/Library/Application Support/typora-user-images/image-20221125212522567.png)

应用中用到的一些配置信息放到 etcd 上进行集中管理。

这类场景的使用方式通常是这样：

- 应用在启动的时候主动从 etcd 获取一次配置信息，同时在 etcd 节点上注册一个 Watcher 并等待，以后每次配置有更新的时候，etcd 都会实时通知订阅者，以此达到获取最新配置信息的目的。



3.分布式通知与协调

这里说到的分布式通知与协调，与消息发布和订阅有些相似。在分布式系统中，最适用的一种组件间通信方式就是消息发布与订阅。即构建一个配置共享中心，数据提供者在这个配置中心发布消息，而消息使用者则订阅他们关心的主题，一旦主题有消息发布，就会实时通知订阅者。通过这种方式可以做到分布式系统配置的集中式管理与动态更新。

这里用到了 etcd 中的 Watcher 机制，通过注册与异步通知机制，实现分布式环境下不同系统之间的通知与协调，从而对数据变更做到实时处理。实现方式通常是这样：不同系统都在 etcd 上对同一个目录进行注册，同时设置 Watcher 观测该目录的变化（如果对子目录的变化也有需要，可以设置递归模式），当某个系统更新了 etcd 的目录，那么设置了 Watcher 的系统就会收到通知，并作出相应处理。

通过 etcd 进行低耦合的心跳检测。检测系统和被检测系统通过 etcd 上某个目录关联而非直接关联起来，这样可以大大减少系统的耦合性。

![image-20221125212855297](/Users/admin/Library/Application Support/typora-user-images/image-20221125212855297.png)





4.分布式锁

当在分布式系统中，数据只有一份（或有限制），此时需要利用锁的技术控制某一时刻修改数据的进程数。与单机模式下的锁不仅需要保证进程可见，分布式环境下还需要考虑进程与锁之间的网络问题。

分布式锁可以将标记存在内存，只是该内存不是某个进程分配的内存而是公共内存如 Redis、Memcache。至于利用数据库、文件等做锁与单机的实现是一样的，只要保证标记能互斥就行。

因为 etcd 使用 Raft 算法保持了数据的强一致性，某次操作存储到集群中的值必然是全局一致的，所以很容易实现分布式锁。锁服务有两种使用方式，一是保持独占，二是控制时序。

- 保持独占

  保持独占即所有获取锁的用户最终只有一个可以得到。etcd 为此提供了一套实现分布式锁原子操作CAS（CompareAndSwap）的 API。通过设置 prevExist 值，可以保证在多个节点同时去创建某个目录时，只有一个成功。而创建成功的用户就可以认为是获得了锁。

- 控制时序

  控制时序，即所有想要获得锁的用户都会被安排执行，但是获得锁的顺序也是全局唯一的，同时决定了执行顺序。etcd 为此也提供了一套API（自动创建有序键），对一个目录建值时指定为 POST 动作，这样 etcd 会自动在目录下生成一个当前最大的值为键，存储这个新的值（客户端编号）。同时还可以使用 API 按顺序列出所有当前目录下的键值。此时这些键的值就是客户端的时序，而这些键中存储的值可以是代表客户端的编号。





# etcd operator开发

[etcd operator 开发](https://www.notion.so/etcd-operator-a5a8084d409b408490fba8eeba0df97b)

设计yaml

```yaml
apiVersion: etcd.ydzs.io/v1alpha1
kind: EtcdCluster
metadata:
  name: demo
spec:
	size: 3  # 副本数量
	image: cnych/etcd:v3.4.13  # 镜像
```

因为其他信息都是通过脚本获取的，所以基本上我们通过 size 和 image 两个字段就可以确定一个 Etcd 集群部署的样子了，所以我们的第一个版本非常简单，只要能够写出正确的部署脚本即可，然后我们在 Operator 当中根据上面我们定义的 EtcdCluster 这个 CR 资源来组装一个 StatefulSet 和 Headless SVC 对象就可以了



## kubebuilder

安装kubebuilder

curl -LO https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.3.1/kubebuilder_2.3.1_linux_arm64.tar.gz
https://github.com/kubernetes-sigs/kubebuilder

```
git clone git@github.com:jinmuyano/kubebuilder.git
git branch -a
git checkout release-3.7
./kubebuilder/bin/kubebuilder

../install-kubebuilder/bin/kubebuilder  init --domain ydzs.io --owner cnych --repo github.com/cnych/etcd-operator

../install-kubebuilder/bin/kubebuilder create api --group etcd --version v1alpha1 --kind EtcdCluster
```







- 构建脚手架

```yaml
➜  kubebuilder init --domain ydzs.io --owner cnych --repo github.com/cnych/etcd-operator
Writing scaffold for you to edit...
Get controller runtime:
$ go get sigs.k8s.io/controller-runtime@v0.5.0
Update go.mod:
$ go mod tidy
Running make:
$ make
/Users/ych/devs/projects/go/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go build -o bin/manager main.go
Next: define a resource with:
# $ kubebuilder create api
```

定义资源api

```shell
➜  kubebuilder create api --group etcd --version v1alpha1 --kind EtcdCluster
Create Resource [y/n]
y
Create Controller [y/n]
y
Writing scaffold for you to edit...
api/v1alpha1/etcdcluster_types.go
controllers/etcdcluster_controller.go
Running make:
$ make
/Users/ych/devs/projects/go/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go build -o bin/manager main.go
```

```shell
➜  etcd-operator tree -L 2
.
├── Dockerfile
├── Makefile
├── PROJECT
├── api
│   └── v1alpha1
├── bin
│   └── manager
├── config
│   ├── certmanager
│   ├── crd
│   ├── default
│   ├── manager
│   ├── prometheus
│   ├── rbac
│   ├── samples
│   └── webhook
├── controllers
│   ├── etcdcluster_controller.go
│   └── suite_test.go
├── go.mod
├── go.sum
├── hack
│   └── boilerplate.go.txt
└── main.go

14 directories, 10 files
```





创建完成后，在项目中会新增 EtcdBackup 相关的 API 和对应的控制器，我们可以用上面设计的 CR 资源覆盖 samples 目录中的 EtcdBackup 对象。

然后可以根据上面的设计重新修改 etcdrestore_types.go 文件中的资源结构体：

```go
// api/v1alpha1/etcdcluster_types.go

// EtcdClusterSpec defines the desired state of EtcdCluster
type EtcdClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Size  *int32  `json:"size"`
	Image string  `json:"image"`
}
```



## 业务逻辑

控制器的 Reconcile 函数中来实现我们自己的业务逻辑了。
```go
// controllers/resource.go

package controllers

import (
	"strconv"

	"github.com/cnych/etcd-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	EtcdClusterLabelKey = "etcd.ydzs.io/cluster"
	EtcdClusterCommonLabelKey = "app"
	EtcdDataDirName     = "datadir"
)

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
					EtcdClusterLabelKey: cluster.Name,
					EtcdClusterCommonLabelKey: "etcd",
				},
			},
			Spec: corev1.PodSpec{
				Containers: newContainers(cluster),
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			corev1.PersistentVolumeClaim{
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

func newContainers(cluster *v1alpha1.EtcdCluster) []corev1.Container {
	return []corev1.Container{
		corev1.Container{
			Name:  "etcd",
			Image: cluster.Spec.Image,
			Ports: []corev1.ContainerPort{
				corev1.ContainerPort{
					Name:          "peer",
					ContainerPort: 2380,
				},
				corev1.ContainerPort{
					Name:          "client",
					ContainerPort: 2379,
				},
			},
			Env: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "INITIAL_CLUSTER_SIZE",
					Value: strconv.Itoa(int(*cluster.Spec.Size)),
				},
				corev1.EnvVar{
					Name:  "SET_NAME",
					Value: cluster.Name,
				},
				corev1.EnvVar{
					Name: "POD_IP",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
				},
				corev1.EnvVar{
					Name: "MY_NAMESPACE",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.namespace",
						},
					},
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				corev1.VolumeMount{
					Name:      EtcdDataDirName,
					MountPath: "/var/run/etcd",
				},
			},
			Command: []string{
				"/bin/sh", "-ec",
				"HOSTNAME=$(hostname)\n\n              ETCDCTL_API=3\n\n              eps() {\n                  EPS=\"\"\n                  for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do\n                      EPS=\"${EPS}${EPS:+,}http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379\"\n                  done\n                  echo ${EPS}\n              }\n\n              member_hash() {\n                  etcdctl member list | grep -w \"$HOSTNAME\" | awk '{ print $1}' | awk -F \",\" '{ print $1}'\n              }\n\n              initial_peers() {\n                  PEERS=\"\"\n                  for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do\n                    PEERS=\"${PEERS}${PEERS:+,}${SET_NAME}-${i}=http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380\"\n                  done\n                  echo ${PEERS}\n              }\n\n              # etcd-SET_ID\n              SET_ID=${HOSTNAME##*-}\n\n              # adding a new member to existing cluster (assuming all initial pods are available)\n              if [ \"${SET_ID}\" -ge ${INITIAL_CLUSTER_SIZE} ]; then\n                  # export ETCDCTL_ENDPOINTS=$(eps)\n                  # member already added?\n\n                  MEMBER_HASH=$(member_hash)\n                  if [ -n \"${MEMBER_HASH}\" ]; then\n                      # the member hash exists but for some reason etcd failed\n                      # as the datadir has not be created, we can remove the member\n                      # and retrieve new hash\n                      echo \"Remove member ${MEMBER_HASH}\"\n                      etcdctl --endpoints=$(eps) member remove ${MEMBER_HASH}\n                  fi\n\n                  echo \"Adding new member\"\n\n                  etcdctl member --endpoints=$(eps) add ${HOSTNAME} --peer-urls=http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380 | grep \"^ETCD_\" > /var/run/etcd/new_member_envs\n\n                  if [ $? -ne 0 ]; then\n                      echo \"member add ${HOSTNAME} error.\"\n                      rm -f /var/run/etcd/new_member_envs\n                      exit 1\n                  fi\n\n                  echo \"==> Loading env vars of existing cluster...\"\n                  sed -ie \"s/^/export /\" /var/run/etcd/new_member_envs\n                  cat /var/run/etcd/new_member_envs\n                  . /var/run/etcd/new_member_envs\n\n                  exec etcd --listen-peer-urls http://${POD_IP}:2380 \\\n                      --listen-client-urls http://${POD_IP}:2379,http://127.0.0.1:2379 \\\n                      --advertise-client-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379 \\\n                      --data-dir /var/run/etcd/default.etcd\n              fi\n\n              for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do\n                  while true; do\n                      echo \"Waiting for ${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local to come up\"\n                      ping -W 1 -c 1 ${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local > /dev/null && break\n                      sleep 1s\n                  done\n              done\n\n              echo \"join member ${HOSTNAME}\"\n              # join member\n              exec etcd --name ${HOSTNAME} \\\n                  --initial-advertise-peer-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380 \\\n                  --listen-peer-urls http://${POD_IP}:2380 \\\n                  --listen-client-urls http://${POD_IP}:2379,http://127.0.0.1:2379 \\\n                  --advertise-client-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379 \\\n                  --initial-cluster-token etcd-cluster-1 \\\n                  --data-dir /var/run/etcd/default.etcd \\\n                  --initial-cluster $(initial_peers) \\\n                  --initial-cluster-state new",
			},
			Lifecycle: &corev1.Lifecycle{
				PreStop: &corev1.Handler{
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
			corev1.ServicePort{
				Name: "peer",
				Port: 2380,
			},
			corev1.ServicePort{
				Name: "client",
				Port: 2379,
			},
		},
	}
}
```

