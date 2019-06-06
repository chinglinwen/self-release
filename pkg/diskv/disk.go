// Store backend for diskv.
//
// Using this as the backend of store package,
// If for registering backend only (it can import as blank identifier).
package diskv

import (
	// "hdgit.com/golib/store"
	"github.com/peterbourgon/diskv"
)

type D struct {
	*diskv.Diskv
}

// type newer struct{}

// var defaultDir = ""

func New(path string) (*D, error) {
	d := diskv.New(diskv.Options{
		BasePath:     path,
		CacheSizeMax: 1024 * 1024,
	})
	return &D{Diskv: d}, nil
}

// // Write write any bytes to diskv.
// func (d *D) Write(key string, value []byte) (err error) {
// 	return d.Disk.Write(key, value)
// }

// // Read read bytes from diskv.
// func (d *D) Read(key string) ([]byte, error) {
// 	return d.Disk.Read(key)
// }

// func init() {
// 	store.Register("diskv", &newer{})
// }
