package main

import (
	"flag"
	"fmt"
	"os"

	examplev1 "github.com/jesusfcr/timeprinter-controller/api/v1alpha1"
	"github.com/jesusfcr/timeprinter-controller/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(examplev1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address for the metric endpoint")
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
	})
	if err != nil {
		panic(err)
	}
	cfg, err := ctrl.GetConfig()
	if err != nil {
		fmt.Println("Error getting config:", err)
		os.Exit(1)
	}
	fmt.Println(cfg.Host) // Should print your cluster API server URL
	reconciler := controllers.NewTimePrinterReconciler(mgr)
	ctrl.NewControllerManagedBy(mgr).
		For(&examplev1.TimePrinter{}).
		Complete(reconciler)

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		os.Exit(1)
	}
}
