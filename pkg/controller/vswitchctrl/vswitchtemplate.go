package vswitchctrl

const (
	vSwitchTpl = `apiVersion: github.com/ddx2x/v1
kind: VSwitch
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
  id: {{.Id}}
  mask: "{{.Mask}}"
  regionId: {{.Region}}
  status: {{.Status}}
  vpcId: {{.VpcId}}
  zone: {{.Zone}}
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
	Id        string
	Mask      string
	Region    string
	Status    string
	VpcId     string
	Zone      string
	Labels    []vpcLabel
	LocalName string
}
