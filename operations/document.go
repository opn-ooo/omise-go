package operations

import (
	"github.com/omise/omise-go/internal"
)

type UploadDocument struct {
	File     []byte
	Filename string
	Kind     string
}

func (req *UploadDocument) Describe() *internal.Description {
	return &internal.Description{
		Endpoint: internal.APIStaging,
		Method:   "POST",
		Path:     "/documents",
	}
}
