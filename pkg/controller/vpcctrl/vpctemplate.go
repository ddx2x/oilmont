package vpcctrl

const (
	virtualPrivateCloudTpl = `
apiVersion: github.com/ddx2x/v1
kind: VPC
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
  {{ range .Labels}}
    {{ .Key }}: {{ .Value }}
  {{ end }}
spec:
  localName: {{.LocalName}}
  ip: {{.Ip}}
  mask: "{{.Mask}}"
  regionId: {{.Region}}
  status: {{.Status}}
  vpcId: {{.Id}}
`
)

type vpcLabel struct {
	Key   string
	Value interface{}
}

type virtualPrivateCloudParams struct {
	Name      string
	Namespace string
	Ip        string
	Mask      string
	Region    string
	Status    string
	Id        string
	Labels    []vpcLabel
	LocalName string
}
