package watcher

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/logging"
	spaceliftRepository "github.com/spacelift-io/spacelift-operator/internal/spacelift/repository"
)

type RunWatcher struct {
	Watcher

	lock             sync.Mutex
	watchedRuns      map[string]struct{}
	k8sRunRepo       *repository.RunRepository
	spaceliftRunRepo spaceliftRepository.RunRepository
}

func NewRunWatcher(k8sRunRepo *repository.RunRepository, spaceliftRunRepo spaceliftRepository.RunRepository) *RunWatcher {
	return &RunWatcher{
		Watcher:          DefaultWatcher,
		lock:             sync.Mutex{},
		watchedRuns:      map[string]struct{}{},
		k8sRunRepo:       k8sRunRepo,
		spaceliftRunRepo: spaceliftRunRepo,
	}
}

func (w *RunWatcher) IsWatched(run *v1beta1.Run) bool {
	if run != nil && run.Status.Id != "" {
		w.lock.Lock()
		defer w.lock.Unlock()
		if _, found := w.watchedRuns[run.Status.Id]; found {
			return true
		}
	}
	return false
}

func (w *RunWatcher) Start(ctx context.Context, run *v1beta1.Run) error {
	if run.Status.Id == "" {
		return errors.New("Can't watch a run that does not have any status.id")
	}
	if w.IsWatched(run) {
		return errors.New("Cannot watch run because it is already being watched")
	}
	logger := log.FromContext(ctx).
		WithName("run_watcher").WithValues(
		logging.RunId, run.Status.Id,
	)
	runId := run.Status.Id
	w.lock.Lock()
	w.watchedRuns[runId] = struct{}{}
	w.lock.Unlock()
	logger.Info("Starting watch")
	go func() {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, w.Timeout)
		defer func() {
			cancel()
			w.lock.Lock()
			delete(w.watchedRuns, runId)
			w.lock.Unlock()
		}()
		for {
			select {
			case <-ctxWithTimeout.Done():
				logger.WithValues(logging.RunState, run.Status.State).Info("Timeout watching for run changes")
				return
			default:
				var err error
				run, err = w.k8sRunRepo.Get(ctxWithTimeout, types.NamespacedName{Namespace: run.Namespace, Name: run.Name})
				if err != nil {
					if k8sErrors.IsNotFound(err) {
						logger.Info("Stopping run watcher since run has been removed from kube API")
						return
					}
					logger.Error(err, "Error fetching run from k8s API")
					time.Sleep(w.ErrInterval)
					continue
				}
				spaceliftRun, err := w.spaceliftRunRepo.Get(ctxWithTimeout, run)
				if err != nil {
					logger.Error(err, "Error fetching run from spacelift API")
					time.Sleep(w.ErrInterval)
					continue
				}

				run.SetRun(spaceliftRun)
				run.UpdateArgoHealth()
				if err := w.k8sRunRepo.UpdateStatus(ctxWithTimeout, run); err != nil {
					if k8sErrors.IsConflict(err) {
						logger.Info("Conflict updating run status, retrying immediately")
						continue
					}
					logger.Error(err, "Error updating run status")
					time.Sleep(w.ErrInterval)
					continue
				}

				if run.IsTerminated() {
					logger.WithValues(logging.RunState, run.Status.State).Info("Run is terminated, stopping run watcher")
					return
				}

				time.Sleep(w.Interval)
			}
		}
	}()
	return nil
}
