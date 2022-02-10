package networkinterfacectrl

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/resource/networking"
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

func TestNetworkInterfaceConstructor(t *testing.T) {
	tt = template.Must(tt.Parse(networkInterfaceTpl))
	o := &Output{}
	err := tt.Execute(o,
		&NetworkInterfaceParams{
			Name:      "dxp",
			Namespace: "aws",
			//Attachment:networking.ENIAttachment{},
			Attachment: networking.ENIAttachment{
				AttachTime: "abc",
				InstanceId: "123",
				Status:     "ok",
			},
			PrivateIpSets: []networking.ENIPrivateIpSet{
				{
					PublicIpAddress: "111",
					PrivateDnsName:  "aaa",
					Primary:         true,
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s", o.Data)
}
