<!--
 * @Date: 2022-11-21 11:06:55
 * @LastEditors: zhangwenqiang
 * @LastEditTime: 2022-11-21 13:50:58
 * @FilePath: /k8s/client-go-ym/3-operator/README.md
-->
# 安装kubebuilder
curl -LO https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.3.1/kubebuilder_2.3.1_linux_arm64.tar.gz
https://github.com/kubernetes-sigs/kubebuilder



# operator

operator 和kubebuilder用法类似,用一个就可以了

## 安装 operator-sdk

https://sdk.operatorframework.io/docs/installation/

```

-- 编译安装
Compile and install from master
Prerequisites
git
go version 1.18
Ensure that your GOPROXY is set to "https://proxy.golang.org|direct"

git clone https://github.com/operator-framework/operator-sdk
cd operator-sdk
git checkout master
make install  # 下载依赖包,生成go.mod go.sum
make build    # 生成二进制包

```







https://v1-1-x.sdk.operatorframework.io/docs/installation/install-operator-sdk/#install-from-homebrew-macos

```
RELEASE_VERSION=v1.1.0
# Linux
$ curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu

$ curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/ansible-operator-${RELEASE_VERSION}-x86_64-linux-gnu

$ curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/helm-operator-${RELEASE_VERSION}-x86_64-linux-gnu

# macOS
curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk-${RELEASE_VERSION}-x86_64-apple-darwin

curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/ansible-operator-${RELEASE_VERSION}-x86_64-apple-darwin

curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/helm-operator-${RELEASE_VERSION}-x86_64-apple-darwin


# Linux
$ chmod +x operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu && sudo mkdir -p /usr/local/bin/ && sudo cp operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu /usr/local/bin/operator-sdk && rm operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu

$ chmod +x ansible-operator-${RELEASE_VERSION}-x86_64-linux-gnu && sudo mkdir -p /usr/local/bin/ && sudo cp ansible-operator-${RELEASE_VERSION}-x86_64-linux-gnu /usr/local/bin/ansible-operator && rm ansible-operator-${RELEASE_VERSION}-x86_64-linux-gnu

$ chmod +x helm-operator-${RELEASE_VERSION}-x86_64-linux-gnu && sudo mkdir -p /usr/local/bin/ && sudo cp helm-operator-${RELEASE_VERSION}-x86_64-linux-gnu /usr/local/bin/helm-operator && rm helm-operator-${RELEASE_VERSION}-x86_64-linux-gnu

# macOS
chmod +x operator-sdk-${RELEASE_VERSION}-x86_64-apple-darwin && sudo mkdir -p /usr/local/bin/ && sudo cp operator-sdk-${RELEASE_VERSION}-x86_64-apple-darwin /usr/local/bin/operator-sdk && rm operator-sdk-${RELEASE_VERSION}-x86_64-apple-darwin
```





## Operator-sdk 创建项目

```
# 创建项目目录
$ mkdir -p opdemo && cd opdemo
$ export GO111MODULE=on  # 使用gomodules包管理工具
$ export GOPROXY="https://goproxy.cn" 
# 使用包代理，加速# 使用 sdk 创建一个名为 opdemo 的 operator 项目，如果在 GOPATH 之外需要指定 repo 参数
$ go mod init github.com/cnych/opdemo/v2      ⭐️
# go mod init opdemov2(或者)
# 使用下面的命令初始化项目
$ operator-sdk init --domain ydzs.io --license apache2 --owner "cnych"  ⭐️
Writing scaffold for you to edit...
Get controller runtime:
$ go get sigs.k8s.io/controller-runtime@v0.6.2
### go get sigs.k8s.io/controller-runtime@v0.13.0
go: downloading sigs.k8s.io/controller-runtime v0.6.2
go: downloading k8s.io/client-go v0.18.6
go: downloading k8s.io/utils v0.0.0-20200603063816-c1c6865ac451
go: downloading github.com/prometheus/procfs v0.0.11
go: downloading golang.org/x/net v0.0.0-20200520004742-59133d7f0dd7
go: downloading github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef
go: downloading k8s.io/apiextensions-apiserver v0.18.6
Update go.mod:
$ go mod tidy
Running make:
$ make
/Users/ych/devs/projects/go/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go build -o bin/manager main.go
Next: define a resource with:
$ operator-sdk create api
```

## 项目结构

使用 `operator-sdk init` 命令创建新的 Operator 项目后，项目目录就包含了很多生成的文件夹和文件。

- go.mod/go.sum  - Go Modules 包管理清单，用来描述当前 Operator 的依赖包。
- main.go 文件，使用 operator-sdk API 初始化和启动当前 Operator 的入口。
- deploy - 包含一组用于在 Kubernetes 集群上进行部署的通用的 Kubernetes 资源清单文件。
- pkg/apis - 包含定义的 API 和自定义资源（CRD）的目录树，这些文件允许 sdk 为 CRD 生成代码并注册对应的类型，以便正确解码自定义资源对象。
- pkg/controller - 用于编写所有的操作业务逻辑的地方
- version - 版本定义
- build - Dockerfile 定义目录

```
$ tree -L 2
.
├── Dockerfile
├── Makefile
├── PROJECT
├── bin
│   └── manager
├── config
│   ├── certmanager
│   ├── default
│   ├── manager
│   ├── prometheus
│   ├── rbac
│   ├── scorecard
│   └── webhook
├── go.mod
├── go.sum
├── hack
│   └── boilerplate.go.txt
└── main.go

10 directories, 8 files
```

## 生成资源 控制器代码

```
$ operator-sdk create api --group app --version v1beta1 --kind AppService
Create Resource [y/n]
y
Create Controller [y/n]
y
Writing scaffold for you to edit...
api/v1beta1/appservice_types.go
controllers/appservice_controller.go
Running make:
$ make
/Users/ych/devs/projects/go/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go build -o bin/manager main.go
```

## 处理error

```
上面操作一直报错,重新生成
mkdir -p github.com/cnych/opdemo
mkdir -p opdemo && cd opdemo
go mod init github.com/cnych/opdemo/v2
operator-sdk init --domain ydzs.io --license apache2 --owner "cnych"
go get github.com/onsi/ginkgo@v1.12.1
go get github.com/onsi/gomega@v1.10.1
go get github.com/go-logr/logr@v0.1.0
go get k8s.io/api/core/v1@v0.18.6
operator-sdk create api --group app --version v1beta1 --kind AppService
```



## 自定义api 资源类型

```
-- opdemov2/api/v1beta1/appservice_types.go

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

package v1beta1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppServiceSpec defines the desired state of AppService
type AppServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AppService. Edit appservice_types.go to remove/update
	Size      *int32                      `json:"size"`
	Image     string                      `json:"image"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	Envs      []corev1.EnvVar             `json:"envs,omitempty"`
	Ports     []corev1.ServicePort        `json:"ports,omitempty"`
}

// AppServiceStatus defines the observed state of AppService
type AppServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	appsv1.DeploymentStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AppService is the Schema for the appservices API
type AppService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppServiceSpec   `json:"spec,omitempty"`
	Status AppServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AppServiceList contains a list of AppService
type AppServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppService{}, &AppServiceList{})
}

```





## 添加控制器业务逻辑

### 实现

```go
$ 3-operator/opdemo/controllers/appservice_controller.go
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
	"reflect"

	"encoding/json"

	//"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1beta1 "opdemov2/api/v1beta1"
)

var (
	oldSpecAnnotation = "old/spec"
)

// AppServiceReconciler reconciles a AppService object
type AppServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	//Log    logr.Logger
}

//+kubebuilder:rbac:groups=app.ydzs.io,resources=appservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.ydzs.io,resources=appservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.ydzs.io,resources=appservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AppServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//ctx := context.Background()
	logger := log.FromContext(ctx)

	// TODO(user): your logic here
	// 业务逻辑实现
	// 获取 AppService 实例
	var appService appv1beta1.AppService
	err := r.Get(ctx, req.NamespacedName, &appService) //从req查询队列中的appservice对象
	if err != nil {
		// MyApp 被删除的时候，忽略
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
		// 在删除一个不存在的对象的时候,可能报not-found的错误
		// 这种情况不需要重新入队列排队修复
		return ctrl.Result{}, nil
	}

	// 当前对象被标记为了删除
	if appService.DeletionTimestamp != nil {
		return ctrl.Result{}, err
	}

	logger.Info("fetch appservice objects", "appservice", appService)

	// 如果不存在，则创建关联资源
	// 如果存在，判断是否需要更新
	//   如果需要更新，则直接更新
	//   如果不需要更新，则正常返回
	deploy := &appsv1.Deployment{}

	// 错误如果是找不到资源对象(NamespacedName pod的名称,用于查找对象,找不到就创建)
	if err := r.Get(ctx, req.NamespacedName, deploy); err != nil && errors.IsNotFound(err) {
		// 1. 关联 Annotations(注释上加上资源的spec信息,用来判断资源是否发生更改)
		data, _ := json.Marshal(appService.Spec)
		if appService.Annotations != nil {
			appService.Annotations[oldSpecAnnotation] = string(data)
		} else {
			//没有annotations的appService
			appService.Annotations = map[string]string{oldSpecAnnotation: string(data)}
		}
		if err := r.Client.Update(ctx, &appService); err != nil {
			return ctrl.Result{}, err
		}
		// 创建关联资源(deployment和service)
		// 2. 创建 Deployment
		deploy := NewDeploy(&appService)
		if err := r.Client.Create(ctx, deploy); err != nil {
			return ctrl.Result{}, err
		}

		// 3. 创建 Service
		service := NewService(&appService)
		if err := r.Create(ctx, service); err != nil {
			return ctrl.Result{}, err
		}
		// 创建成功返回
		return ctrl.Result{}, nil
	}

	// TODO,更新,判断是否需要更新(yaml文件是否发生变化)
	// yaml -> old yaml 我们可以从  annotations 里面去获取
	oldspec := appv1beta1.AppServiceSpec{}
	if err := json.Unmarshal([]byte(appService.Annotations[oldSpecAnnotation]), &oldspec); err != nil {
		return ctrl.Result{}, err
	}
	// 当前规范与旧的对象不一致，则需要更新
	if !reflect.DeepEqual(appService.Spec, oldspec) {
		// 更新关联资源
		newDeploy := NewDeploy(&appService) //new一个deployment对象,并不创建

		//查询olddeploy信息
		oldDeploy := &appsv1.Deployment{}
		if err := r.Get(ctx, req.NamespacedName, oldDeploy); err != nil {
			return ctrl.Result{}, err
		}

		//将appservice对象信息更新的部分,替换olddeploy
		oldDeploy.Spec = newDeploy.Spec
		//调接口更新

		// 更新重试: 如果更新函数返回“冲突”错误，RetryOnConflict将等待一段时间，如后退所述，然后重试
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Client.Update(ctx, oldDeploy)
		}); err != nil {
			return ctrl.Result{}, err
		}

		//套路一样,更新service
		newService := NewService(&appService)
		oldService := &corev1.Service{}
		if err := r.Get(ctx, req.NamespacedName, oldService); err != nil {
			return ctrl.Result{}, err
		}
		// 需要指定 ClusterIP 为之前的，不然更新会报错
		newService.Spec.ClusterIP = oldService.Spec.ClusterIP
		oldService.Spec = newService.Spec
		// 更新重试: 如果更新函数返回“冲突”错误，RetryOnConflict将等待一段时间，如后退所述，然后重试
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Client.Update(ctx, oldService)
		}); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.AppService{}).
		Complete(r)
}
```

```go
$ 3-operator/project/opdemov2/controllers/resources.go
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
```





### 优化

```
```







## 启动

```
make   #生成二进制包,变动了api/v1beta1中的资源文件
make install #在集群中创建crd自定义资源类型  检查:kubectl get crd
make run ENABLE_WEBHOOKS=false     #在集群外启动控制器
kubectl apply -f 资源.yaml
kubectl get appservice
kubectl get deployment
kubectl get service
```





## 卸载

```
$ kubectl delete -f config/samples/app_v1beta1_appservice.yaml
$ make uninstall
```





## 容器部署

```
如果是容器部署,运行在k8s中,需要创建对应的rbac权限给当前控制器watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.ydzs.io,resources=appservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.ydzs.io,resources=appservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.ydzs.io,resources=appservices/finalizers,verbs=update


-- 使用make生成对应的rbac yaml



$ export USERNAME=<dockerbub-username>
$ export USERNAME=cnych/opdemo:v1.0.0
$ make docker-build IMG=$USERNAME/opdemo:v1.0.0

# 还要改下部署的镜像,gcr镜像地址在外网,需要改成国内的,再Makefile中修改⭐️
docker-build:test   #test去掉


# 改下Dockerfile使用国内镜像⭐️
改下:RUN go mod download   --> RUN GOPROXY=https://goproxy.cn && go mod download
# 其他镜像
manager_auth_proxy_patch.yaml   grc.io/kubebuilder/kube-rbac-proxy:v0.5.0

$ make docker-push IMG=$USERNAME/opdemo:v1.0.0
$ make deploy IMG=$USERNAME/opdemo:v1.0.0

make docker-push IMG=cnych/opdemo:v1.0.0  #将镜像推送到dockerhub


#部署到k8s集群
make deploy IMG=cnych/opdemo:v1.0.0  


$ kubectl apply -f config/samples/app_v1beta1_appservice.yaml
$ kubectl get crd |grep myapp
myapps.app.ydzs.io   2020-11-06T07:06:54Z
```







