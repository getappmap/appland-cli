package metadata

import jsonpatch "github.com/evanphx/json-patch"

type Provider interface {
	Get(path string) (Metadata, error)
}

type Metadata interface {
	AsPatch() (*jsonpatch.Patch, error)
	IsValid() bool
}
