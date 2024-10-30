package ossx

import "io"

// Oss defines the interface for interacting with OSS.
type Oss interface {
	PutFile(filename string, file io.Reader) (string, error)
}
