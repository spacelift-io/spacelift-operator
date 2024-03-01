/*
Copyright 2024.

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

package controller

import (
	"context"

	"github.com/shurcooL/graphql"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appspaceliftiov1beta1 "github.com/spacelift-io/spacelift-operator/api/v1beta1"
	spaceliftclient "github.com/spacelift-io/spacelift-operator/client"
)

// StackReconciler reconciles a Stack object
type StackReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.spacelift.io,resources=stacks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.spacelift.io,resources=stacks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.spacelift.io,resources=stacks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Stack object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *StackReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	spaceliftClient, err := spaceliftclient.GetSpaceliftClient(r.Client, req.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	var query struct {
		Stack struct {
			ID             string `graphql:"id"`
			Administrative bool   `graphql:"administrative"`
		} `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID("end-to-end-autoconfirm"),
	}

	err = spaceliftClient.Query(context.Background(), &query, variables)
	if err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Succesfully fetched info for stack", "id", query.Stack.ID, "administrative", query.Stack.Administrative)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StackReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appspaceliftiov1beta1.Stack{}).
		Complete(r)
}
