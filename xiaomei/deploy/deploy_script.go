package deploy

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/lovego/xiaomei/xiaomei/deploy/conf"
	"github.com/lovego/xiaomei/xiaomei/images"
	"github.com/lovego/xiaomei/xiaomei/release"
)

// TODO: keep container history
const deployScriptTmpl = `set -e
{{ range .VolumesToCreate }}
docker volume create {{ . }}
{{- end }}

deploy() {
	local type=$1
	local name=$2
	local args=$3
	docker stop $name >/dev/null 2>&1 && docker rm $name
	id=$(docker run --name=$name -d --restart=always $args)
	echo -n "$name starting "
	test "$type" = app && until docker exec "$id" sh -c 'wget -qO- http://localhost:${GOPORT:-3000}/_alive'; do
	  echo -n .; sleep 1s;
	done; echo
}

{{ range .Services -}}
args='{{.CommonArgs}}'
{{ $svc := . -}}
{{ range .Instances -}}
deploy {{$svc.Type}} {{$svc.Name}}.{{.}} "-e {{$svc.InstanceEnvName}}={{.}} $args"
{{ else -}}
deploy {{.Name}} "$args"
{{ end }}
{{ end -}}
`

func getDeployScript(svcNames []string, env, timeTag string) (string, error) {
	tmpl := template.Must(template.New(``).Parse(deployScriptTmpl))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, getDeployConfig(svcNames, env, timeTag)); err != nil {
		return ``, err
	}
	return buf.String(), nil
}

type deployConfig struct {
	VolumesToCreate []string
	Services        []serviceConfig
}
type serviceConfig struct {
	Type, Name, InstanceEnvName, CommonArgs string
	Instances                               []string
}

func getDeployConfig(svcNames []string, env, timeTag string) deployConfig {
	data := deployConfig{
		VolumesToCreate: conf.Get(env).VolumesToCreate,
	}
	for _, svcName := range svcNames {
		data.Services = append(data.Services, getServiceConf(svcName, env, timeTag))
	}
	return data
}

func getServiceConf(svcName, env, timeTag string) serviceConfig {
	commonArgs := getCommonArgs(svcName, env, timeTag)
	data := serviceConfig{
		Type:            svcName,
		Name:            release.ServiceName(svcName, env),
		InstanceEnvName: images.Get(svcName).InstanceEnvName(),
		CommonArgs:      strings.Join(commonArgs, ` `),
	}
	if data.InstanceEnvName != `` {
		data.Instances = conf.GetService(svcName, env).Instances()
	}
	return data
}
