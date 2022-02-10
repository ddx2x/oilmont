package storagectrl

import "github.com/ddx2x/oilmont/pkg/resource/compute"

const (
	storageTpl = `
apiVersion: github.com/ddx2x/v1
kind: Storage
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
  {{ range .Labels}}
    {{ .Key }}: {{ .Value }}
  {{ end }}
spec:
  localName: {{.LocalName}}
  categoryType: {{.CategoryType}}
  deleteWithInstance: {{.DeleteWithInstance}}
  description: {{.Description}}
  diskChargeType: {{.DiskChargeType}}
  iops: {{.IOPS}}
  message: {{.Message}}
  region: {{.Region}}
  size: {{.Size}}
  throughput: {{.Throughput}}
  diskType: {{.DiskType}}
  state: {{.State}}
  status: {{.Status}}
  storageId: {{.StorageId}}
  zone: {{.Zone}}
  attachments:
  {{ range .Attachments}}
  - attachedTime: {{.AttachedTime}}
    device: {{.Device}}
    instanceId: {{.InstanceId}}
    state: {{.State}}
  {{ end }}
`
)

type storageLabel struct {
	Key   string
	Value interface{}
}

type StorageParams struct {
	Name               string
	Namespace          string
	Labels             []storageLabel
	Attachments        []compute.Attachment
	CategoryType       string
	DeleteWithInstance bool
	Description        string
	DiskChargeType     string
	DiskType           string
	IOPS               int
	LocalName          string
	Message            string
	Region             string
	Zone               string
	Size               int
	State              string
	Status             string
	StorageId          string
	Throughput         int
}
