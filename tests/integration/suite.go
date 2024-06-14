package integration

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	kubezap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/repository/mocks"
)

const (
	DefaultTimeout  = 10 * time.Second
	DefaultInterval = 100 * time.Millisecond
)

type IntegrationTestSuite struct {
	suite.Suite

	SetupManager func(manager.Manager)

	ctx       context.Context
	cancel    context.CancelFunc
	k8sClient client.Client
	testEnv   *envtest.Environment
	Logs      *observer.ObservedLogs

	FakeSpaceliftRunRepo     *mocks.RunRepository
	FakeSpaceliftStackRepo   *mocks.StackRepository
	FakeSpaceliftSpaceRepo   *mocks.SpaceRepository
	FakeSpaceliftContextRepo *mocks.ContextRepository
	FakeSpaceliftPolicyRepo  *mocks.PolicyRepository

	RunRepo     *repository.RunRepository
	StackRepo   *repository.StackRepository
	SpaceRepo   *repository.SpaceRepository
	ContextRepo *repository.ContextRepository
	SecretRepo  *repository.SecretRepository
	PolicyRepo  *repository.PolicyRepository
}

func (s *IntegrationTestSuite) SetupSuite() {
	logger := kubezap.New(
		kubezap.Level(AllLogLevelEnabler{}),
		kubezap.WriteTo(os.Stdout),
		kubezap.UseDevMode(true),
		kubezap.RawZapOpts(
			zap.WrapCore(func(core zapcore.Core) zapcore.Core {
				zapCoreObserver, logs := observer.New(AllLogLevelEnabler{})
				s.Logs = logs
				return zapcore.NewTee(core, zapCoreObserver)
			})),
	)
	s.ctx, s.cancel = context.WithCancel(context.Background())

	_, filename, _, _ := runtime.Caller(0)
	basePath := path.Join(path.Dir(filename), "..", "..")
	s.testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join(basePath, "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join(basePath, "bin", "k8s",
			fmt.Sprintf("1.28.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	cfg, err := s.testEnv.Start()
	s.Require().NoError(err)
	s.Require().NotNil(cfg)

	err = v1beta1.AddToScheme(scheme.Scheme)
	s.Require().NoError(err)

	//+kubebuilder:scaffold:scheme

	s.k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	s.Require().NoError(err)
	s.Require().NotNil(s.k8sClient)

	// The SetLogger is here just to prevent a warning, but it's a no-op
	// See sigs.k8s.io/controller-runtime@v0.16.0/pkg/log/log.go:79
	ctrl.SetLogger(logr.New(log.NullLogSink{}))
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Logger: logger,
		Scheme: scheme.Scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	s.Require().NoError(err)

	s.Require().NotNil(s.SetupManager, "SetupManager should be defined")
	s.SetupManager(mgr)

	go func() {
		err := mgr.Start(s.ctx)
		s.Require().NoError(err)
	}()
}

func (s *IntegrationTestSuite) SetupTest() {
	s.Logs.TakeAll()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.cancel()
	err := s.testEnv.Stop()
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) Client() client.Client {
	return s.k8sClient
}

func (s *IntegrationTestSuite) Context() context.Context {
	return s.ctx
}
