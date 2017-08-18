package cmds

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	proto "github.com/appscode/api/kubernetes/v1beta1"
	"github.com/appscode/krank/cloud"
	"github.com/appscode/log"
	"github.com/appscode/pharmer/api"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"
)

func NewCmdStartup() *cobra.Command {
	var resp proto.ClusterStartupConfigResponse
	cmd := &cobra.Command{
		Use:               "krank",
		Short:             "Bootstrap as a Kubernetes master or node",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			conf, err := ioutil.ReadAll(reader)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			log.Infof("conf: %v", string(conf))

			err = jsonpb.UnmarshalString(string(conf), &resp)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			var req api.ClusterStartupConfig
			err = json.Unmarshal([]byte(resp.Configuration), &req)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			osFamily := OSFamily(req.OS)
			fmt.Printf("Using Provider=%v and OSFamily=%v\n", req.Provider, osFamily)

			if req.Provider == "" {
				fmt.Fprintln(os.Stderr, "Select a cloud.")
				os.Exit(1)
			}
			if osFamily == "" {
				fmt.Fprintln(os.Stderr, "Select an os.")
				os.Exit(1)
			}
			if req.Role == "" {
				fmt.Fprintln(os.Stderr, "Select a Kubernetes role for this node.")
				os.Exit(1)
			}

			labels, err := api.ParseNodeLabels(req.NodeLabels)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to parse node labels.")
				os.Exit(1)
			}
			roleLabel := "node"
			if req.KubernetesMaster {
				roleLabel = "master"
			}
			req.NodeLabels = labels.
				WithInt64(api.NodeLabelKey_ContextVersion, 1 /*req.ContextVersion*/).
				WithString(api.NodeLabelKey_Role, roleLabel).
				WithString(api.NodeLabelKey_SKU, resp.Sku).
				String()

			selector := req.Provider + "-" + osFamily
			for k := range cloud.CloudProviders() {
				log.Infof("Found kube starters for %v", k)
			}
			script, err := cloud.GetCloudProvider(selector, nil)
			if err != nil {
				req.Provider = "generic"
				log.Infof("Using selector: %v-%v", req.Provider, osFamily)
				script, err = cloud.GetCloudProvider(req.Provider+"-"+osFamily, nil)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
			err = script.Run(&req)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	cmd.Flags().StringVar(&resp.Sku, "sku", "", "SKU of this instance")

	return cmd
}

func OSFamily(os string) string {
	os = strings.ToLower(os)
	os = strings.Replace(os, "_", " ", -1)
	os = strings.Replace(os, "-", " ", -1)
	return strings.Split(os, " ")[0]
}
