package controllers

import (
	"context"
	"fmt"
	"time"

	examplev1 "github.com/jesusfcr/timeprinter-controller/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type TimePrinterReconciler struct {
	client.Client
	runners map[string]context.CancelFunc
}

func NewTimePrinterReconciler(mgr ctrl.Manager) *TimePrinterReconciler {
	return &TimePrinterReconciler{
		Client:  mgr.GetClient(),
		runners: make(map[string]context.CancelFunc),
	}
}

func (r *TimePrinterReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	var tp examplev1.TimePrinter
	err := r.Get(ctx, req.NamespacedName, &tp)
	if err != nil {
		// Resource deleted
		if cancel, ok := r.runners[req.NamespacedName.String()]; ok {
			cancel()
			delete(r.runners, req.NamespacedName.String())
			fmt.Printf("🛑 Stopped timeprinter %s\n", req.NamespacedName)
		}
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Already running
	if _, ok := r.runners[req.NamespacedName.String()]; ok {
		return reconcile.Result{}, nil
	}

	// Start new runner
	cctx, cancel := context.WithCancel(context.Background())
	r.runners[req.NamespacedName.String()] = cancel

	go func() {
		fmt.Printf("▶️ Starting timeprinter %s every %d seconds\n", req.NamespacedName, tp.Spec.IntervalSeconds)
		ticker := time.NewTicker(time.Duration(tp.Spec.IntervalSeconds) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-cctx.Done():
				return
			case <-ticker.C:
				fmt.Printf("⏰ [%s] %s: %s\n", time.Now().Format(time.RFC3339), tp.Name, tp.Namespace)
			}
		}
	}()

	return reconcile.Result{}, nil
}
