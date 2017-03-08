package stack

import (
	"bytes"
	"gopkg.in/yaml.v2"
	"text/template"

	"github.com/bughou-go/xiaomei/config"
	"github.com/bughou-go/xiaomei/utils/cmd"
	"github.com/bughou-go/xiaomei/xiaomei/cluster"
	"github.com/fatih/color"
)

func Deploy(svcName string, doBuild, doPush bool) error {
	if doBuild {
		if err := BuildImage(svcName); err != nil {
			return err
		}
	}
	if doPush {
		if err := PushImage(svcName); err != nil {
			return err
		}
	}
	if svcName == `` {
		config.Log(color.GreenString(`deploying all services.`))
	} else {
		config.Log(color.GreenString(`deploying ` + svcName + ` service.`))
	}
	stack, err := getDeployStack(svcName)
	if err != nil {
		return err
	}
	script, err := getDeployScript(svcName)
	if err != nil {
		return err
	}
	return cluster.Run(cmd.O{Stdin: bytes.NewReader(stack)}, config.Env(), script)
}

func getDeployStack(svcName string) ([]byte, error) {
	stack := getStack()
	if svcName != `` {
		stack.Services = map[string]Service{svcName: stack.Services[svcName]}
	}
	for svcName, service := range stack.Services {
		if svcName == `app` {
			service[`environment`] = map[string]string{`GOENV`: config.Env()}
		}
	}
	return yaml.Marshal(stack)
}

const deployScriptTmpl = `
	cd && mkdir -p {{ .DeployName }} && cd {{ .DeployName }} &&
	cat - > {{ .FileName }}.yml &&
	docker stack deploy --compose-file={{ .FileName }}.yml {{ .Name }}
`

func getDeployScript(svcName string) (string, error) {
	deployConf := struct {
		Name, DeployName, FileName string
	}{
		Name: config.Name(), DeployName: config.DeployName(), FileName: `stack`,
	}
	if svcName != `` {
		deployConf.FileName = svcName
	}

	tmpl := template.Must(template.New(``).Parse(deployScriptTmpl))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, deployConf); err != nil {
		return ``, err
	}
	return buf.String(), nil
}