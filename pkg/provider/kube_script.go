package provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"appscode.com/ark/pkg/contexts/env"
	"appscode.com/ark/pkg/system"
	"appscode.com/krank/pkg/templates"
	api "github.com/appscode/api/kubernetes/v1beta1"
	"github.com/appscode/errors"
	"github.com/appscode/go/crypto/rand"
	netutil "github.com/appscode/go/net"
	"github.com/appscode/go/net/httpclient"
	"github.com/appscode/go/os/script"
	pu "github.com/appscode/go/pongo2"
	"github.com/appscode/log"
	sh "github.com/codeskyblue/go-sh"
	"github.com/flosch/pongo2"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/mgutz/str"
)

var (
	knownTokensFile string = "/srv/salt-overlay/salt/kube-apiserver/known_tokens.csv"
	basicAuthFile   string = "/srv/salt-overlay/salt/kube-apiserver/basic_auth.csv"
)

type HostData struct {
	APIServers       string
	HostnameOverride string
	InternalIP       string
	Iface            string
}

type KubeScript struct {
	script.Script
	Shell *sh.Session

	HostData HostData
	PongoCtx *pongo2.Context
	Req      *env.ClusterStartupConfig

	SetHostname           bool
	EnsureBasicNetworking func()
	FindMasterPd          func() (string, bool)
	WriteCloudConfig      func() error
}

func (ks *KubeScript) Perform() error {
	// create temporary working directory
	workingDir, err := ioutil.TempDir(os.TempDir(), "kubernetes")
	if err != nil {
		return errors.FromErr(err).Err()
	}
	fmt.Println(workingDir)
	defer os.RemoveAll(workingDir)

	ks.Script.Init(workingDir)
	ks.Shell = ks.Raw().(*sh.Session)

	// ref: http://forums.debian.net/viewtopic.php?f=10&t=112926
	err = os.RemoveAll("/var/lib/apt/lists")
	if err != nil {
		log.Errorf("Failed to clear apt listing due to %v", err)
	}

	// ks.Shell.Command("/usr/bin/apt-get", "update").Run()
	ks.InstallPkgs("curl")

	if ks.Req.KubernetesMaster {
		env := struct {
			env.MasterKubeEnv
			env.CommonKubeEnv
			env.KubeStartupConfig
		}{
			ks.Req.MasterKubeEnv,
			ks.Req.CommonKubeEnv,
			ks.Req.KubeStartupConfig,
		}
		// TODO(tamal): Should this be the actual private ip of the instance?
		// 127.0.0.1 is just less work for me.
		env.MasterInternalIP = "127.0.0.1"
		env.AzureCloudConfig = nil
		ks.PongoCtx, err = pu.YAMLSafeContext(env)
	} else {
		env := struct {
			env.NodeKubeEnv
			env.CommonKubeEnv
			env.KubeStartupConfig
		}{
			ks.Req.NodeKubeEnv,
			ks.Req.CommonKubeEnv,
			ks.Req.KubeStartupConfig,
		}
		ks.PongoCtx, err = pu.YAMLSafeContext(env)
	}
	if err != nil {
		return errors.FromErr(err).Err()
	}

	log.Info("== start-kubernetes node config starting ==")

	ks.SetBrokenMOTD()
	if ks.EnsureBasicNetworking != nil {
		ks.EnsureBasicNetworking()
	}
	// wait for network to be ready so that InternalIP() works
	if ks.SetHostname {
		err = ks.AssignHostname()
		if err != nil {
			return err
		}
	}

	ks.UpdateHostData()
	ks.EnsureInstallDir()
	err = ks.WriteEnv()
	if err != nil {
		return errors.FromErr(err).Err()
	}

	// Remove support for local disk
	// ref: https://github.com/kubernetes/kubernetes/blob/fe18055adc9ce44a907999f99d239ead47345d5d/cluster/aws/templates/configure-vm-aws.sh#L29

	time.Sleep(3 * time.Second)
	if ks.Req.Role == system.RoleKubernetesMaster {
		if ks.FindMasterPd != nil {
			if path, mount := ks.FindMasterPd(); mount {
				ks.MountMasterPd(path)
			}
		}
	}

	ks.CreateSaltPillar()
	ks.CreateCaCert()
	if ks.Req.Role == system.RoleKubernetesMaster {
		err = ks.CreateKubeMasterAuth()
		if err != nil {
			return err
		}

		// api token is used by kubed. This must be written from code instead of addon manager to
		// avoid potential race between kubed and addon manager.
		if ks.Req.AppsCodeNamespace != "" && ks.Req.AppsCodeApiToken != "" {
			ks.Mkdir("/var/run/secrets/appscode")
			tokenBytes, _ := json.Marshal(map[string]string{
				"namespace": ks.Req.AppsCodeNamespace,
				"token":     ks.Req.AppsCodeApiToken,
			})
			ks.WriteBytes("/var/run/secrets/appscode/api-token", tokenBytes)
		}
	}
	if ks.Req.Role == system.RoleKubernetesPool || ks.Req.RegisterMasterKubelet {
		err = ks.CreateSaltKubeletAuth()
		if err != nil {
			return errors.FromErr(err).Err()
		}
		ks.CreateSaltKubeproxyAuth()
		if err != nil {
			return errors.FromErr(err).Err()
		}
	}

	// download release
	ks.DownloadApps()
	ks.InstallHostfacts()
	ks.InstallSaltBase()

	ks.ConfigureSalt()
	ks.RemoveDockerArtifacts()
	if ks.WriteCloudConfig != nil {
		ks.WriteCloudConfig()
	}
	ks.Highstate()

	ks.SetGoodMOTD()
	log.Info("== start-kubernetes node config done ==")
	return nil
}

func (ks *KubeScript) SetBrokenMOTD() {
	startupLog := "/var/log/cloud-init-output.log"
	if ks.Req.Provider == "gce" {
		startupLog = "/var/log/startupscript.log"
	}
	ks.WriteString("/etc/motd", "Broken (or in progress) Kubernetes node setup! Suggested first step:\n  tail "+startupLog)
}

func Getent(key string, msg string) {
	for {
		out, _ := exec.Command("getent", str.ToArgv("hosts "+key)...).Output()
		result := strings.TrimSpace(string(out))
		log.Debug("RESULT=", result)
		if len(result) > 0 {
			break
		}
		log.Debug(msg)
		time.Sleep(3 * time.Second)
	}
}

/*
This is intended to be used for cloud providers which does not set hostname for instances, eg: Linode.
This should be run as soon as the networking in enabled. This will call the instance-by-ip (external ip) api to
retrieve hostname and sets it by running commands:

```sh
# ref: https://www.linode.com/docs/getting-started
hostnamectl set-hostname hostname
echo "$external_ip $hostname" >> /etc/hosts
```
If you are running local api server, then it will use Firebase to read the instance-by-ip response.
*/
func (ks *KubeScript) AssignHostname() error {
	externalIPs, _, err := netutil.HostIPs()
	if err != nil {
		return err
	}
	if len(externalIPs) == 0 {
		return errors.New("No external ip found.").Err()
	}
	externalIP := externalIPs[0].String()
	fmt.Print("Using extenal ip", externalIP)

	var respJson bytes.Buffer
	firebaseUid := os.Getenv("FIREBASE_UID")
	if firebaseUid != "" {
		path := fmt.Sprintf(`/k8s/%v/%v/%v/instance-by-ip/%v.json?auth=%v`,
			firebaseUid,
			ks.Req.AppsCodeNamespace,
			ks.Req.Name, // phid is grpc api
			strings.Replace(externalIP, ".", "_", -1),
			ks.Req.StartupConfigToken)
		_, err = httpclient.New(nil, nil, nil).
			WithBaseURL("https://tigerworks-kube.firebaseio.com").
			Call(http.MethodGet, path, nil, &respJson, false)
	} else {
		path := fmt.Sprintf("/kubernetes/v1beta1/clusters/%v/instance-by-ip/%v/json", ks.Req.PHID, externalIP)
		_, err = httpclient.New(nil, nil, nil).
			WithBaseURL(system.PublicAPIHttpEndpoint()).
			WithBearerToken(ks.Req.AppsCodeNamespace+":"+ks.Req.StartupConfigToken).
			Call(http.MethodGet, path, nil, &respJson, true)
	}
	if err != nil {
		return err
	}

	var resp api.ClusterInstanceByIPResponse
	err = jsonpb.Unmarshal(&respJson, &resp)
	if err != nil {
		return err
	}

	// https://www.linode.com/docs/getting-started
	err = ks.Run("hostnamectl", "set-hostname", resp.Instance.Name)
	if err != nil {
		return err
	}
	ks.AddLine("/etc/hosts", fmt.Sprintf("%v %v", resp.Instance.InternalIp, resp.Instance.Name))
	return nil
}

func (ks *KubeScript) UpdateHostData() {
	// used with initial_etcd_cluster, only needed in master instances
	// https://github.com/kubernetes/kubernetes/blob/7e92025cfd5b7e4eed7ba5f6d1bb315be5136e1c/cluster/saltbase/salt/etcd/etcd.manifest#L2
	// TODO(tamal): shortHostname := ks.CmdOut("", "hostname", "-s")
	if ks.Req.KubernetesMaster {
		log.Infof("Using HOSTNAME = %v", ks.Req.KubernetesMasterName)
		ks.PongoCtx.Update(pongo2.Context{"HOSTNAME": ks.Req.KubernetesMasterName})
	}

	if ks.HostData.HostnameOverride != "" {
		log.Infof("Using HOSTNAME_OVERRIDE = %v", ks.HostData.InternalIP)
		ks.PongoCtx.Update(pongo2.Context{"HOSTNAME_OVERRIDE": ks.HostData.HostnameOverride})
	}

	// Used to set kubelet.node-ip
	// https://github.com/kubernetes/kubernetes/blob/7ada99181cdcf6bcbde72ddffe7d975c00444090/pkg/kubelet/kubelet_node_status.go#L380
	// Used with flannel server --public-ip
	if ks.HostData.InternalIP != "" {
		log.Infof("Using INTERNAL_IP = %v", ks.HostData.InternalIP)
		ks.PongoCtx.Update(pongo2.Context{"INTERNAL_IP": ks.HostData.InternalIP})
	}

	// Used with flannel server --iface
	if ks.HostData.Iface != "" {
		log.Infof("Using HOST_IFACE = %v", ks.HostData.Iface)
		ks.PongoCtx.Update(pongo2.Context{"HOST_IFACE": ks.HostData.Iface})
	}

	if ks.HostData.APIServers != "" {
		log.Infof("Using API_SERVERS = %v", ks.HostData.APIServers)
		ks.PongoCtx.Update(pongo2.Context{"API_SERVERS": ks.HostData.APIServers})
	}
}

func (ks *KubeScript) SetGoodMOTD() {
	ks.WriteString("/etc/motd", `Welcome to Kubernetes!
You can find documentation for using Kubernetes at:
  https://appscode.com/docs/pharm/`)
}

func (ks *KubeScript) EnsureInstallDir() {
	ks.Mkdir("/var/cache/kubernetes-install")
}

func (ks *KubeScript) WriteEnv() error {
	var data []byte
	var err error
	if ks.Req.KubernetesMaster {
		env := struct {
			env.MasterKubeEnv
			env.CommonKubeEnv
		}{
			ks.Req.MasterKubeEnv,
			ks.Req.CommonKubeEnv,
		}
		// TODO(tamal): Should this be the actual private ip of the instance?
		// 127.0.0.1 is just less work for me.
		env.MasterInternalIP = "127.0.0.1"
		env.AzureCloudConfig = nil
		data, err = json.Marshal(env)
		if err != nil {
			return errors.FromErr(err).Err()
		}
	} else {
		env := struct {
			env.NodeKubeEnv
			env.CommonKubeEnv
		}{
			ks.Req.NodeKubeEnv,
			ks.Req.CommonKubeEnv,
		}
		data, err = json.Marshal(env)
		if err != nil {
			return errors.FromErr(err).Err()
		}
	}

	before := make(map[string]interface{})
	err = json.Unmarshal(data, &before)
	if err != nil {
		return errors.FromErr(err).Err()
	}

	after := make(map[string]interface{})
	for k, v := range before {
		switch u := v.(type) {
		case bool:
			if u {
				after[k] = "true"
			} else {
				after[k] = "false"
			}
		default:
			after[k] = v
		}
	}

	data, err = yaml.Marshal(after)
	if err != nil {
		return errors.FromErr(err).Err()
	}
	ks.WriteBytes("/var/cache/kubernetes-install/kube_env.yaml", data)
	return nil
}

func (ks *KubeScript) RemoveDockerArtifacts() {
	ks.InstallPkgs("bridge-utils")
	ks.Shell.Command("iptables", "-t", "nat", "-F").Run()
	ks.Shell.Command("ifconfig", "docker0", "down").Run()
	ks.Shell.Command("brctl", "delbr", "docker0").Run()
}

// --- KUBE MASTER ---
// Mounts a persistent disk (formatting if needed) to store the persistent data
// on the master -- etcd's data, a few settings, and security certs/keys/tokens.
//
// This function can be reused to mount an existing PD because all of its
// operations modifying the disk are idempotent -- safe_format_and_mount only
// formats an unformatted disk, and mkdir -p will leave a directory be if it
// already exists.
// ref: https://github.com/kubernetes/kubernetes/blob/c406665b2b1fdec98cd321c427896f6e4b959530/cluster/gce/configure-vm.sh#L276
func (ks *KubeScript) MountMasterPd(devicePath string) {
	// Format and mount the disk, create directories on it for all of the master's
	// persistent data, and link them to where they're used.
	fmt.Println("<><><><><><>>>>> Mounting master-pd")
	ks.Mkdir("/mnt/master-pd")

	ks.Shell.Command("/usr/share/google/safe_format_and_mount", "-m", "\"mkfs.ext4 -F\"", devicePath, "/mnt/master-pd").
		Command("tee", "/var/log/master-pd-mount.log").
		Run()
	// { echo "!!! master-pd mount failed, review /var/log/master-pd-mount.log !!!"; return 1; }

	// Contains all the data stored in etcd
	ks.Mkdir("/mnt/master-pd/var/etcd")
	ks.Chmod("/mnt/master-pd/var/etcd", 0700)
	// Contains the dynamically generated apiserver auth certs and keys
	ks.Mkdir("/mnt/master-pd/srv/kubernetes")
	// Contains the cluster's initial config parameters and auth tokens
	ks.Mkdir("/mnt/master-pd/srv/salt-overlay")
	// Directory for kube-apiserver to store SSH key (if necessary)
	ks.Mkdir("/mnt/master-pd/srv/sshproxy")
	ks.Symlink("/mnt/master-pd/var/etcd", "/var/etcd")
	ks.Symlink("/mnt/master-pd/srv/etcd", "/srv/etcd")
	ks.Symlink("/mnt/master-pd/srv/kubernetes", "/srv/kubernetes")
	ks.Symlink("/mnt/master-pd/srv/sshproxy", "/srv/sshproxy")
	ks.Symlink("/mnt/master-pd/srv/salt-overlay", "/srv/salt-overlay")

	// This is a bit of a hack to get around the fact that salt has to run after the
	// PD and mounted directory are already set up. We can't give ownership of the
	// directory to etcd until the etcd user and group exist, but they don't exist
	// until salt runs if we don't create them here. We could alternatively make the
	// permissions on the directory more permissive, but this seems less bad.
	if !ks.UserExists("etcd") {
		// Proactively delete locks. These seem to be problems during upgrades.
		os.Remove("/etc/passwd.lock")
		os.Remove("/etc/shadow.lock")
		os.Remove("/etc/group.lock")
		os.Remove("/etc/gshadow.lock")

		fmt.Println("Creating user: etcd")
		ks.Run("useradd", "-s", "/sbin/nologin", "-d", "/var/etcd", "etcd")
	}
	ks.ChownRecurse("/mnt/master-pd/var/etcd", "etcd")
	ks.ChgrpRecurse("/mnt/master-pd/var/etcd", "etcd")
}

func (ks *KubeScript) DownloadApps() {
	ks.Req.Apps[system.AppKubeServer].Download(ks.AppDir())
	ks.Req.Apps[system.AppKubeSaltbase].Download(ks.AppDir())
	ks.Req.Apps[system.AppHostfacts].Download(ks.AppDir())
}

// https://github.com/kubernetes/kubernetes/blob/master/cluster/juju/charms/trusty/kubernetes-master/hooks/hooks.py#L203
func (ks *KubeScript) CreateBasicAuthFile() error {
	f, err := os.Create(basicAuthFile)
	if err != nil {
		return errors.FromErr(err).Err()
	}
	defer f.Close()
	// password, user name, user id
	_, err = f.WriteString(ks.Req.KubePassword + "," + ks.Req.KubeUser + ",admin\n")
	if err != nil {
		return errors.FromErr(err).Err()
	}
	ks.Chmod(basicAuthFile, 600)
	return nil
}

func (ks *KubeScript) CreateTokensFile() error {
	f, err := os.Create(knownTokensFile)
	if err != nil {
		return errors.FromErr(err).Err()
	}
	defer f.Close()

	_, err = f.WriteString(ks.Req.KubeBearerToken + ",admin,admin\n") // TODO: Use KubeBeaerToken
	if err != nil {
		return errors.FromErr(err).Err()
	}
	_, err = f.WriteString(ks.Req.KubeletToken + ",kubelet,kubelet\n")
	if err != nil {
		return errors.FromErr(err).Err()
	}
	_, err = f.WriteString(ks.Req.KubeProxyToken + ",kube_proxy,kube_proxy\n")
	if err != nil {
		return errors.FromErr(err).Err()
	}
	serviceAccounts := []string{
		"system:scheduler",
		"system:controller_manager",
		"system:logging",
		"system:monitoring",
		"system:dns",
	}
	for _, account := range serviceAccounts {
		f.WriteString(rand.GenerateToken() + "," + account + "," + account + "\n")
	}
	ks.Chmod(knownTokensFile, 600)
	return nil
}

func (ks *KubeScript) CreateSaltPillar() {
	ks.Mkdir("/srv/salt-overlay/pillar")
	templates.Write("/srv/salt-overlay/pillar/cluster-params.sls", ks.PongoCtx, "kubernetes/cluster-params.sls")
}

func (ks *KubeScript) CreateSaltKubeletAuth() error {
	ks.Mkdir("/srv/salt-overlay/salt/kubelet")
	return templates.Write("/srv/salt-overlay/salt/kubelet/kubeconfig", ks.PongoCtx, "kubernetes/"+ks.Req.Provider+"/kubeconfig-kubelet.yml")
}

func (ks *KubeScript) CreateSaltKubeproxyAuth() error {
	ks.Mkdir("/srv/salt-overlay/salt/kube-proxy")
	return templates.Write("/srv/salt-overlay/salt/kube-proxy/kubeconfig", ks.PongoCtx, "kubernetes/"+ks.Req.Provider+"/kubeconfig-kube-proxy.yml")
}

func (ks *KubeScript) CreateCaCert() error {
	ks.Mkdir("/srv/kubernetes")
	caCert, err := base64.StdEncoding.DecodeString(ks.Req.CaCert)
	if err != nil {
		return errors.FromErr(err).Err()
	}
	ks.WriteBytes("/srv/kubernetes/ca.crt", caCert)
	return nil
}

func (ks *KubeScript) CreateKubeMasterAuth() error {
	ks.Mkdir("/srv/kubernetes")
	masterCert, _ := base64.StdEncoding.DecodeString(ks.Req.MasterCert)
	masterKey, _ := base64.StdEncoding.DecodeString(ks.Req.MasterKey)

	ks.WriteBytes("/srv/kubernetes/server.cert", masterCert)
	ks.WriteBytes("/srv/kubernetes/server.key", masterKey)

	// Kubelet auth: https://kubernetes.io/docs/admin/kubelet-authentication-authorization/#kubelet-authorization
	if ks.Req.EnableClusterVPN == "h2h-psk" {
		apiServerCert, _ := base64.StdEncoding.DecodeString(ks.Req.KubeAPIServerCert)
		apiServerKey, _ := base64.StdEncoding.DecodeString(ks.Req.KubeAPIServerKey)

		ks.WriteBytes("/srv/kubernetes/kube-apiserver.cert", apiServerCert)
		ks.WriteBytes("/srv/kubernetes/kube-apiserver.key", apiServerKey)

		hostfactsCert, _ := base64.StdEncoding.DecodeString(ks.Req.HostfactsCert)
		hostfactsKey, _ := base64.StdEncoding.DecodeString(ks.Req.HostfactsKey)

		ks.WriteBytes("/srv/kubernetes/hostfacts.cert", hostfactsCert)
		ks.WriteBytes("/srv/kubernetes/hostfacts.key", hostfactsKey)
	}

	ks.Mkdir("/srv/salt-overlay/salt/kube-apiserver")
	err := ks.CreateBasicAuthFile()
	if err != nil {
		return errors.FromErr(err).Err()
	}
	err = ks.CreateTokensFile()
	if err != nil {
		return errors.FromErr(err).Err()
	}

	//if _, err := os.Stat("/srv/kubernetes/ca.crt"); err != nil {
	//	if ks.Req.CACert != "" && ks.Req.MasterCert != "" && ks.Req.MasterKey != "" {
	//	}
	//}
	return nil
}

func (ks *KubeScript) ConfigureSalt() {
	if ks.Req.OS == "debian" {
		if ks.Req.Provider == "gce" {
			ks.UncommentLine("/etc/apt/sources.list", "deb.*http://http.debian.net/debian")
			ks.UncommentLine("/etc/apt/sources.list.d/backports.list", "deb.*http://ftp.debian.org/debian")
		}
		if ks.Req.Provider == "digitalocean" {
			ks.WriteString("/etc/apt/sources.list.d/backports.list", "deb http://http.debian.net/debian jessie-backports main")
		}
	}

	ks.Mkdir("/etc/salt/minion.d")
	templates.Write("/etc/salt/minion.d/local.conf", ks.PongoCtx, "kubernetes/local.conf")
	templates.Write("/etc/salt/minion.d/failhard.conf", ks.PongoCtx, "kubernetes/failhard.conf")

	if ks.Req.Role == system.RoleKubernetesMaster {
		// /etc/gce.conf
		if ks.Req.AppscodeAuthnUrl != "" {
			templates.Write("/etc/appscode_authn.config", ks.PongoCtx, "kubernetes/authn.config")
		}
		if ks.Req.AppscodeAuthzUrl != "" {
			templates.Write("/etc/appscode_authz.config", ks.PongoCtx, "kubernetes/authz.config")
		}
		templates.Write("/etc/salt/minion.d/grains.conf", ks.PongoCtx, "kubernetes/master-grains.conf")
	} else if ks.Req.Role == system.RoleKubernetesPool {
		// https://github.com/kubernetes/kubernetes/blob/c406665b2b1fdec98cd321c427896f6e4b959530/cluster/gce/configure-vm.sh#L742
		// if [[ -n "${EXTRA_DOCKER_OPTS-}" ]]; then
		// 	DOCKER_OPTS="${DOCKER_OPTS:-} ${EXTRA_DOCKER_OPTS}"
		// fi
		if ks.Req.ExtraDockerOpts != "" {
			dockerOpts := ks.Req.ExtraDockerOpts
			if extra, ok := (*ks.PongoCtx)["DOCKER_OPTS"]; ok && extra.(string) != "" {
				dockerOpts += (extra.(string) + " ")
			}
			ks.PongoCtx.Update(pongo2.Context{"DOCKER_OPTS": dockerOpts})
		}
		templates.Write("/etc/salt/minion.d/grains.conf", ks.PongoCtx, "kubernetes/node-grains.conf")
	}

	// --- install-salt ---

	// Based on https://major.io/2014/06/26/install-debian-packages-without-starting-daemons/
	// We do this to prevent Salt from starting the salt-minion
	// daemon. The other packages don't have relevant daemons. (If you
	// add a package that needs a daemon started, add it to a different
	// list.)
	ks.DeactivateDaemons()
	ks.InstallSaltMinion()
	ks.ActivateDaemons()
	ks.ProcessDisable("salt-minion") // stop-salt-minion
}

func (ks *KubeScript) InstallSaltBase() {
	// tar -xz -C /srv -f saltbase.tar.gz --strip=1
	saltbasePath := ks.AppDir() + "/" + ks.Req.Apps[system.AppKubeSaltbase].Name
	ks.Shell.Command("tar", "-xz", "-C", "/srv", "-f", saltbasePath, "--strip", "2").Run()
	ks.CheckPathExists(saltbasePath)
	ks.Shell.Command("/srv/install.sh", ks.AppDir()+"/"+ks.Req.Apps[system.AppKubeServer].Name).Run()
	// Validate that saltbase was installed correctly by checking existence of kube apiserver docker image.
	ks.CheckPathExists("/srv/salt/kube-bins/kube-apiserver.tar")
}

func (ks *KubeScript) InstallHostfacts() {
	o := ks.AppDir() + "/" + ks.Req.Apps[system.AppHostfacts].Name
	n := "/usr/bin/" + ks.Req.Apps[system.AppHostfacts].Name
	if err := os.Rename(o, n); err != nil {
		log.Fatal(errors.FromErr(err).Err())
	}
	ks.Script.Chmod(n, 0755)
	ks.CheckPathExists("/usr/bin/hostfacts")
}

func (ks *KubeScript) Highstate() {
	ks.Shell.Command("salt-call", "--local", "-l", "info", "state.highstate").Run()
}
