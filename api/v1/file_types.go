package v1

import "helm.sh/helm/v3/pkg/chart/loader"

type BufferedFileList []*BufferedFile

func (b BufferedFileList) Slice() []*loader.BufferedFile {
	slice := make([]*loader.BufferedFile, len(b))
	for i := range b {
		slice[i] = &loader.BufferedFile{
			Name: b[i].Name,
			Data: []byte(b[i].Data),
		}
	}
	return slice
}

type BufferedFile struct {
	Name string `json:"name"`
	Data string `json:"data"`
}
