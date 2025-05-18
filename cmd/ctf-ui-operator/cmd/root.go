package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/backbone81/ctf-ui-operator/internal/controller"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

var (
	enableDeveloperMode bool
	logLevel            int

	metricsBindAddress      string
	healthProbeBindAddress  string
	leaderElectionEnabled   bool
	leaderElectionNamespace string
	leaderElectionId        string

	kubernetesClientQPS   float32
	kubernetesClientBurst int
)

var rootCmd = &cobra.Command{
	Use:          "ctf-ui-operator",
	Short:        "This operator manages CTF UI instances.",
	Long:         `This operator manages CTF UI instances.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, zapLogger, err := utils.CreateLogger(logLevel, enableDeveloperMode)
		if err != nil {
			return fmt.Errorf("setting up logger: %w", err)
		}
		defer zapLogger.Sync() //nolint:errcheck // This is the logger we are flushing, no way to log the error here.

		if enableDeveloperMode {
			logger.Info("WARNING: Developer mode is enabled. This must not be used in production!")
		}

		ctrl.SetLogger(logger)

		restConfig, err := ctrl.GetConfig()
		if err != nil {
			return fmt.Errorf("setting up kubernetes config: %w", err)
		}
		restConfig.QPS = kubernetesClientQPS
		restConfig.Burst = kubernetesClientBurst

		mgr, err := ctrl.NewManager(
			restConfig,
			ctrl.Options{
				LeaderElection:          leaderElectionEnabled,
				LeaderElectionNamespace: leaderElectionNamespace,
				LeaderElectionID:        leaderElectionId,
				Metrics: metricsserver.Options{
					BindAddress: metricsBindAddress,
				},
				HealthProbeBindAddress: healthProbeBindAddress,
			},
		)
		if err != nil {
			return fmt.Errorf("setting up manager: %w", err)
		}

		reconciler := controller.NewReconciler(
			utils.NewLoggingClient(mgr.GetClient(), logger),
			controller.WithDefaultReconcilers(mgr.GetEventRecorderFor("ctf-ui-operator")),
		)
		if err := reconciler.SetupWithManager(mgr); err != nil {
			return fmt.Errorf("setting up reconciler with manager: %w", err)
		}

		if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
			return fmt.Errorf("setting up health check: %w", err)
		}
		if err := mgr.AddReadyzCheck("ready", healthz.Ping); err != nil {
			return fmt.Errorf("setting up ready check: %w", err)
		}
		return mgr.Start(ctrl.SetupSignalHandler())
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		cobra.CheckErr(bindFlagsToViper(rootCmd))
	})

	rootCmd.PersistentFlags().BoolVar(
		&enableDeveloperMode,
		"enable-developer-mode",
		false,
		"This option makes the log output friendlier to humans.",
	)
	rootCmd.PersistentFlags().IntVar(
		&logLevel,
		"log-level",
		0,
		"How verbose the logs are. Level 0 will show info, warning and error. Level 1 and up will show increasing details.",
	)

	initControllerRuntime()
	initKubernetesClient()
}

func initControllerRuntime() {
	rootCmd.PersistentFlags().StringVar(
		&metricsBindAddress,
		"metrics-bind-address",
		"0",
		"The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 "+
			"to disable the metrics service.",
	)
	rootCmd.PersistentFlags().StringVar(
		&healthProbeBindAddress,
		"health-probe-bind-address",
		"0",
		"The address the probe endpoint binds to.",
	)
	rootCmd.PersistentFlags().BoolVar(
		&leaderElectionEnabled,
		"leader-election-enabled",
		false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one "+
			"active controller manager.",
	)
	rootCmd.PersistentFlags().StringVar(
		&leaderElectionNamespace,
		"leader-election-namespace",
		"ctf-ui-operator",
		"The namespace in which leader election should happen.",
	)
	rootCmd.PersistentFlags().StringVar(
		&leaderElectionId,
		"leader-election-id",
		"ctf-ui-operator",
		"The ID to use for leader election.",
	)
}

func initKubernetesClient() {
	rootCmd.PersistentFlags().Float32Var(
		&kubernetesClientQPS,
		"kubernetes-client-qps",
		5.0,
		"The number of queries per second the Kubernetes client is allowed to send against the Kubernetes API.",
	)
	rootCmd.PersistentFlags().IntVar(
		&kubernetesClientBurst,
		"kubernetes-client-burst",
		10,
		"The number of burst queries the Kubernetes client is allowed to send against the Kubernetes API.",
	)
}

func bindFlagsToViper(cmd *cobra.Command) error {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("binding cobra flags to viper: %w", err)
	}

	var resultErr error
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if err := cmd.Flags().Set(flag.Name, fmt.Sprint(viper.Get(flag.Name))); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	})
	return resultErr
}
