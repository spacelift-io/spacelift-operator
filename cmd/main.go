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

package main

import (
	"flag"
	"os"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	kubezap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	appspaceliftiov1beta1 "github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/build"
	"github.com/spacelift-io/spacelift-operator/internal/controller"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/logging"
	"github.com/spacelift-io/spacelift-operator/internal/logging/encoders"
	spaceliftRepository "github.com/spacelift-io/spacelift-operator/internal/spacelift/repository"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/watcher"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(appspaceliftiov1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := kubezap.Options{
		Level: zap.NewAtomicLevelAt(zapcore.Level(-logging.Level2)),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	flag.Parse()
	zapOptions := kubezap.UseFlagOptions(&opts)
	if opts.Development {
		cfg := zap.NewDevelopmentEncoderConfig()
		cfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
		cfg.EncodeName = func(s string, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(color.YellowString(s))
		}
		opts.Encoder = encoders.NewPrettyConsoleEncoder(cfg)
	}
	ctrl.SetLogger(kubezap.New(zapOptions))
	setupLog.Info("Logger initialized", "level", opts.Level)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "4e414fab.app.spacelift.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	runRepo := repository.NewRunRepository(mgr.GetClient(), mgr.GetScheme())
	stackRepo := repository.NewStackRepository(mgr.GetClient(), mgr.GetScheme())
	stackOutputRepo := repository.NewStackOutputRepository(mgr.GetClient(), mgr.GetScheme(), mgr.GetEventRecorderFor("stack-output-repository"))
	spaceRepo := repository.NewSpaceRepository(mgr.GetClient())
	contextRepo := repository.NewContextRepository(mgr.GetClient(), mgr.GetScheme())
	secretRepo := repository.NewSecretRepository(mgr.GetClient())
	policyRepo := repository.NewPolicyRepository(mgr.GetClient(), mgr.GetScheme())
	spaceliftRunRepo := spaceliftRepository.NewRunRepository(mgr.GetClient())
	spaceliftStackRepo := spaceliftRepository.NewStackRepository(mgr.GetClient())
	spaceliftContextRepo := spaceliftRepository.NewContextRepository(mgr.GetClient())
	spaceliftPolicyRepo := spaceliftRepository.NewPolicyRepository(mgr.GetClient())
	runWatcher := watcher.NewRunWatcher(runRepo, spaceliftRunRepo)

	if err = (&controller.RunReconciler{
		RunRepository:            runRepo,
		StackRepository:          stackRepo,
		StackOutputRepository:    stackOutputRepo,
		SpaceliftRunRepository:   spaceliftRunRepo,
		SpaceliftStackRepository: spaceliftStackRepo,
		RunWatcher:               runWatcher,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Run")
		os.Exit(1)
	}
	if err = (&controller.StackReconciler{
		StackRepository:          stackRepo,
		SpaceRepository:          spaceRepo,
		SpaceliftStackRepository: spaceliftStackRepo,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Stack")
		os.Exit(1)
	}
	if err = (&controller.SpaceReconciler{
		SpaceRepository:          spaceRepo,
		SpaceliftSpaceRepository: spaceliftRepository.NewSpaceRepository(mgr.GetClient()),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Space")
		os.Exit(1)
	}
	if err = (&controller.ContextReconciler{
		ContextRepository:          contextRepo,
		StackRepository:            stackRepo,
		SpaceRepository:            spaceRepo,
		SecretRepository:           secretRepo,
		SpaceliftContextRepository: spaceliftContextRepo,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Context")
		os.Exit(1)
	}
	if err = (&controller.PolicyReconciler{
		StackRepository:           stackRepo,
		PolicyRepository:          policyRepo,
		SpaceRepository:           spaceRepo,
		SpaceliftPolicyRepository: spaceliftPolicyRepo,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Policy")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager", "version", build.Version)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
