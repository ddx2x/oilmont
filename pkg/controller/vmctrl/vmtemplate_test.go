package vmctrl

import (
	"fmt"
	"testing"
	"text/template"
)

var tt = template.New("template")

type Output struct{ Data []byte }

func (o *Output) Write(p []byte) (n int, err error) {
	o.Data = append(o.Data, p...)
	if len(o.Data) < 1 {
		err = fmt.Errorf("can't not copy")
	}
	return
}

func TestDataVolumeConstructor(t *testing.T) {
	tt = template.Must(tt.Parse(dataVolumeTpl))
	o := &Output{}
	err := tt.Execute(o,
		&dataVolumeParams{
			Name:      "dxp",
			Namespace: "dxp",
			Url:       "docker://laiks/fedora:latest",
			Storage:   "10Gi",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s", o.Data)
}

func TestVirtualMachineConstructor(t *testing.T) {
	tt = template.Must(tt.Parse(liziVirtualMachineTpl))
	o := &Output{}
	err := tt.Execute(o,
		&liziVirtualMachineParams{
			Name:           "dxp",
			Namespace:      "aws",
			DataVolumeDisk: "dxpDisk",
			CloudInitDisk:  "dxpCloudDisk",
			DataVolumeName: "dxpVolume",
			Memory:         "1Gi",
			Password:       "123456",
			Storage:        "10Gi",
			UserName:       "root",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s", o.Data)
}

func TestThirdVirtualMachineConstructor(t *testing.T) {
	tt = template.Must(tt.Parse(thirdVirtualMachineTpl))
	o := &Output{}
	err := tt.Execute(o,
		&thirdVirtualMachineParams{
			Name:   "dxp",
			Image:  "centos",
			Vendor: "AliCLoud",
			Region: "HK",
			Zone:   "HK1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s", o.Data)
}
