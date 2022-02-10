package networkinterfacectrl

import "github.com/ddx2x/oilmont/pkg/resource/networking"

const (
	networkInterfaceTpl = `
apiVersion: github.com/ddx2x/v1
kind: NetworkInterface
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
  {{ range .Labels}}
    {{ .Key }}: {{ .Value }}
  {{ end }}
spec:
  {{- if .Attachment}}
  attachment:
    attachTime: {{.Attachment.AttachTime}}
    instanceId: {{.Attachment.InstanceId}}
    status: {{.Attachment.Status}}
  {{- end }}
  description: {{.Description}}
  localName: {{.LocalName}}
  macAddress: {{.MacAddress}}
  message: {{.Message}}
  networkInterfaceId: {{.NetworkInterfaceId}}
  privateDnsName: {{.PrivateDnsName}}
  {{- if .PrivateIpSets}}
  privateIpSets:
  {{range .PrivateIpSets}}
  - primary: {{.Primary}}
    privateDnsName: {{.PrivateDnsName}}
    privateIpAddress: {{.PrivateIpAddress}}
    publicIpAddress: {{.PublicIpAddress}}
  {{end}}
  {{- end }}
  publicDnsName: {{.PublicDnsName}}
  publicIpAddress: {{.PublicIpAddress}}
  region: {{.Region}}
  securityGroupIds:
  {{range .SecurityGroupIds}}
  - {{.}}
  {{end}}
  state: {{.State}}
  status: {{.Status}}
  subnetId: {{.SubnetId}}
  type: {{.Type}}
  vpcId: {{.VpcId}}
  zone: {{.Zone}}
`
)

type eniLabel struct {
	Key   string
	Value interface{}
}

type NetworkInterfaceParams struct {
	Name               string
	Namespace          string
	Labels             []eniLabel
	Attachment         networking.ENIAttachment
	Description        string
	NetworkInterfaceId string
	PrivateIpAddress   string
	PrivateDnsName     string
	Type               string
	VpcId              string
	SubnetId           string
	MacAddress         string
	PublicIpAddress    string
	PublicDnsName      string
	PrivateIpSets      []networking.ENIPrivateIpSet
	Ipv6               []string
	LocalName          string
	SecurityGroupIds   []string
	Region             string
	Zone               string
	Status             string
	State              string
	Message            string
}
