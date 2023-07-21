package agent

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	stateMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "computeblade_state",
		Name:      "state",
		Help:      "ComputeBlade state (label values are critical, identify, normal)",
	}, []string{"state"})
)

type computebladeState struct {
	mutex sync.Mutex

	// identifyActive indicates whether the blade is currently in identify mode
	identifyActive    bool
	identifyConfirmChan chan struct{}
	// criticalActive indicates whether the blade is currently in critical mode
	criticalActive    bool
	criticalConfirmChan chan struct{}
}

func NewComputeBladeState() *computebladeState {
	return &computebladeState{
		identifyConfirmChan: make(chan struct{}),
		criticalConfirmChan: make(chan struct{}),
	}
}

func (s *computebladeState) RegisterEvent(event Event) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	switch event {
	case IdentifyEvent:
		s.identifyActive = true
	case IdentifyConfirmEvent:
		s.identifyActive = false
		close(s.identifyConfirmChan)
		s.identifyConfirmChan = make(chan struct{})
	case CriticalEvent:
		s.criticalActive = true
		s.identifyActive = false
	case CriticalResetEvent:
		s.criticalActive = false
		close(s.criticalConfirmChan)
		s.criticalConfirmChan = make(chan struct{})
	}

	// Set identify state metric
	if s.identifyActive {
		stateMetric.WithLabelValues("identify").Set(1)
	} else {
		stateMetric.WithLabelValues("identify").Set(0)
	}

	// Set critical state metric
	if s.criticalActive {
		stateMetric.WithLabelValues("critical").Set(1)
	} else {
		stateMetric.WithLabelValues("critical").Set(0)
	}

	// Set critical state metric
	if !s.criticalActive && !s.identifyActive {
		stateMetric.WithLabelValues("normal").Set(1)
	} else {
		stateMetric.WithLabelValues("normal").Set(0)
	}
}

func (s *computebladeState) IdentifyActive() bool {
	return s.identifyActive
}

func (s *computebladeState) WaitForIdentifyConfirm(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.identifyConfirmChan:
		return nil
	}
}

func (s *computebladeState) CriticalActive() bool {
	return s.criticalActive
}

func (s *computebladeState) WaitForCriticalConfirm(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.criticalConfirmChan:
		return nil
	}
}
