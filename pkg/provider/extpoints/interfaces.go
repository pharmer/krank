package extpoints

import (
	"github.com/appscode/pharmer/pkg/contexts/env"
)

type KubeStarter interface {
	Run(req *env.ClusterStartupConfig) error
}
