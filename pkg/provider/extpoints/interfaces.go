package extpoints

import (
	"appscode.com/ark/pkg/contexts/env"
)

type KubeStarter interface {
	Run(req *env.ClusterStartupConfig) error
}
