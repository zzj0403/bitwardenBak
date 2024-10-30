package ossx

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
)

// Oss defines the interface for interacting with OSS.
type Oss interface {
	PutFile(filename string, file io.Reader) (string, error)
	DirFilesList(dir string) (FilesList []oss.ObjectProperties, err error)
	DownloadFile(filename string, localPath string) error
}
