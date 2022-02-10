package obj

import (
	"encoding/json"
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/util/json"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func UnstructuredObjectToInstanceObj(src interface{}, dst interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

var _ io.Writer = &Output{}

type Output struct{ Data []byte }

func (o *Output) Write(p []byte) (n int, err error) {
	o.Data = append(o.Data, p...)
	if len(o.Data) < 1 {
		err = fmt.Errorf("can't not copy")
	}
	return
}

func Render(data interface{}, tpl string) (*unstructured.Unstructured, error) {
	t, err := template.New("").Parse(tpl)
	if err != nil {
		return nil, err
	}
	o := &Output{}
	if err := t.Execute(o, data); err != nil {
		return nil, err
	}

	object := make(map[string]interface{})
	if err := yaml.Unmarshal(o.Data, &object); err != nil {
		return nil, err
	}

	return &unstructured.Unstructured{Object: object}, nil
}

func RenderTemplate(data interface{}, tpl string) (string, error) {
	t, err := template.New("").Parse(tpl)
	if err != nil {
		return "", err
	}
	o := &Output{}
	if err := t.Execute(o, data); err != nil {
		return "", err
	}

	return string(o.Data), nil
}

func GetNestedString(obj map[string]interface{}, fields ...string) string {
	val, found, err := unstructured.NestedString(obj, fields...)
	if !found || err != nil {
		return ""
	}
	return val
}

func GetNestedMap(obj map[string]interface{}, fields ...string) map[string]interface{} {
	val, found, err := unstructured.NestedMap(obj, fields...)
	if !found || err != nil {
		return nil
	}
	return val
}

func GetNestedInt64(obj map[string]interface{}, fields ...string) int64 {
	val, found, err := unstructured.NestedInt64(obj, fields...)
	if !found || err != nil {
		return 0
	}
	return val
}

func GetNestedInt(obj map[string]interface{}, fields ...string) int {
	val, found, err := unstructured.NestedInt64(obj, fields...)
	if !found || err != nil {
		return 0
	}
	return int(val)
}

func GetNestedBool(obj map[string]interface{}, fields ...string) bool {
	val, found, err := unstructured.NestedBool(obj, fields...)
	if !found || err != nil {
		return false
	}
	return val
}
func Unmarshal(dest interface{}, obj runtime.Object) error {
	bs, err := k8sjson.Marshal(obj)
	if err != nil {
		return err
	}
	return k8sjson.Unmarshal(bs, dest)
}
