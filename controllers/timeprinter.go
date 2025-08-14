package controllers

import (
	"context"
	"fmt"
	"time"

	examplev1 "github.com/jesusfcr/timeprinter-controller/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	timePrintedGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "timeprinter_last_printed_timestamp",
			Help: "The last printed time per TimePrinter, as Unix timestamp",
		},
		[]string{"name", "namespace"},
	)
	timesPrintedGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "timeprinter_printed_count",
			Help: "Number of times printed by each TimePrinter",
		},
		[]string{"name", "namespace"},
	)
)

func init() {
	metrics.Registry.MustRegister(activeTimePrinters, timePrintedGauge, timesPrintedGauge)
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

	cond := metav1.Condition{
		Type:               "Running",
		Status:             metav1.ConditionTrue,
		Reason:             "TimePrinterStarted",
		Message:            "The time printer is running",
		LastTransitionTime: metav1.Now(),
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

		cond.Type = "Reconcilied"
		cond.Reason = "TimePrinterUpdated"
	} else {
		activeTimePrinters.Inc()
	}

	meta.SetStatusCondition(&tp.Status.Conditions, cond)

	if tp.Status.StartTime == "" {
		tp.Status.StartTime = time.Now().UTC().Format(time.RFC3339)
	}
	err = r.Status().Update(ctx, &tp)
	if err != nil {
		fmt.Printf("❌ Failed to update status for %s: %v\n", req.NamespacedName, err)
		return reconcile.Result{}, err
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
				timePrintedGauge.WithLabelValues(tp.Name, tp.Namespace).Set(float64(time.Now().Unix()))
				timesPrintedGauge.WithLabelValues(tp.Name, tp.Namespace).Inc()
				fmt.Printf("⏰ [%s] %s: %s every %d\n", time.Now().Format(time.RFC3339), tp.Name, tp.Namespace, rd.IntervalSeconds)
			}
		}
	}(rd)

	return reconcile.Result{}, nil
}
