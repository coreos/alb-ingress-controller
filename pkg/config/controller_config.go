package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	flagLogLevel                                  = "log-level"
	flagK8sClusterName                            = "cluster-name"
	flagServiceMaxConcurrentReconciles            = "service-max-concurrent-reconciles"
	flagTargetgroupBindingMaxConcurrentReconciles = "targetgroupbinding-max-concurrent-reconciles"
	defaultLogLevel                               = "info"
	defaultMaxConcurrentReconciles                = 3
	// High enough QPS to fit all expected use cases. QPS=0 is not set here, because
	// client code is overriding it.
	defaultQPS = 1e6
	// High enough Burst to fit all expected use cases. Burst=0 is not set here, because
	// client code is overriding it.
	defaultBurst = 1e6

)

// ControllerConfig contains the controller configuration
type ControllerConfig struct {
	// Log level for the controller logs
	LogLevel string
	// Name of the Kubernetes cluster
	ClusterName string
	// Configurations for the Ingress controller
	IngressConfig IngressConfig
	// Configurations for Addons feature
	AddonsConfig AddonsConfig
	// Configurations for the Controller Runtime
	RuntimeConfig RuntimeConfig
	// Max concurrent reconcile loops for Service objects
	ServiceMaxConcurrentReconciles int
	// Max concurrent reconcile loops for TargetGroupBinding objects
	TargetgroupBindingMaxConcurrentReconciles int
}

// NewControllerConfig constructs a new ControllerConfig object
func NewControllerConfig(scheme *runtime.Scheme) ControllerConfig {
	return ControllerConfig{
		RuntimeConfig: NewRuntimeConfig(scheme),
	}
}

// BindFlags binds the command line flags to the fields in the config object
func (cfg *ControllerConfig) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&cfg.LogLevel, flagLogLevel, defaultLogLevel,
		"Set the controller log level - info(default), debug")
	fs.StringVar(&cfg.ClusterName, flagK8sClusterName, "", "Kubernetes cluster name")
	fs.IntVar(&cfg.ServiceMaxConcurrentReconciles, flagServiceMaxConcurrentReconciles, defaultMaxConcurrentReconciles,
		"Maximum number of concurrently running reconcile loops for service")
	fs.IntVar(&cfg.TargetgroupBindingMaxConcurrentReconciles, flagTargetgroupBindingMaxConcurrentReconciles, defaultMaxConcurrentReconciles,
		"Maximum number of concurrently running reconcile loops for targetgroup binding")

	cfg.IngressConfig.BindFlags(fs)
	cfg.AddonsConfig.BindFlags(fs)
	cfg.RuntimeConfig.BindFlags(fs)
}

// Validate the controller configuration
func (cfg *ControllerConfig) Validate() error {
	if len(cfg.ClusterName) == 0 {
		return fmt.Errorf("Kubernetes cluster name must be specified")
	}
	if cfg.RuntimeConfig.Scheme == nil {
		return fmt.Errorf("Controller runtime scheme not initialzied")
	}
	return nil
}

// Get REST config for the controller runtime
func (cfg *ControllerConfig) BuildRestConfig() (*rest.Config, error) {
	restCfg, err := clientcmd.BuildConfigFromFlags(cfg.RuntimeConfig.APIServer, cfg.RuntimeConfig.KubeConfig)
	if err != nil {
		return nil, err
	}
	restCfg.QPS = defaultQPS
	restCfg.Burst = defaultBurst
	return restCfg, nil
}

// Get options for the controller runtime based on config
func (cfg *ControllerConfig) BuildRuntimeOptions() ctrl.Options {
	return ctrl.Options{
		Scheme:                  cfg.RuntimeConfig.Scheme,
		Port:                    cfg.RuntimeConfig.ControllerPort,
		MetricsBindAddress:      cfg.RuntimeConfig.MetricsBindAddress,
		LeaderElection:          cfg.RuntimeConfig.EnableLeaderElection,
		LeaderElectionID:        cfg.RuntimeConfig.LeaderElectionID,
		LeaderElectionNamespace: cfg.RuntimeConfig.LeaderElectionNamespace,
		Namespace:               cfg.RuntimeConfig.WatchNamespace,
		SyncPeriod:              &cfg.RuntimeConfig.SyncPeriod,
	}
}
