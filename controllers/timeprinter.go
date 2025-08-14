package controllers

import (
	"context"
	"fmt"
	"time"

	examplev1 "github.com/jesusfcr/timeprinter-controller/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	activeTimePrinters = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mycontroller_active_timeprinters",
		Help: "Number of active timeprinter goroutines running",
	})
)

func init() {
	metrics.Registry.MustRegister(activeTimePrinters)
}

type runnerData struct {
	Cancel          context.CancelFunc
	IntervalSeconds int
}

type TimePrinterReconciler struct {
	client.Client
	runners map[string]runnerData
}

func NewTimePrinterReconciler(mgr ctrl.Manager) *TimePrinterReconciler {
	return &TimePrinterReconciler{
		Client:  mgr.GetClient(),
		runners: make(map[string]runnerData),
	}
}

func (r *TimePrinterReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	var tp examplev1.TimePrinter
	err := r.Get(ctx, req.NamespacedName, &tp)
	if err != nil {
		// Resource deleted
		if rd, ok := r.runners[req.NamespacedName.String()]; ok {
			rd.Cancel()
			delete(r.runners, req.NamespacedName.String())
			fmt.Printf("🛑 Stopped timeprinter %s\n", req.NamespacedName)
			activeTimePrinters.Dec()
		}
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if tp.Status.StartTime == "" {
		tp.Status.StartTime = time.Now().UTC().Format(time.RFC3339)
		r.Status().Update(ctx, &tp)
	}
	// Already running
	if existing, ok := r.runners[req.NamespacedName.String()]; ok {
		if existing.IntervalSeconds == tp.Spec.IntervalSeconds {
			fmt.Printf("ℹ️ Timeprinter %s is already running every %d seconds\n", req.NamespacedName, tp.Spec.IntervalSeconds)
			return reconcile.Result{}, nil
		}
		// Cancel existing runner if interval has changed
		fmt.Printf("🔄 Updating timeprinter %s interval from %d to %d seconds\n", req.NamespacedName, existing.IntervalSeconds, tp.Spec.IntervalSeconds)
		existing.Cancel()
		delete(r.runners, req.NamespacedName.String())
	} else {
		activeTimePrinters.Inc()
	}

	// Start new runner
	cctx, cancel := context.WithCancel(context.Background())
	rd := runnerData{Cancel: cancel, IntervalSeconds: tp.Spec.IntervalSeconds}
	r.runners[req.NamespacedName.String()] = rd

	go func(rd runnerData) {
		fmt.Printf("▶️ Starting timeprinter %s every %d seconds\n", req.NamespacedName, rd.IntervalSeconds)
		ticker := time.NewTicker(time.Duration(tp.Spec.IntervalSeconds) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-cctx.Done():
				return
			case <-ticker.C:
				fmt.Printf("⏰ [%s] %s: %s every %d\n", time.Now().Format(time.RFC3339), tp.Name, tp.Namespace, rd.IntervalSeconds)
			}
		}
	}(rd)

	return reconcile.Result{}, nil
}
