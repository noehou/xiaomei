package stack

import (
	"bytes"
	"gopkg.in/yaml.v2"
	"text/template"

	"github.com/fatih/color"
	"github.com/lovego/xiaomei/utils"
	"github.com/lovego/xiaomei/utils/cmd"
	"github.com/lovego/xiaomei/xiaomei/cluster"
	"github.com/lovego/xiaomei/xiaomei/release"
)

var Driver driver

type driver struct{}

func (d driver) Deploy(svcName string, rmCurrent bool) error {
	if svcName == `` {
		utils.Log(color.GreenString(`deploying all services.`))
	} else {
		utils.Log(color.GreenString(`deploying ` + svcName + ` service.`))
	}
	stackYaml, err := getDeployStack(svcName)
	if err != nil {
		return err
	}
	script, err := getDeployScript(svcName, rmCurrent)
	if err != nil {
		return err
	}
	if _, err := cluster.Run(cmd.O{Stdin: bytes.NewReader(stackYaml)}, script); err != nil {
		return err
	}
	return d.Ps(svcName, true, nil)
}

func getDeployStack(svcName string) ([]byte, error) {
	stack := release.GetStack()
	if svcName != `` {
		stack.Services = map[string]release.Service{svcName: release.GetService(svcName)}
	}
	if app, ok := stack.Services[`app`]; ok {
		app[`environment`] = map[string]string{`GOENV`: release.Env()}
	}

	return yaml.Marshal(stack)
}

const deployScriptTmpl = `
	cd && mkdir -p {{ .DirName }} && cd {{ .DirName }} &&
	cat - > {{ .FileName }}.yml &&
	{{ .BeforeDeploy }}
	docker stack deploy --compose-file={{ .FileName }}.yml {{ .Name }}
`

func getDeployScript(svcName string, rmCurrent bool) (string, error) {
	deployConf := struct {
		Name, DirName, FileName, BeforeDeploy string
	}{
		Name: release.Name(), DirName: release.Name() + `_` + release.Env(), FileName: svcName,
	}
	if svcName == `` {
		deployConf.FileName = `stack`
	}
	if rmCurrent {
		if svcName == `` {
			deployConf.BeforeDeploy = `docker stack rm ` + release.Name() + waitUntilNetworkRemoved()
		} else {
			deployConf.BeforeDeploy = `docker service rm ` + release.Name() + `_` + svcName
		}
	}

	tmpl := template.Must(template.New(``).Parse(deployScriptTmpl))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, deployConf); err != nil {
		return ``, err
	}
	return buf.String(), nil
}

func waitUntilNetworkRemoved() string {
	return `
	until test -z "$(docker network ls -qf label=com.docker.stack.namespace=` + release.Name() + `)"; do
	sleep 0.1
	done
	`
}