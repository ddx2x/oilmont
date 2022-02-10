package securitygroupctrl

import "github.com/ddx2x/oilmont/pkg/resource/system"

const (
	securityGroupTpl = `apiVersion: github.com/ddx2x/v1
kind: SecurityGroup
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
  {{ range .Labels}}
    {{ .Key }}: {{ .Value }}
  {{ end }}
spec:
  localName: {{.LocalName}}
  id: {{.Id}}
  regionId: "{{.Region}}"
  status: {{.Status}}
  vpcId: {{.VpcId}}
  ingress:
  {{ range .Ingress }}
  - ipProtocol: {{ .IpProtocol }}
    portRange: {{ .PortRange }}
    sourceCidrIp: {{ .SourceCidrIp }}
  {{ end }}
  egress:
  {{ range .Egress }}
  - ipProtocol: {{ .IpProtocol }}
    portRange: {{ .PortRange }}
    sourceCidrIp: {{ .SourceCidrIp }}
  {{ end }}
`
)

type securityGroupLabel struct {
	Key   string
	Value interface{}
}

type securityGroupParams struct {
	Ingress   []system.SecurityGroupRole
	Egress    []system.SecurityGroupRole
	Id        string
	Region    string
	Status    string
	VpcId     string
	Name      string
	Namespace string
	Labels    []securityGroupLabel
	LocalName string
}
