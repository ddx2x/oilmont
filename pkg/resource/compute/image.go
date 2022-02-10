package compute

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const ImageKind core.Kind = "image"

type ImageSpec struct {
	Version string `json:"version"`
	Os      string `json:"os"`
	Region  string `json:"region"`
	ID      string `json:"id"`
}

type Image struct {
	core.Metadata `json:"metadata"`
	Spec          ImageSpec `json:"spec"`
}

func (i *Image) Clone() core.IObject {
	result := &Image{}
	core.Clone(i, result)
	return result
}

func (*Image) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Image{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type ImageList struct {
	core.Metadata `json:"metadata"`
	Items         []Image `json:"items"`
}

func (a *ImageList) GenerateListVersion() {
	var maxVersion string
	for _, item := range a.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	a.Metadata = core.Metadata{
		Kind:    "imageList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(ImageKind), &Image{})
}
