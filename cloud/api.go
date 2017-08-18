package cloud

import "github.com/appscode/pharmer/api"

type KubeStarter interface {
	Run(req *api.ClusterStartupConfig) error
}
