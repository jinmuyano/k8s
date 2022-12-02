<!--
 * @Date: 2022-11-21 11:06:55
 * @LastEditors: zhangwenqiang
 * @LastEditTime: 2022-11-21 13:50:58
 * @FilePath: /k8s/client-go-ym/3-operator/README.md
-->
# 安装kubebuilder
curl -LO https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.3.1/kubebuilder_2.3.1_linux_arm64.tar.gz
https://github.com/kubernetes-sigs/kubebuilder



# 安装operator
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





## 创建项目

```
# 创建项目目录
$ mkdir -p opdemo && cd opdemo
$ export GO111MODULE=on  # 使用gomodules包管理工具
$ export GOPROXY="https://goproxy.cn" 
# 使用包代理，加速# 使用 sdk 创建一个名为 opdemo 的 operator 项目，如果在 GOPATH 之外需要指定 repo 参数
$ go mod init github.com/cnych/opdemo/v2
# 使用下面的命令初始化项目
$ operator-sdk init --domain ydzs.io --license apache2 --owner "cnych"
Writing scaffold for you to edit...
Get controller runtime:
$ go get sigs.k8s.io/controller-runtime@v0.6.2
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

## 添加API

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



## 自定义api

```
api/v1beta1/appservice_types.go
AppServiceSpec


import (
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

--新增需要的字段
// AppServiceSpec defines the desired state of AppService
type AppServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Size      *int32                        `json:"size"`
	Image     string                        `json:"image"`
	Resources corev1.ResourceRequirements   `json:"resources,omitempty"`  //limits,requests
	Envs      []corev1.EnvVar               `json:"envs,omitempty"` 			//环境变量
	Ports     []corev1.ServicePort          `json:"ports,omitempty"`
}


--新增字段
// AppServiceStatus defines the observed state of AppService
type AppServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	appsv1.DeploymentStatus `json:",inline"`  //inline 内联属性到该结构体
}

-- 添加后执行
make
```





## 添加控制器业务逻辑

```
3-operator/opdemo/controllers/appservice_controller.go
```



