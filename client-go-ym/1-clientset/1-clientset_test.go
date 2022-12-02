/*
 * @Date: 2022-10-13 16:03:14
 * @LastEditors: zhangwenqiang
 * @LastEditTime: 2022-11-17 14:33:55
 * @FilePath: /k8s/client-go-ym/1-clientset/clientset_test.go
 */
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	// "encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestDeploymentList(t *testing.T) {
	var err error
	var config *rest.Config
	var kubeconfig *string

	if home := homeDir(); home != "" {
		fmt.Println(home)
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// 使用 ServiceAccount 创建集群配置（InCluster模式）
	if config, err = rest.InClusterConfig(); err != nil {
		// 使用 KubeConfig 文件创建集群配置
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			panic(err.Error())
		}
	}

	// 创建 clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 使用 clientsent 获取 Deployments,获取deployment信息
	ctx := context.Background()
	deployments, err := clientset.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for idx, deploy := range deployments.Items {
		fmt.Printf("%d -> %s\n", idx+1, deploy.Name)
	}

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
