package vmctrl

//docker://laiks/fedora:latest
const (
	dataVolumeTpl = `
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    vm: {{.VMName}}
spec:
  source:
    registry:
      url: {{.Url}}
  pvc:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: {{.VmStorage}}
`

	liziVirtualMachineTpl = `apiVersion: kubevirt.io/v1alpha3
kind: VirtualMachineInstance
metadata:
  labels:
    kubevirt.io/vm: {{.Name}}
  labels:
  {{ range .Labels}}
    {{ .Key }}: {{ .Value }}
  {{ end }}
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  runStrategy: "RerunOnFailure"
  template:
    metadata:
      labels:
        kubevirt.io/vm: {{.Name}}
    spec:
      domain:
        devices:
          disks:
          - disk:
              bus: virtio
            name: {{.DataVolumeDisk}}
          - disk:
              bus: virtio
            name: {{.CloudInitDisk}}
        machine:
          type: ""
        resources:
          requests:
            memory: {{.Memory}}
      terminationGracePeriodSeconds: 0
      volumes:
      - dataVolume:
          name: {{.DataVolumeName}}
        name: {{.DataVolumeDisk}}
      - cloudInitNoCloud:
          userData: |-
            #cloud-config
            password: {{.Password}}
            chpasswd: 
              expire: False
              list:
              - {{.UserName}}:{{.Password}}
        name: {{.CloudInitDisk}}
  dataVolumeTemplates:
  - metadata:
      name: {{.DataVolumeName}}
    spec:
      pvc:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: {{.VmStorage}}
      source:
        pvc:
          namespace: {{.Namespace}}
          name: {{.DataVolumeName}}`

	thirdVirtualMachineTpl = `apiVersion: github.com/ddx2x/v1
kind: VirtualMachine
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
  {{ range .Labels}}
    {{ .Key }}: {{ .Value }}
  {{ end }}
spec:
  cpu: {{.CPU}}
  memory: {{.Memory}}
  localName: {{.LocalName}}
  az: {{.Zone}}
  image: {{.Image}}
  instanceType: {{.InstanceType}}
  regionId: {{.Region}}
  rootDevice: {{.RootDevice}}
  vpcId:  {{.VpcId}}
  vSwitchId: {{.VSwitchId}}
  instanceId: {{.Id}}
  status: {{.Status}}
  state: {{.State}}
  os: {{.Os}}
  createTime: {{.CreateTime}}
  securityGroup:
  {{ range .SecurityGroup}}
  - {{.}}
  {{ end }}
`
)

type vmLabel struct {
	Key   string
	Value interface{}
}

type thirdVirtualMachineParams struct {
	Name          string
	Region        string
	SecurityGroup []string
	CPU           string
	Memory        string
	Zone          string
	Vendor        string
	Image         string
	InstanceType  string
	Namespace     string
	VpcId         string
	VSwitchId     string
	Id            string
	RootDevice    string
	LocalName     string
	Status        string
	State         string
	Labels        []vmLabel
	CreateTime    string
	Os            string
}

type liziVirtualMachineParams struct {
	Name           string
	Namespace      string
	DataVolumeName string
	DataVolumeDisk string
	CloudInitDisk  string
	Memory         string
	UserName       string
	Password       string
	Storage        string
	Labels         []vmLabel
}

type dataVolumeParams struct {
	VMName    string
	Namespace string
	Name      string
	Url       string
	Storage   string
}
