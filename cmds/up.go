package cmds

import (
	"fmt"
	"os"

	"github.com/appscode/krank/cloud/providers/digitalocean"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app/options"
	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus" // for client metric registration
	"k8s.io/kubernetes/pkg/cloudprovider"
	_ "k8s.io/kubernetes/pkg/cloudprovider/providers"
	_ "k8s.io/kubernetes/pkg/version/prometheus" // for version metric registration
	"k8s.io/kubernetes/pkg/version/verflag"
)

func init() {
	healthz.DefaultHealthz()
}

func NewCmdUp() *cobra.Command {
	s := options.NewCloudControllerManagerServer()
	cmd := &cobra.Command{
		Use:               "up",
		Short:             "Bootstrap as a Kubernetes master or node",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			cloud, err := cloudprovider.InitCloudProvider(digitalocean.ProviderName, s.CloudConfigFile)
			fmt.Println(s.CloudConfigFile, "----")
			if err != nil {
				glog.Fatalf("Cloud provider could not be initialized: %v", err)
			}

			if err := app.Run(s, cloud); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}
	s.AddFlags(cmd.Flags())
	return cmd
}

func main2() {
	s := options.NewCloudControllerManagerServer()
	s.AddFlags(pflag.CommandLine)

	flag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	verflag.PrintAndExitIfRequested()

	cloud, err := cloudprovider.InitCloudProvider(digitalocean.ProviderName, s.CloudConfigFile)
	fmt.Println(s.CloudConfigFile, "----")
	if err != nil {
		glog.Fatalf("Cloud provider could not be initialized: %v", err)
	}

	if err := app.Run(s, cloud); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
