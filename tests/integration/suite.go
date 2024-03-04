package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	kubezap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/controller"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/repository/mocks"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/watcher"
)

const (
	DefaultTimeout  = 10 * time.Second
	DefaultInterval = 100 * time.Millisecond
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx       context.Context
	cancel    context.CancelFunc
	k8sClient client.Client
	testEnv   *envtest.Environment
	Logs      *observer.ObservedLogs

	FakeSpaceliftRunRepo *mocks.RunRepository

	runRepo *repository.RunRepository
}

func (s *IntegrationTestSuite) SetupSuite() {
	logf.SetLogger(kubezap.New(
		kubezap.WriteTo(os.Stdout),
		kubezap.UseDevMode(true),
		kubezap.RawZapOpts(
			zap.WrapCore(func(core zapcore.Core) zapcore.Core {
				zapCoreObserver, logs := observer.New(zapcore.InfoLevel)
				s.Logs = logs
				return zapcore.NewTee(core, zapCoreObserver)
			})),
	))
	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
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

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	s.Require().NoError(err)

	s.runRepo = repository.NewRunRepository(mgr.GetClient())
	s.FakeSpaceliftRunRepo = new(mocks.RunRepository)
	w := watcher.NewRunWatcher(s.runRepo, s.FakeSpaceliftRunRepo)
	err = (&controller.RunReconciler{
		RunRepository:          s.runRepo,
		SpaceliftRunRepository: s.FakeSpaceliftRunRepo,
		RunWatcher:             w,
	}).SetupWithManager(mgr)
	s.Require().NoError(err)

	go func() {
		err := mgr.Start(s.ctx)
		s.Require().NoError(err)
	}()
}

func (s *IntegrationTestSuite) SetupTest() {
	s.FakeSpaceliftRunRepo.Test(s.T())
}

func (s *IntegrationTestSuite) TearDownTest() {
	s.FakeSpaceliftRunRepo.AssertExpectations(s.T())
	s.FakeSpaceliftRunRepo.Calls = nil
	s.FakeSpaceliftRunRepo.ExpectedCalls = nil
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.cancel()
	err := s.testEnv.Stop()
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) Client() client.Client {
	return s.k8sClient
}

func (s *IntegrationTestSuite) RunRepo() *repository.RunRepository {
	return s.runRepo
}

func (s *IntegrationTestSuite) Context() context.Context {
	return s.ctx
}
