package aws

import (
	"bufio"
	"bytes"
	"net/http"
	"os"
	"strings"
	"time"

	"appscode.com/ark/pkg/contexts/env"
	"appscode.com/krank/pkg/provider"
	"appscode.com/krank/pkg/provider/extpoints"
	"github.com/appscode/errors"
	"github.com/appscode/go/net/httpclient"
	"github.com/appscode/go/os/script"
	"github.com/appscode/log"
)

func init() {
	extpoints.KubeStarters.Register(new(debian), "aws-debian")
}

type debian struct {
}

func (*debian) Run(req *env.ClusterStartupConfig) error {
	script := &provider.KubeScript{
		Script: &script.DebianGeneric{},
		Req:    req,
	}
	// https://github.com/kubernetes/kubernetes/blob/fe18055adc9ce44a907999f99d239ead47345d5d/cluster/aws/templates/configure-vm-aws.sh#L117
	var privateDNS bytes.Buffer
	_, err := httpclient.New(nil, nil, nil).
		WithBaseURL("http://169.254.169.254").
		Call(http.MethodGet, "/2007-01-19/meta-data/local-hostname", nil, &privateDNS, false)
	if err != nil {
		return errors.FromErr(err).Err()
	}
	script.HostData = provider.HostData{
		HostnameOverride: privateDNS.String(),
	}
	script.HostData.APIServers = req.KubernetesMasterName

	// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-using-volumes.html
	script.FindMasterPd = func() (string, bool) {
		// ref: https://github.com/kubernetes/kubernetes/blob/fe18055adc9ce44a907999f99d239ead47345d5d/cluster/aws/templates/configure-vm-aws.sh#L53
		// $ grep "/mnt/master-pd" /proc/mounts
		file, err := os.Open("/proc/mounts")
		if err != nil {
			log.Error(err)
			return "", false
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "/mnt/master-pd") {
				log.Info("Master PD already mounted; won't remount")
				return "", false
			}
		}
		if err := scanner.Err(); err != nil {
			log.Error(err)
			return "", false
		}

		// Waiting for master pd to be attached
		attempt := 0
		for true {
			log.Infof("Attempt %v to check for /dev/xvdb", attempt)
			if _, err := os.Stat("/dev/xvdb"); err == nil {
				log.Info("Found /dev/xvdb")
				break
			}
			attempt += 1
			time.Sleep(1 * time.Second)
		}

		// Mount the master PD as early as possible
		script.AddLine("/etc/fstab", "/dev/xvdb /mnt/master-pd ext4 noatime 0 0")
		return "/dev/xvdb", true
	}

	return script.Perform()
}
