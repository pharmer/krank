package cmd

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"appscode.com/ark/pkg/contexts/env"
	"appscode.com/krank/pkg/provider/extpoints"
	api "github.com/appscode/api/kubernetes/v1beta1"
	"github.com/appscode/log"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"
)

func NewCmdStartup() *cobra.Command {
	var resp api.ClusterStartupConfigResponse
	cmd := &cobra.Command{
		Use:   "start-kubernetes",
		Short: "Bootstrap as a Kubernetes master or node",
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

			var req env.ClusterStartupConfig
			err = json.Unmarshal([]byte(resp.Configuration), &req)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			osFamily := OSFamily(req.OS)
			fmt.Printf("Using Provider=%v and OSFamily=%v\n", req.Provider, osFamily)

			if req.Provider == "" {
				fmt.Fprintln(os.Stderr, "Select a provider.")
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

			labels, err := env.ParseNodeLabels(req.NodeLabels)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to parse node labels.")
				os.Exit(1)
			}
			roleLabel := "node"
			if req.KubernetesMaster {
				roleLabel = "master"
			}
			req.NodeLabels = labels.
				WithInt64(env.NodeLabelKey_ContextVersion, req.ContextVersion).
				WithString(env.NodeLabelKey_Role, roleLabel).
				WithString(env.NodeLabelKey_SKU, resp.Sku).
				String()

			selector := req.Provider + "-" + osFamily
			for k := range extpoints.KubeStarters.All() {
				log.Infof("Found kube starters for %v", k)
			}
			script := extpoints.KubeStarters.Lookup(selector)
			if script == nil {
				req.Provider = "generic"
				log.Infof("Using selector: %v-%v", req.Provider, osFamily)
				script = extpoints.KubeStarters.Lookup(req.Provider + "-" + osFamily)
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
