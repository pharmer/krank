package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/codeskyblue/go-sh"
)

func TestRemoveAPTLists(t *testing.T) {
	err := os.RemoveAll("/var/lib/apt/lists")
	fmt.Println(err)
}

func TestMountMasterPd(t *testing.T) {
	session := sh.NewSession()
	session.SetDir("/")
	session.ShowCMD = true
	session.Command("/usr/share/google/safe_format_and_mount", "-m", "\"mkfs.ext4 -F\"", "abc", "/mnt/master-pd").
		Command("tee", "/var/log/master-pd-mount.log").
		Run()
}
