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
