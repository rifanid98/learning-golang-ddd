package fileupload

import (
	"path"

	"github.com/gofrs/uuid"
)

func FormatFile(fn string) string {
	ext := path.Ext(fn)
	uuid, _ := uuid.NewV4()
	newFilename := uuid.String() + ext
	return newFilename
}
