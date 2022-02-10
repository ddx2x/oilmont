package cr

import "testing"

func TestCustomResourceServer_CreateCustomDataRequest(t *testing.T) {
	data := `id,name,workspace,age,username
58215096-14a5-4d45-9105-affc3fcee4a1,李佳明1,default,1234,cde`
	ur := uploadRequest{Data: data}
	cds, _ := ur.toCDS("test")
	if len(cds) == 0 {
		t.Fatal("not expect value length")
	}

	_ = cds

	data2 := `name,workspace,age,username
李佳明1,default,1234,cde`
	ur2 := uploadRequest{Data: data2}
	cds, _ = ur2.toCDS("test")
	if len(cds) == 0 {
		t.Fatal("not expect value length")
	}

	_ = cds

	data3 := `name,workspace,age,username
李佳明1,default,1234,cde
李佳明2,default,1235,cde
,,,
`
	ur3 := uploadRequest{Data: data3}
	cds, err := ur3.toCDS("test")
	if err == nil {
		t.Fatal("not expect error")
	}

	_ = cds
}
