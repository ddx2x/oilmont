package mongo

import (
	"context"
	"fmt"
	"testing"

	"github.com/ddx2x/oilmont/pkg/datasource"

	"github.com/ddx2x/oilmont/pkg/core"
)

const TEST_RESOURCE_KIND = "test_resource_kind"

var ctx = context.Background()

var _ core.IObject = &TestResource{}

type TestResourceSpec struct{}

type TestResource struct {
	// Metadata default IObject Metadata
	core.Metadata `json:"metadata"`
	// Spec default TestResourceSpec Spec
	Spec TestResourceSpec `json:"spec"`
}

func (*TestResource) Decode(opData map[string]interface{}) (core.IObject, error) {
	t := &TestResource{}
	if err := core.UnmarshalToIObject(opData, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (a *TestResource) Clone() core.IObject {
	result := &TestResource{}
	core.Clone(a, result)
	return result
}

func init() {
	datasource.RegistryCoder(TEST_RESOURCE_KIND, &TestResource{})
}

// To go_test this code you need to use mongodb
/*
	docker run -itd --name mongo --net=host mongo mongod --replSet rs0
	docker exec -ti mongo mongo
	use admin;
	var cfg = {
		"_id": "rs0",
		"protocolVersion": 1,
		"members": [
			{
				"_id": 0,
				"host": "172.16.241.131:27017"
			},
		]
	};
	rs.initiate(cfg, { force: true });
	rs.reconfig(cfg, { force: true });
*/

const testIp = "127.0.0.1:27017"

// To go_test this code you need to use mongodb
func TestMongo_Apply(t *testing.T) {
	client, err, _ := NewMongo(ctx, "mongodb://"+testIp+"/admin")
	if err != nil {
		t.Fatal("open client error")
	}
	defer client.Close()
	testResource := &TestResource{
		Metadata: core.Metadata{
			Name:    "test_name",
			Kind:    core.Kind("test_resource_kind"),
			Version: "0",
			Labels:  map[string]interface{}{"who": "iam"},
		},
		Spec: TestResourceSpec{},
	}
	_ = testResource

	err = client.Get("default", "abc", "abc", nil, false)
	fmt.Println(err)
	// if _, _, err := client.Apply("default", "test_resource_kind", "test_name", testResource); err != nil {
	// 	t.Fatal(err)
	// }

	// testResource.Metadata.Name = "test_name1"
	// if _, _, err := client.Apply("default", "test_resource_kind", "test_name", testResource); err != nil {
	// 	t.Fatal(err)
	// }

}

func TestMongo_Watch(t *testing.T) {
	client, err, _ := NewMongo(ctx, "mongodb://"+testIp+"/admin")
	if err != nil {
		t.Fatal("open client error")
	}
	defer client.Close()

	testResource := &TestResource{
		Metadata: core.Metadata{
			Name:    "test_name",
			Kind:    core.Kind("test_resource_kind"),
			Version: "0",
			Labels:  map[string]interface{}{"who": "iam"},
		},
		Spec: TestResourceSpec{},
	}
	_ = testResource
	// if _, _, err := client.Apply("default", "test_resource_kind", "test_name", testResource); err != nil {
	// 	t.Fatal(err)
	// }

	// testResource.Metadata.Name = "test_name1"
	// testResource.Metadata.Version = 1
	// if _, _, err := client.Apply("default", "test_resource_kind", "test_name", testResource); err != nil {
	// 	t.Fatal(err)
	// }

	// watchChan := datasource.NewWatch(datasource.GetCoder(TEST_RESOURCE_KIND))
	// client.Watch("default", "fake", 0, watchChan)

	// item, ok := <-watchChan.ResultChan()
	// if !ok {
	// 	t.Fatal("watch item not ok")
	// }

	// if !(item.GetResourceVersion() > 0) {
	// 	t.Fatal("expected version failed")
	// }
}
