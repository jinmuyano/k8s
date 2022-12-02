/*
* @Date: 2022-10-23 17:42:08
  - @LastEditors: zhangwenqiang
  - @LastEditTime: 2022-11-18 08:42:23
  - @FilePath: /k8s/client-go-ym/1-clientset/2-informer_test.go

监听deployment资源事件
*/
package main

import (
	"flag"
	"fmt"
	"path/filepath"

	// "reflect"
	"testing"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func TestInformer(*testing.T) {
	var err error
	var config *rest.Config

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "[可选] kubeconfig 绝对路径")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "kubeconfig 绝对路径")
	}
	// 初始化 rest.Config 对象
	if config, err = rest.InClusterConfig(); err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			panic(err.Error())
		}
	}
	// 创建 Clientset 对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 初始化 informer factory（为了测试方便这里设置每30s重新 List 一次）
	/*
		factory := &sharedInformerFactory{
			client:           client,
			namespace:        v1.NamespaceAll,
			defaultResync:    defaultResync,
			informers:        make(map[reflect.Type]cache.SharedIndexInformer),
			startedInformers: make(map[reflect.Type]bool),
			customResync:     make(map[reflect.Type]time.Duration),
		}
	*/
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	/*
		type deploymentInformer struct {
		factory          internalinterfaces.SharedInformerFactory
		tweakListOptions internalinterfaces.TweakListOptionsFunc
		namespace        string
		}
	*/
	deployInformer := informerFactory.Apps().V1().Deployments()

	/*
		调用func (f *sharedInformerFactory) InformerFor 注册deployInformer注册到Informer工厂
	*/
	// /Users/admin/go/pkg/mod/k8s.io/client-go@v0.25.3/informers/apps/v1/deployment.go  64  NewFilteredDeploymentInformer
	/*
		f.factory.InformerFor(&appsv1.Deployment{}, f.defaultInformer)
		    f.defaultInformer
			--> return cache.NewSharedIndexInformer(
			- 每个infomer需要实现自己的listwatch函数
			list:最后调用的是clientset获取全量数据
			client.AppsV1().Deployments(namespace).List(context.TODO(), options)

			watch:
			client.AppsV1().Deployments(namespace).Watch(context.TODO(), options)
	*/
	// 返回的是cache.NewSharedIndexInformer(listandwatch,type
	informer := deployInformer.Informer()

	// 注册事件处理程序,注册的事件会自动调用
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})

	// 创建 Lister⭐️,第一次全量获取deployment数据,相当于clientset的get,但是读的是本地缓存
	// &deploymentLister{indexer: indexer} 包含list方法
	deployLister := deployInformer.Lister()

	stopper := make(chan struct{})
	defer close(stopper)

	// 启动所有注册的informer，List & Watch
	// 调用每个注册的,informer(SharedIndexInformer).Run
	informerFactory.Start(stopper)
	// 等待所有启动的 Informer 的缓存被同步
	informerFactory.WaitForCacheSync(stopper)

	// 从本地缓存中获取 default 中的所有 deployment 列表
	deployments, err := deployLister.Deployments("default").List(labels.Everything())
	if err != nil {
		panic(err)
	}
	for idx, deploy := range deployments {
		fmt.Printf("%d -> %s\n", idx+1, deploy.Name)
	}
	// 阻塞当前程序
	<-stopper
}

func onAdd(obj interface{}) {
	deploy := obj.(*v1.Deployment)
	fmt.Println("add a deployment:", deploy.Name)
}

func onUpdate(old, new interface{}) {
	oldDeploy := old.(*v1.Deployment)
	newDeploy := new.(*v1.Deployment)
	fmt.Println("update deployment:", oldDeploy.Name, newDeploy.Name)
}

func onDelete(obj interface{}) {
	deploy := obj.(*v1.Deployment)
	fmt.Println("delete a deployment:", deploy.Name)
}
