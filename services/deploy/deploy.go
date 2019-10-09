package deploy

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/lovego/cmd"
	"github.com/lovego/xiaomei/access"
	"github.com/lovego/xiaomei/release"
	"github.com/lovego/xiaomei/services/oam"
)

func deploy(svcName, env, timeTag, feature string, noWatch bool) error {
	psScript := fmt.Sprintf(` docker ps -f name=^/%s`, release.ServiceName(svcName, env))
	if !noWatch {
		psScript = oam.WatchCmd() + psScript
	}
	expectHighAvailable := len(release.GetCluster(env).GetNodes("")) >= 2
	var recoverAccess bool
	for _, node := range release.GetCluster(env).GetNodes(feature) {
		if svcs := node.Services(env, svcName); len(svcs) > 0 {
			if access.HasAccess(svcs) {
				if expectHighAvailable {
					if err := access.SetupNginx(env, "", node.Addr); err != nil {
						return err
					}
					time.Sleep(time.Second) // wait for nginx reloading finished.
				}
				recoverAccess = true
			}
			if err := deployNode(svcs, env, timeTag, node, psScript); err != nil {
				return err
			}
		}
	}
	if recoverAccess {
		if expectHighAvailable {
			return access.SetupNginx(env, "", "")
		} else {
			return access.ReloadNginx(env, "")
		}
	}
	return nil
}

func deployNode(svcs []string, env, timeTag string, node release.Node, psScript string) error {
	log.Println(color.GreenString(`deploying ` + node.SshAddr()))
	deployScript, err := getDeployScript(svcs, env, timeTag)
	if err != nil {
		return err
	}
	_, err = node.Run(cmd.O{}, deployScript+psScript)
	return err
}
