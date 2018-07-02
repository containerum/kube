// Code generated by fileb0x at "2018-07-02 13:34:29.364048513 +0300 MSK m=+0.024900984" from config file "b0x.yaml" DO NOT EDIT.
// modification hash(5193e4a76ed62bfa85a4c58e3a03fb48.d47d734c826bc110696e7160789432d0)

package static

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path"

	"context"
	"golang.org/x/net/webdav"
)

var (
	// CTX is a context for webdav vfs
	CTX = context.Background()

	// FS is a virtual memory file system
	FS = webdav.NewMemFS()

	// Handler is used to server files through a http handler
	Handler *webdav.Handler

	// HTTP is the http file system
	HTTP http.FileSystem = new(HTTPFS)
)

// HTTPFS implements http.FileSystem
type HTTPFS struct{}

// FileSwaggerJSON is "swagger.json"
var FileSwaggerJSON = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\xec\x5d\xeb\x6f\xdc\x38\x92\xff\x9e\xbf\x82\xe8\x3b\x60\x66\x70\x69\xf7\x6c\x76\xb1\xc0\xcd\xb7\xac\x3d\x0f\x23\x71\xae\x61\xc7\x33\x0b\x6c\x02\x83\x96\xd8\xdd\xdc\xa8\x49\x0d\x49\x39\xf1\x1a\xfe\xdf\x0f\x22\xf5\x6e\x3d\xa8\x67\xcb\x76\x7d\x8a\x63\x53\x64\x91\xc5\xfa\x55\xb1\xaa\x58\x7c\x78\x85\xd0\xc2\xe1\x4c\x06\x7b\x22\x17\x3f\xa1\x7f\xbd\x42\x08\xa1\x05\xf6\x7d\x8f\x3a\x58\x51\xce\x56\xff\x96\x9c\x2d\x5e\x21\xf4\xf9\x75\xd8\xd6\x17\xdc\x0d\x1c\xbb\xb6\xf2\x2b\xde\x6e\x89\x58\xfc\x84\x16\x6f\x4e\x7e\x5c\xe8\xdf\x51\xb6\xe1\x8b\x9f\xd0\x83\xf9\xd6\x25\xd2\x11\xd4\x0f\xbf\x0d\x5b\xbd\x0b\x6e\x09\x7a\xbb\x3e\x47\x92\x88\x3b\xea\x10\x44\x65\xf2\xe3\x86\x0b\xe4\x70\xc6\x88\x13\xb6\x46\x5f\xa9\xda\xa1\xb0\xbd\x60\x44\x11\x79\xa2\x7b\x47\x68\xa1\xa8\xf2\x48\xb6\xaf\xf8\x0f\x77\x44\xc8\x68\x98\x1f\x4f\x7e\x3c\xf9\x4b\x48\xe6\xa3\x99\x12\x56\x3b\x99\xd2\xb4\x72\x38\xdb\xd0\xed\x1e\xfb\xe9\x2f\x11\x5a\x6c\x89\xca\xfc\x37\x1c\x09\x6f\xd3\x55\x88\x7e\x77\xaa\x3f\xbd\xc0\xfe\x22\xf9\xed\xe7\xd7\xe9\x27\x32\xd8\xef\xb1\xb8\x0f\x69\xf8\x95\x28\x64\x06\x42\xe1\x48\x68\x23\xf8\x1e\x61\xcf\x43\x81\x24\x02\x31\xbc\x27\xd2\xc7\x4e\x3a\x31\xfd\x3d\xf7\x89\xd0\x0b\x7d\xee\x46\x7d\x5c\x11\x8f\x38\x8a\xb8\xc9\xc0\x32\xdb\xde\xc7\x02\xef\x89\x22\xa2\x48\xe8\x43\xe6\xe7\x70\x2a\xf7\xbe\x5e\x33\xa9\x04\x65\xdb\x4c\x0f\xfa\xaf\x1b\x2e\xf6\x38\x9c\xfc\x22\x08\xa8\x5b\xfc\x6b\x48\x6a\xf8\xb7\x7f\x2e\xaf\x25\x11\xcb\xf3\xb3\x62\x03\xaa\xd7\x7c\x47\xb0\x4b\x44\xf1\x6f\x82\xfc\x19\x50\x41\xc2\xd9\x28\x11\x90\xcc\x1f\x1f\x5f\x57\x93\x4b\x58\xb0\x2f\x4c\x48\xff\x3e\x5c\xba\xc2\x08\xe1\xfe\x74\xf7\x94\x2d\x72\xbf\xfd\xfc\xba\xcd\xfc\x0b\x33\xbc\xe4\x1e\x19\x68\x8e\xa5\x9b\x44\x10\xe9\x73\x26\x89\xcc\x6d\x37\x84\x16\x6f\x7e\xfc\xb1\xf0\xab\x43\x09\xca\x6e\x29\x8f\x4a\x95\xdf\x57\x32\xb3\xb1\x8a\x64\x4a\x67\x47\xf6\xf8\x60\x00\x84\x16\xff\x2d\xc8\x26\xec\xfb\xbf\x56\x2e\xd9\x50\x46\xc3\xb1\xe4\xea\x70\xe7\xbd\xa7\x52\xe5\xd7\xf9\xb1\x8a\xa1\x0b\x97\x6c\x70\xe0\xa9\xe6\xf9\xec\x88\x10\xf7\x88\x08\xc1\x45\x5f\x8a\x89\x10\x35\xe4\xbd\x2a\x21\x74\xf1\x6d\xb9\x27\x6a\xc7\xdd\xe5\x1d\x95\xf4\x96\x7a\x54\x69\xe1\xf5\x83\x5b\x8f\x3a\x71\x67\xe6\xd3\xe8\xb3\xc5\x8a\xb2\xad\x20\x32\xcf\x3f\x3b\xf4\x38\x37\x5f\x5a\x61\x47\x32\x4a\x5f\xe4\x38\x4f\xc8\x05\xe0\x78\xc1\xc0\x91\xee\xa7\x69\x60\x23\xd9\x76\x2f\x08\x35\x04\xbd\xc3\x8a\x54\xc0\x46\x66\x89\xdb\xe2\xc6\x87\xf8\x53\x2b\xe4\x48\x07\xd2\xac\x6e\x80\x89\xa4\x6f\xcd\x27\x80\x88\x27\x03\x11\xaf\xbb\xb2\x24\x26\x89\x7f\x65\x87\x03\x1a\x62\xfe\x0c\x88\xb8\x5f\x8c\x0b\x48\x85\x6d\xda\x57\x8c\x93\x7d\xfc\x72\xcd\x94\xf8\x83\x85\xcf\xe5\xa0\xa8\x72\x2a\x08\x56\x24\x05\x96\x3a\x48\x31\x6d\xd3\xbe\x01\x51\x9e\x03\xa2\xc4\x83\xde\x72\xf7\xbe\x7c\xb4\xb2\xbf\x74\x96\x61\x1b\xf9\xb0\x04\xa2\xbf\xd8\x03\x11\x72\xf4\xde\x75\xa7\x98\xc6\x8b\x31\x7b\x7c\xc1\xff\x4d\x1c\x25\x57\x0f\xd1\x4f\x8f\x15\x96\x90\x4b\x3c\xa2\xc8\x90\xb0\x75\xa6\x7b\x6c\x73\x6c\x32\x5f\x84\x72\xf5\xa1\xcc\x20\x06\x04\x7b\xee\x36\x51\xb4\x47\xcb\xc9\xf1\xb1\xda\x8d\x78\x86\x7b\xd3\xc6\x64\x32\xd2\xe2\x2e\x00\x53\x8a\x98\xb2\x7a\x48\x7e\x7e\x9c\xea\xa4\x65\x7b\xc6\x02\x2c\x01\x2c\xe9\x84\x25\xbd\x89\x61\x25\x5b\x70\x3a\x68\x6b\x71\x1a\x04\xe3\x6b\x98\x23\x60\x30\x28\xda\x5d\xfb\xae\xf5\x09\xd0\xb4\x05\xcc\x03\xcc\x7b\x56\x98\xf7\xe2\x4f\xc7\x81\x96\x6b\x38\x1d\xf7\xb3\x64\x5f\x8f\x7e\xde\xb5\x02\x69\xd3\x16\x40\x1a\x40\xfa\x45\x1b\xa6\x2d\xce\xdc\x70\xe4\xb6\x3a\x72\x1f\x27\xbb\xca\x22\xd0\x99\x74\x0e\x81\x4e\xc0\x3b\x38\x88\xdb\x26\x98\xf5\x05\x28\x48\x1f\x6b\x1b\x97\xb5\x47\xc1\x28\x2e\x9b\xb2\xac\x39\x30\x9b\x76\x0e\x08\x08\x08\x08\xc7\xf2\xfe\xc7\xf2\x12\x71\x9d\xe0\x58\x9e\xca\xfc\x50\x51\x6b\x9b\x79\xbc\x94\x14\xdf\x2e\xd6\xee\xea\x21\xf9\xf9\x71\x32\xcb\xd7\xd6\xe6\x05\xb4\x07\xb4\x7f\xe6\x68\x6f\x47\x4e\x22\xa3\xb3\x37\xbf\x01\xcf\xa7\x0b\x84\xd9\xc3\x6f\x14\x08\xb3\x43\x60\xd3\x18\x40\x18\x40\x18\x40\x78\x3c\x10\x7e\x96\x27\x80\x37\x6d\x4e\x00\x03\x45\xe6\x40\x63\xb4\x0a\xcc\xd9\x2b\x8d\x28\x30\x67\xa7\x34\x4c\x63\x50\x1a\xa0\x34\x40\x69\xcc\xc7\x72\x6f\x85\xc7\xcf\x3d\x52\x38\x8c\xeb\xc4\x25\xbe\xc7\xef\xf7\x84\xa9\xf6\x91\xc2\xb3\xe4\x5b\x2b\x87\x49\x66\x28\x9b\x50\x61\xda\x3b\xc4\x0a\x01\x81\x01\x81\xe7\x71\x61\xb4\x28\xc4\x7d\x01\x31\x95\x72\x08\x4d\x5a\x86\x26\x5b\xc0\x6e\x14\x9b\x4c\x99\xd6\x1c\x9b\xcc\xf4\x0e\x90\x0b\x90\x0b\xc1\xc9\xfe\xae\x89\x32\x81\x9d\x20\x3a\x99\x4a\xfd\x50\xd1\x49\xab\x89\x40\x36\x5e\x9d\x91\xbd\x7a\x48\xff\xf3\x38\x9d\xc5\x6d\x6d\x6b\x03\xe8\x03\xe8\x83\x9d\x9d\x43\xcf\x99\x06\x29\xab\x09\x04\x58\x1f\x31\x4a\xd9\x02\x82\xa3\x30\xa5\x1d\x0a\x9b\xc6\x00\xc4\x00\xc4\x00\xc4\x63\x02\xf1\xf3\x3c\x0d\xbc\x69\x73\x1a\x18\x28\x52\xf9\xd2\x4f\x03\xad\x63\x95\x2d\x54\x47\x14\xac\xb4\x53\x1d\xa6\x31\xa8\x0e\x50\x1d\xa0\x3a\xe6\x64\xc3\xb7\x02\x65\xb8\xd8\xd8\xcb\x95\xb2\xa2\x7b\xbc\xcd\x62\xf0\x58\xd6\xbc\x1e\x07\x51\x96\x0b\x67\x3a\x9c\x29\x4c\x19\x11\x6d\x2c\xfc\x73\x4d\x31\x60\x35\x60\x35\x60\xf5\xf3\x33\xf3\x8d\xb4\x1b\x19\x07\x3b\x1f\x54\x55\x46\x55\x09\xa2\xdf\x72\x91\x13\x68\xab\xac\x92\x8a\x87\x45\x0e\x0f\xda\xf9\xa2\x2e\x63\x82\x41\x59\x81\xb2\x02\x65\xf5\x5c\x95\x55\x22\xe6\xa0\xaf\x5e\xa2\xbe\x9a\xfa\x49\x1d\x8b\x1c\xd0\xa8\x63\x48\x00\x05\xdd\x03\xc5\x62\xaa\x50\x29\x2f\x51\x7d\x71\xe9\x25\x3e\x19\xd4\x2b\x1d\xd3\x16\xfc\xa2\x5c\xcc\x88\x5b\xcd\x89\x98\x71\xbf\x80\x7b\x80\x7b\x90\x85\xd9\xdf\xc6\x3d\x90\xd3\x09\x52\x30\x23\x61\x1f\x2a\xff\xb2\x79\x0a\x60\xd6\x56\x9b\xb5\xab\x87\xe8\xc7\xc7\xd9\x19\xb8\x00\xf2\x00\xf2\xe0\x58\x49\x01\x73\xee\x96\x36\x00\xf9\x54\xe9\x96\xb6\xc8\x1b\x47\x67\x9b\xad\xeb\x28\x30\x03\xc0\x0b\xc0\x0b\xc0\x3b\x0a\xf0\xbe\x68\x53\x7f\x20\x27\xf6\x8b\x36\xf5\x5b\x67\x56\xda\x6a\x89\x28\xad\xd2\x42\x4b\x98\x96\xa0\x25\x40\x4b\x80\x96\x98\x87\x79\xfe\xc6\x1a\x82\x21\x95\xd2\xca\x31\xe2\x73\xb7\x7d\xa8\x6f\xcd\x5d\x2b\x2f\x48\xd8\xb9\x8d\x03\x64\xcd\x5d\x88\xee\x01\xc2\x02\xc2\xce\xa3\xbc\x8b\xcf\xdd\x41\xa2\x88\x6b\xee\xbe\xdc\x00\x62\x1f\x40\x5e\x3d\xf8\xdc\x7d\x1c\x13\x96\x9b\x01\x19\xc0\x18\xc0\x18\xc0\xd8\xa0\xe1\x4c\x3d\xd1\x25\x94\x75\x80\x68\x28\x31\x6b\xe3\x5c\xb0\xc1\xd6\xc8\xb1\xd0\x00\xaf\xa6\x15\x20\x2c\x20\x2c\x20\xec\x71\x11\xf6\x8d\x95\x25\x0c\x25\x64\xdb\x99\xad\x2b\x8f\x6f\xc7\x34\x5d\x91\xc7\xb7\xd2\xc2\xa1\xc0\xb7\xe0\xb2\x7d\x09\x08\x9b\xb2\xa4\x9e\xe4\x6b\x7f\x2b\xb0\xfb\x64\xc8\x3d\xe5\x8c\x11\x47\x63\xc7\x13\xa1\xf8\x8a\x38\xcb\x3f\xc8\xed\x15\x77\xbe\x10\xb5\x7c\x47\xee\xa7\x23\x9c\x32\x45\xb6\x87\x9d\xa6\x94\xd3\xc3\x4b\x3a\x05\xb2\xa5\x21\xfb\x77\x22\xe4\x0c\xd6\x1c\xcc\x86\xe9\xcc\x86\xde\x84\x6c\xb8\xe7\xf1\xaf\xb6\xfe\xba\xde\xc3\x29\x4c\xbd\xc9\x06\x4b\x8a\x1a\x4c\x36\xa2\x2f\xc8\x1d\xe5\x81\x1c\xc7\xff\xf9\x17\x8b\x28\x7e\x6c\x64\x80\xc9\x57\x6b\xf2\x49\xe2\x08\xd2\xe1\xc5\x80\x2b\xfd\x9d\x95\xb9\x17\x0d\x61\x13\x43\x32\xbd\x42\x18\x09\xce\xd5\x70\x49\xac\x0a\x8e\xb2\xe2\xd4\x17\x8e\x8c\xbc\xc1\x05\x31\xcb\x0b\x62\x96\xa0\x17\xdd\x0f\x33\x8c\x6a\xbe\x1e\x16\xf5\x0a\x80\x07\x80\x07\xb7\xc3\xfa\xa7\x8c\x1a\x71\xfa\x83\xaa\xdd\x3a\x14\xa3\x49\x53\x47\x8d\xc8\x0f\x75\x49\xcc\x7e\x26\x90\x13\x55\x65\xd8\xae\x1e\xcc\x0f\x8f\xe3\x5b\xb8\x56\xb6\x2d\xc0\x3c\xc0\x3c\x38\x7e\x12\xa8\x9c\xb5\x8d\x0d\xf8\x3d\xa4\x85\x1d\x0c\x87\xb9\xd1\x15\xb1\x66\xd8\x35\x0d\x01\x79\x01\x79\x01\x79\xc7\x40\xde\x67\x6c\xec\xbf\xb1\x35\xf6\x07\xba\x26\x06\xc6\x7e\x97\x8c\x2e\x4b\x85\x11\x25\x75\x35\x2b\x0c\xd3\x10\x14\x06\x28\x0c\x50\x18\x73\x30\xd5\xad\x51\x18\x12\xbc\xec\x9c\x22\xe2\x8e\x3a\xa4\x4b\xb8\x4f\x7f\x68\xe9\x0d\x31\x83\xd8\x05\xfc\x74\x5b\x88\xf8\x01\xdc\x42\xc4\xaf\x1a\xe2\x32\x02\xd5\xdf\xce\x34\x9d\x41\xcc\xcf\x36\xe6\x67\x87\x7c\x49\xd0\x4f\x37\xb7\x89\xfa\x99\x7e\x01\xf5\x00\xf5\x20\xec\x37\x84\x27\x40\xcb\xd3\xb1\xe2\x7e\x7a\xf0\xe1\x02\x7f\xd6\x73\x81\xc8\x5f\xa5\x91\xbb\x7a\x88\x7e\x7a\x9c\x9b\xb9\x0b\xa0\x0f\xa0\x0f\x9e\x85\x14\x37\xe7\x6d\x77\x03\x9a\x1f\x21\x0e\x68\x87\xbf\x49\x20\xb0\xd1\xe8\x8e\x23\x81\x80\xbf\x80\xbf\x80\xbf\x63\xe0\xef\x73\x3e\x01\xbc\xb1\x3e\x01\x0c\x16\x0d\x84\x13\x40\xa7\x70\xa0\x9d\xde\x48\xe2\x81\x8d\x7a\x23\x0e\x08\x82\xde\x00\xbd\x01\x7a\x63\x0e\x76\xbb\x3d\x16\x43\x4c\xd0\xca\x5d\xc2\xbd\xc0\x90\xf6\x10\xff\x98\x7b\xf1\xb4\xb5\xf3\xa4\xc5\xdb\xa6\xda\x7f\x12\x0d\x9a\x7b\xe1\xd4\xc2\x97\x92\x0e\x73\x15\xf5\x00\x51\x44\x80\x68\x80\xe8\x18\xa2\x23\xa1\x98\xa9\x6f\xa5\x28\xec\xc3\x3d\x15\x0a\xd5\x2a\x07\xd1\x00\xd3\xa4\x8a\xc4\xd8\xdf\xc1\x89\x0e\xa8\x0f\xa8\x0f\xa8\xff\xb4\x50\x1f\x32\x59\xd0\x31\xeb\x5f\xb6\x75\x8f\x14\xd1\xd9\xda\x51\x22\xaf\x0e\xf7\x21\x00\x33\x00\x33\x00\xf3\xbc\x5d\x26\x12\x7c\x26\x6d\x2c\x66\xc9\x32\x16\x73\x85\xcb\xc4\x16\x9b\x5b\x78\x4d\x8a\xf0\x9c\x19\xb9\x19\xa1\x33\x07\x25\x00\x69\x00\x69\x00\xe9\xa7\x03\xd2\x59\x9f\x09\xbc\x8c\x64\x05\xd4\x77\xdc\x0b\xf6\x59\xf4\xb5\xcc\xfa\xfe\xdd\x7c\x67\x99\xf4\x6d\x46\x69\xce\xf9\x8e\x7a\x05\xb8\x05\xb8\x85\x94\xef\xfe\x09\x1f\x45\x21\x9d\x20\xd1\xdb\x48\xfa\x50\x79\xde\x8d\x13\x00\x44\xaf\x40\xf4\xd5\x83\xf9\xf7\xb1\x8b\xa9\x6d\x09\xee\x91\x99\xdd\x0c\xee\xa6\x21\x80\x3b\x80\x3b\xd8\xd2\x09\x39\x77\x45\x71\x98\x93\x25\x1d\xc1\x38\x38\x3b\xec\x11\xb7\x7d\x20\xd0\x12\x66\x7f\x25\x2a\xc2\x58\xab\xe8\x9f\xe9\x15\x82\x7e\x00\xb5\x70\x61\xbc\x1e\xde\x86\x89\xb2\x19\x79\x83\xa4\x8a\x8e\xa8\x59\x6a\xa8\x1e\x1d\x3e\x01\x3a\x01\x3a\xc1\x4a\x3d\xae\x95\x6a\x0b\xe3\xe0\x64\xe8\x9f\x20\x61\x75\xe9\xd0\x12\x70\xa3\x3b\x87\xcd\x6e\x01\xd3\x10\x00\x17\x00\x17\x00\x77\x0c\xc0\x7d\x7e\xfe\x67\x6b\xc7\xc5\x40\xb7\x0c\xc1\xff\xfc\x2a\xfa\x76\x91\x19\x34\xa1\x6e\xf1\xf6\x0e\x53\x0f\xdf\x7a\x49\xae\x71\xde\xdb\x9c\x9b\xf5\x41\x5b\xb4\x5c\xa6\x29\x1a\x5f\x77\xd4\xd9\xa1\x10\x0a\x91\x83\x19\x12\x41\x1a\x63\x4e\x04\x87\xdf\xe6\x50\x23\x84\x11\x9f\x08\x45\x0b\x3b\x65\x81\x1d\x45\xef\x48\x71\xf7\xc4\xbd\xdc\x72\xee\x11\x9c\x0f\x61\x2f\xbe\x2d\xb7\x7c\x19\x0b\xc5\x5b\xf3\x7d\xe9\xe2\x51\xb7\xaa\xdf\x12\xb1\xce\x77\x7b\x7e\x56\xd1\xe5\x1e\x6f\x0f\x37\x7b\xdc\x2d\x16\x02\xe7\x05\x71\x41\x15\xd9\xcb\xc3\x5d\x56\x20\xa4\x72\x9b\xe6\x89\x32\x83\x97\x12\xe6\xd1\x3d\x55\x07\x84\x55\xe4\x9a\x26\xc9\xe6\xfa\x9b\xd2\xfe\xa2\x31\xbb\xad\xde\x07\xbc\xaf\x60\x49\x20\xbc\xce\xbd\x5e\x5f\xbe\xcf\x74\x7a\x60\x16\xe9\xb6\x3e\x76\xbe\xe0\xad\x6e\xbe\xa5\xea\x24\x79\x2b\x30\xd8\x9f\x30\xa2\x56\xce\x6e\xf5\x25\xb8\x25\x4b\xec\xd3\xd5\x1d\x61\x2e\x17\xab\x2d\x55\xbb\xe0\xf6\xc4\xe1\xfb\x55\xa6\xb5\x69\xe6\x78\x94\x30\xb5\xf2\xbf\x6c\x57\x7b\xee\x12\x6f\x91\x3b\x3f\x1f\x08\x89\xf1\x2b\xb4\x90\x2a\xfd\x41\x28\x5a\xe1\x59\x17\xf1\x0d\xc2\x71\x8b\x44\xd8\x64\x77\xd1\x4a\xbb\xe8\xbd\x5d\xcb\xf7\xd1\x21\xa2\xd8\xed\xe3\x64\xfa\xb3\x62\xe7\x29\x67\x1b\xba\xfd\x85\x7a\xb9\x0c\x8b\xb6\x8b\xee\xe8\x5e\x6e\x36\xf9\x6e\x86\x12\x9f\xa8\x8b\x6e\xdd\x7e\x0c\x1b\xce\x6f\xc5\x2f\xb0\x5f\x23\x33\x49\x9b\x50\x4c\x74\x1f\x68\xc3\x05\x32\xcb\x8c\xf6\xd8\x6f\x96\x8f\x8c\xd5\xf5\xaf\x02\xc0\x65\x96\xd6\xc5\x0a\xc7\xab\xf3\xb9\x99\xcd\x26\x66\x7e\x83\x8b\x16\xc4\x81\xfd\x10\x36\xd4\x79\x8d\xe1\x39\x8b\x32\x74\xf9\xcb\xe9\x5f\xff\xfa\xd7\xff\x45\xd1\x91\xe7\x75\x27\x56\x9a\x0c\x1c\xf7\xad\x2a\xdf\x26\x7a\x2e\x56\xaa\x20\x59\xde\xb3\xcc\xf4\x0b\x9d\x99\xb8\x52\xf3\x54\x4d\xc3\x61\x27\x6a\xa2\xd1\x95\x13\x1d\x49\x4b\xf1\xaf\x8c\x88\xce\xfd\xfe\x9f\xfe\x7a\x96\x82\x76\x96\xdf\x18\x95\xc2\x16\xb6\xab\x12\x38\xa4\x77\x57\xa3\xd4\x61\xd7\xd5\x7b\x0c\x7b\xeb\x0a\x19\x2a\xb7\x81\xe6\xb0\x4c\x4d\x8a\x3c\xdf\xb0\x6a\xa1\x0a\x21\x8b\x8e\xaa\x24\xec\x68\x2c\x05\x9e\x02\xb0\x9d\xe2\x4e\xe7\x3d\xb7\xed\x1d\xbd\x08\x5d\xcb\x32\xd3\xe6\x80\x5b\xd1\xaf\x69\x36\xf5\xbc\xab\x52\xd1\xc7\x83\xc5\xeb\x6a\x2d\xe3\xe5\x4c\x6e\x0b\x3d\xc3\xf7\x7b\xcc\xdc\x63\x9d\x38\x4e\xe3\xe1\x4b\x61\x32\xb2\x75\x46\xde\xa1\x86\x3f\x87\xa7\xfb\xb6\xfb\x34\x43\x38\x61\x77\x63\x11\xfc\x33\xbb\xb3\x24\x32\xd7\xf2\xe0\x8c\xd9\xfd\xe4\xaa\xbf\xee\x7d\x46\xbc\x24\x92\x07\xc2\x21\x93\xea\x5d\x9f\x0b\x35\xfe\x56\x5a\x73\xa1\x2c\x79\xb4\xd6\x04\x95\xd2\x6a\x9c\x57\x37\x7b\x1e\x30\x35\xb3\xed\x6f\x1a\x5f\x18\xca\x66\x09\xd4\x9a\x05\x16\x60\x1d\xb6\xcb\x03\x76\xb8\x45\x42\xac\x4e\x46\x1e\xca\xfe\x0f\x3b\xce\xfd\x5f\x70\xc5\x1d\xee\xd9\x63\xf5\x88\x42\x51\xd5\x2d\x65\x8a\x6c\x0b\xf1\x80\x4c\x30\x85\x32\xf5\xf7\xbf\x2d\xea\xb7\x77\xc5\xa0\xf1\xec\xad\xc0\x62\x5d\x58\xab\x79\x6d\xb6\xdf\x0f\xae\x4d\x54\x6d\x37\xd3\x32\xdc\x70\x91\x6b\xfa\xfb\x9c\x41\xf7\x03\xd2\xd2\x4e\xdc\x41\x36\xa0\xee\xeb\x46\x87\x09\xac\xb7\x58\x38\xcd\xce\x5b\xec\x22\xfc\xb8\x94\xdb\x19\x52\xba\x77\x1e\x30\xb5\xce\x4c\x66\x12\x85\x21\x83\xdb\x7e\x64\x5f\x05\xb7\x05\xa2\x67\xb0\x71\xb3\xd7\x6b\x6a\x76\x6d\xa6\x19\x5a\x2e\xbb\x9f\x33\x3c\x7c\x4b\xba\x7b\x65\xdf\xeb\xaf\xcb\xbd\x55\x58\xd0\xcd\xe6\xa6\x87\x1b\xfe\xa3\xee\x21\xe7\x8c\x3f\x3e\x7f\x32\x97\x5b\xab\xb9\x93\x36\xca\xeb\xaf\xec\xc5\xda\x8e\xc0\x91\x90\x2b\xeb\x4e\x1a\x82\xf8\x1e\x75\x70\x8b\xb3\xc6\x98\xa1\x98\x0c\xcd\x63\x9b\x4a\xf6\x67\x84\x98\xa2\x72\x8a\xe7\xed\xe1\x03\xa7\x5c\x3f\xa7\x5c\x99\xa4\x8c\x63\x64\x5d\x16\x04\xb1\xa0\xc1\xa2\x58\x48\x09\x4a\x16\x38\x99\x84\x62\xcf\xcf\xd0\xf7\x9c\x79\xf7\x88\x6e\x32\x70\x82\xa8\x44\x3e\x16\x3a\x8c\x14\x77\xfa\x43\x47\x06\xc7\x43\x55\x45\x41\xa5\xc2\x2a\xb0\x3c\x48\x66\x2a\x1b\x9a\xaf\xca\x55\x05\x57\xd8\xbb\x71\xfc\xa0\x61\x11\x74\x3b\x74\xba\xbe\x46\x81\xc4\x5b\x82\x6e\xef\x11\xf6\xbc\xd4\x18\x93\xe1\x26\x57\x3b\x2a\xcb\x9c\x39\x6d\xb8\x1a\x34\xb1\xf5\x63\x48\xc9\xe9\xfa\xba\x6e\x3e\x7b\xb2\xe7\x3a\xcf\xaa\x79\x4a\x97\x6f\x2f\xe6\x31\xa5\x0b\x43\x73\xf9\x99\x97\x08\x99\x4f\x1c\xa8\xcb\xc7\x88\x1a\xcf\x53\x75\x5f\x15\xb7\x70\xb5\x02\x37\x4d\x43\x35\x1e\x8e\x20\x18\x51\x44\x22\x23\x02\xa1\xb0\xb5\xf1\x1a\x56\xaa\xde\x38\x92\x7a\x33\x32\x18\x25\x21\xdb\x7a\x54\x12\x04\xbb\xf7\x37\xa3\x03\x23\x76\xef\x9b\xe8\x38\x22\x34\x07\x6c\x32\xb6\x5c\xa7\x43\x35\xd0\x64\x72\xa5\x46\xa7\xc7\x0c\x53\x42\xcb\x9c\x64\xf8\xf7\x03\x38\xaa\x16\xe2\xa8\x6d\x95\x31\x8e\x22\x68\x8b\x92\xd1\xba\x4b\x72\x05\x44\x5a\xab\xde\x79\xa3\x66\x53\x60\xac\xd0\xb2\xf2\xe8\xd3\x33\x32\x56\x5e\x9d\x68\xd0\xe3\x44\x59\x05\xa3\xda\xf3\x44\x66\xea\xb3\x62\xde\xcf\xb9\x58\x47\x91\x61\x3f\xb3\x3b\xad\xd8\xc8\xfd\xf2\x0e\x7b\x01\x41\x3e\xa6\x22\xd4\x6a\x84\xdd\x51\xc1\x99\x91\x0d\x2c\x68\x08\x4d\x9d\x4f\xac\xba\xeb\x83\xc3\xea\xb1\x1d\xab\x86\xaa\xce\x82\xaa\xbf\x9e\x15\xa7\x45\x5d\x00\xf4\x67\xa1\x43\x9f\x52\x61\xe6\x62\xe1\x22\x49\x04\xc5\x1e\xfd\x8f\xce\x36\x7b\xbb\x3e\x37\x29\xae\x9f\xd8\x05\x91\xda\x0c\x5d\x2e\x43\x13\x34\x6c\xae\xcc\x9f\xd0\xde\xfc\xe5\xa7\x4f\xec\x7f\xd0\xa7\x05\x65\x77\xd8\xa3\xae\x4e\x03\x0d\xd7\xe4\xd3\xc2\xfc\xfe\xcf\x80\x2b\x8c\xc8\x37\x87\x10\x97\xb8\xf1\x6f\x75\x5b\x73\x4c\x36\xe3\x2c\x3e\xb1\x93\x93\x13\xa2\x9c\x93\x93\x93\x4f\xec\xfc\x2c\x1c\x2f\x60\xf4\xcf\x80\x44\xa3\x51\x97\x30\x45\x37\xd4\x31\x5f\x39\xdc\x25\x9f\xd8\x19\x51\x98\x7a\xda\x16\xe3\xbe\xc9\x2e\xd0\x86\x32\xf9\x56\x20\x52\xa2\x2f\x94\xb9\xd8\x0c\xbe\xa1\xc4\x73\xd1\x77\xb1\x36\xfb\x0e\xed\x03\xa9\xd0\x2d\x41\x8c\xb3\xe5\x7f\x88\xe0\x48\x6f\x85\x98\x56\xc6\x15\x22\x8c\x07\xdb\x1d\x52\x74\xbb\x53\x12\x29\x8e\x36\x84\xb8\x68\xcb\xfd\x1d\x11\x71\x3b\x11\x85\xc8\xd0\x77\xbf\x72\xf7\x3b\xe4\x72\x22\xbf\x53\x88\x7c\xa3\x52\x85\x4d\x7e\x09\x47\xcd\x93\x2a\x89\x3e\x9e\xe5\xa5\x4d\xf6\xc1\x41\xbd\x1c\x47\x0a\x0e\x47\xcc\x28\x17\x2e\xbd\xe6\x96\x47\x44\xb3\x52\xb6\x39\xbc\x15\x71\x57\x21\xaa\x4e\xab\xd1\x96\xe8\xee\xda\x8e\xbe\xaf\x39\x0a\xdf\xec\x94\xf2\x47\xb2\xc5\xcc\xf1\xe3\xb7\x8f\x1f\xd7\xd3\xa1\x8d\xc9\x7a\x3f\x80\x97\xf3\xb3\x7a\x80\x31\x62\x2c\x88\x2f\x88\xd4\x6a\x3e\x27\xd1\x99\x7b\x3b\xad\x77\x7a\x28\xcd\xd6\x1b\xe1\x5d\xd8\xb8\x9c\x5b\x2d\xb6\xd3\xd5\x94\x6e\xe7\x8a\x05\x7f\x97\x9f\x76\xc9\x92\x87\x2d\x0a\x8b\x1e\xae\x95\x56\xe1\xb9\x3b\x0b\x95\xdb\xb1\xda\x43\x30\xcd\x1c\xaf\x9a\x76\xd5\xd5\xc1\xb6\x8a\xdf\x44\x39\x3f\xab\x99\x67\x41\xae\x27\x98\xcc\x2f\x45\xcc\x2b\x4e\x26\x55\x09\x99\xc9\xa4\x89\x72\x19\xc5\x60\xe0\xb3\x66\x76\x73\x4f\xb7\x2b\x5b\x9f\xf3\x42\x62\x4b\x6b\x14\xf8\xd0\x6c\x08\x96\x8a\xfd\x47\xbc\xb5\xfe\xec\xf8\x76\xdc\x39\xdb\x0a\x22\xeb\xf6\x51\xd4\x22\x7f\xbc\xa2\xd1\x67\x03\xe5\x43\x88\xc0\x23\xf2\xd9\x24\x44\x43\xb8\x64\xb8\x70\x89\xde\x18\x23\x1d\xbb\x2f\x03\xcf\x36\xcb\xe9\x32\xbb\x41\xe7\x24\xb8\xa4\xc9\x4b\x92\x6b\x57\x2a\xc4\xa4\xaf\x87\x24\xe9\x67\x2c\x46\xc5\x20\x65\x79\x73\xad\xd8\x7a\x06\xdc\xfa\x90\xdc\x25\xae\xe6\x54\xd2\x26\xe4\x52\x72\xf9\x38\xd5\xde\x38\x57\x60\xb7\x25\xe2\xc6\xe7\xc7\x56\x21\x7a\x27\xaf\x19\x6a\x38\x74\x2d\x89\xf8\x55\xf0\xc0\x7f\x6b\x3e\x7a\x82\x51\xee\x11\x2e\x70\x96\x26\xbb\x14\xe6\x1a\x48\x22\xcc\x85\x5b\x8f\x20\xfd\x81\x96\x4d\xb5\x23\xa8\xfc\x02\xfa\x20\xd9\x32\x7b\xfc\xed\x86\x7c\x53\x37\xf1\x73\x84\xdd\x4e\x93\x8d\x41\xbe\x0b\xfc\xed\xe7\x6f\xea\xe0\xf1\x8c\x22\x25\x94\x4d\x41\xc9\x39\x6b\xa6\x44\x09\xbc\xd9\x50\x67\x44\x2a\x3e\x46\x23\x4c\xab\x47\x75\xc7\x37\x1e\xdf\x52\xd6\xaf\xfb\xf7\xba\x8b\x8a\xf8\x59\x0c\x31\xad\x92\xbe\xe5\x84\xf9\x5c\xd9\xe0\x96\x1c\x31\x3f\x28\x84\xc3\x22\x12\xd6\xea\xac\x6b\x99\xcf\x0d\x9a\x91\xc6\x6a\x32\x30\xf2\x0d\xf3\x16\x46\x5a\xe1\xab\xa7\x89\x91\x76\x34\x16\xcb\x52\x0d\x6d\xc7\xb1\x74\xda\xb3\x62\xdb\xfa\xb2\x39\x50\x99\xb4\xc9\x79\x06\xd0\x5a\x90\x4b\xe2\x11\x2c\x09\x8a\xfb\xe8\xcc\xaf\x73\xf9\x41\xd7\xb4\xa9\xcd\xee\x2b\x95\xcb\x68\xe8\x9a\xcf\xed\x40\xb8\xae\xf3\x2b\x25\x8e\x79\x44\xbf\xf5\x30\xdb\xae\x24\xd9\xdf\xc5\x70\x9d\x70\x2f\x9f\x5e\x7c\xc0\x38\xac\x76\xa1\x80\x45\x06\x37\xca\x95\x7d\x69\x69\x10\x16\x4a\xc6\xc4\xef\xd8\xdc\x14\x8f\xe6\xf1\xef\xfd\x4c\xfa\x7e\xb3\xed\xd8\x2b\x4f\xba\x3a\xb3\x3b\x47\x64\xe7\x34\x6c\xd3\x49\x4d\xa2\x77\x76\xca\x23\xf9\xd9\xcd\x10\x85\x3b\x11\x33\x80\x8f\x50\x40\xa4\x22\x2c\x2a\x77\x68\x94\xd8\x45\xfe\x26\xc0\x44\x5e\x4f\x1b\x72\x79\x9d\xc7\x7a\xcd\xdd\xe2\x7d\x1e\xb7\xd7\x0d\x59\xc8\x2a\x1e\xd4\x4d\xe6\x7b\xfc\xbe\xb3\x14\x9b\x14\x88\x9a\x4b\x8d\x37\x7e\xe0\x79\x37\x92\x38\x82\x54\x4a\x71\x81\xf7\x76\xce\xed\x9e\x35\x75\xd6\x81\xe7\x5d\x19\xaa\x9e\x85\x8b\xae\x4d\x7e\xee\x9a\xbb\xd3\x26\xe6\x66\x25\xfe\xa9\x67\xe4\x8e\x3b\x97\x83\x54\xdc\x19\x28\xa3\x64\xb7\xd4\x62\x7c\x43\xca\x6c\x2f\xd0\xf7\x77\x58\x76\x97\xc6\xb5\xfe\xba\xea\xb0\xac\xb0\x50\x37\x0e\x0f\x98\x1a\x2d\xe3\x54\x8f\x71\xaa\x87\xa8\x92\x5d\xa1\x9a\xf5\x83\xcf\x5d\xa4\x9b\x0e\xab\x20\xae\xc2\x2e\xdf\xaa\xb9\xed\xb9\xa6\x03\x6f\xdc\xe4\xc0\xb6\xe8\x7b\xc8\x0d\xbb\x18\xcb\xb6\x08\x2d\x25\xdb\x6b\xe8\xee\xcc\x8e\xb4\x87\xd7\x83\x8b\x2c\xc9\x9c\x63\xf5\x9d\xed\xf8\x46\x31\x0a\x97\x0f\x7d\xff\xf1\x74\x8d\xb8\x40\xd7\x67\xeb\x1f\xa6\x0e\xdd\x5b\xcc\x2f\xa9\x7e\x50\x3d\xbf\xb8\x49\x21\x9a\x9f\x06\x09\x42\x5d\x88\x99\x1b\xea\x91\xce\x37\xfd\xfc\x60\x91\x4b\x64\xca\xea\x03\x8b\x98\x6c\xa3\xd6\x0e\x69\xa4\x0c\xed\xc7\x51\x62\x95\xba\xd8\x4a\x0b\x87\xfa\x97\x32\x74\x41\xc7\x21\x6e\x86\xca\x35\xde\x52\x26\xfb\xbe\x90\xec\x50\xb5\xff\xd2\xc6\x51\xca\x66\x38\xa2\x44\x8c\x7c\x45\x49\x82\x62\xce\x81\x70\xe4\xfb\xc1\x73\x5b\xe6\x8c\x5f\xd8\x76\xb1\xd3\x4f\x72\x4b\xae\x0b\xa2\x9a\x08\x9d\x65\x8d\xac\xc9\xc2\x7c\x71\x92\x6e\xf7\xda\x9b\x71\x07\x73\x64\xa4\xb4\xc9\x29\x52\x54\x79\x24\xcb\x4b\x59\x89\xdc\xa6\x4e\x8e\x06\x6f\xcd\xd3\x24\x8a\x72\xd2\x04\xd8\x3b\x2c\x5c\x7b\x7c\xd6\xad\x07\xa8\xc6\x13\x48\xd2\xbd\x9f\x19\xb0\x31\xf0\x6a\x81\x2e\xf0\x48\xd6\xc3\x2a\x02\x8f\x74\x55\xa8\x3b\x5e\x7c\x69\xa9\x4d\xf5\x8d\x5d\xf1\xb5\xd5\x36\x12\xf4\x1b\x97\x55\xb5\x56\x6a\x1c\xb3\xbd\x6d\xcc\x9c\xdf\xb6\xde\xc8\xac\x74\xf1\x2a\x4f\x36\x38\x6d\x9a\x03\x7f\xef\xaf\x0e\x1c\x2c\x73\xd8\x79\x2c\xbe\xea\x7c\x19\x15\x32\xaf\xdd\x88\x07\xad\x0d\x84\x44\x3f\x2b\x8e\x44\xc0\xd2\x32\xd9\xe1\x16\x24\xbd\x0a\x00\x46\x0f\x73\x8e\x73\x14\x8d\x7c\x81\x15\x15\xd2\x84\xe0\xe2\x58\x17\x0b\x7e\x36\x83\x97\xbb\xe2\xb8\xba\x19\x77\x5d\x3e\x70\x75\xb8\x34\xc7\xdf\xab\x57\x45\x01\x2c\x6e\xcf\xf4\xdc\x1b\xc9\x2a\x14\xc6\xb5\x2a\x8c\x7b\x24\xc7\x33\x14\xdc\x7d\xc6\x05\x77\x8d\xb0\xfe\x41\xd5\x6e\x8d\x05\xde\xd7\x48\x6d\xa1\x65\xde\x81\x65\x04\x19\x7d\xa5\x6a\x87\xcc\x12\x81\x4c\x83\x4c\x0f\x2d\xd3\x3b\xea\xba\x84\x35\xd0\xb5\xa3\x2e\x89\xf7\xe3\x46\xf0\x3d\x32\xc9\x5a\xaf\xbb\x95\x8a\xfa\xcd\x0c\xf9\x82\x21\xc6\x0e\x38\x1a\xdc\xde\x25\xad\xcb\x00\xa4\xaf\x13\x3c\xea\x65\xac\x33\x4a\x11\x2c\x2d\x1f\x91\x88\x88\x9a\x8a\x27\xd2\x8a\x17\xf2\x49\xf3\xe0\x78\x4b\x3f\x80\xc6\xf5\x88\xa3\x88\x6b\x5d\xc3\xbd\xfc\x83\xfa\x5a\xee\x06\xf9\xb0\xe7\x65\xb2\x28\x87\xb8\x9a\x57\xce\x13\x59\x4a\xe1\xdb\xb2\xde\xe6\xb3\xfe\xb6\x77\x60\x4a\xdb\xd7\xdd\x85\x39\xce\xda\xe7\xc8\x9b\xed\xd2\x17\xd3\xf4\x6b\x62\x71\x71\x4e\xff\x80\x55\x93\xe1\x96\xe0\x04\x46\xda\x88\xa9\x59\x2e\xdf\xe3\x1e\x77\x10\xce\xcc\xe7\xe5\x59\x5f\xfe\xd1\xde\x4a\x5b\xcb\xe7\x91\xcb\x35\x6a\x4d\xfc\xd2\x9c\xd7\xae\x15\xf1\x7b\x96\xb2\x8c\x2f\xf5\x8f\x5d\xc7\x72\x36\x80\xdd\x50\x04\x3f\x01\xed\x64\x65\x72\x25\xea\x7b\x02\xb7\xc2\x62\x4b\xd4\x0d\x54\xbd\x1f\xb4\xea\x7d\xc5\x02\x8f\x33\x8f\x8f\x7a\x84\xd9\xe5\xab\x47\xbb\xfb\x63\xfe\xa9\xba\x66\x93\x44\x27\x06\xcd\x30\x17\x28\x9a\x8f\x9d\x33\x2f\xdf\xb4\x78\x10\x34\xf3\x1c\xde\x9d\x07\x56\x18\x58\x61\x35\x56\x58\x1b\x07\x9f\xd9\xa2\xe3\x7a\xf8\xc0\x2a\x04\xab\xf0\xd9\x59\x85\x96\xda\xa3\xd1\x2d\x72\xd8\xbc\x54\x8b\xf4\xf7\x27\x9a\x6e\x46\xde\x43\xed\xbd\xba\x11\x59\x93\x71\x46\xda\x71\x44\x3e\x75\x4e\x1c\x91\x01\xfd\x4d\xb0\x48\x62\xeb\x6b\xd9\x66\x5a\xe5\x9e\xab\xef\x58\xca\xb6\x8a\x5f\x35\x8f\x07\x4e\x1c\x6e\xcc\xbf\x20\x38\x1f\x36\xbd\x2f\xbe\x33\x58\xc5\x29\xd3\x30\xc7\xac\x24\xd7\x32\xca\xc2\xec\x91\x3d\x75\x98\xfd\x6e\x6f\xa2\x56\x65\xae\xe7\xed\xff\x76\x9d\x5e\xbe\xbd\x98\x25\xbb\x4a\xd3\x68\x2b\x38\x96\x4b\x9e\x8d\x5f\xad\x3f\x64\x5e\x77\xae\x55\x56\x2c\x19\x4e\xc6\xca\x8e\xdd\x25\x07\x6f\xcb\x2a\x64\x25\x25\x53\x8e\xcf\x5a\x93\xa7\x7e\x50\x6f\xb1\xc0\xd4\x4c\xab\xbc\x66\x33\x45\xf0\xb3\x8f\xe3\xea\x36\xd6\x09\x8d\x0d\xef\x56\x1d\x14\x34\xa0\xd9\x37\x4b\x3f\xdb\x5f\xfb\xee\x57\x74\xa0\xe4\x6e\xf7\x78\xcf\xb0\xce\x65\x4f\x24\x35\x52\x1a\xae\x97\x94\xb4\x3e\xb8\x5f\x92\xc9\x98\xd7\x97\x3c\x3a\xee\x0b\x2f\x7b\x43\xe4\xf3\x8b\xbb\x8e\x62\x56\xfa\xf2\xf0\x41\x8f\x72\x96\xc4\x0d\x0f\xb9\x11\xec\x6f\x89\xae\xdd\x9f\x3c\x0e\xd2\xb9\x30\x5d\xdb\xa7\xe3\xa6\x7f\xa7\x65\x06\x7c\x2b\xbd\x38\xd4\xed\xaa\xcf\x8d\x47\xee\x88\x07\x17\x7e\x2c\x96\x5c\x4f\xbd\x4e\x4a\xe2\x36\xa1\x80\x6c\xf5\x0f\x7c\x53\xf0\x65\x1d\xff\x42\x16\xd4\x5d\x1c\xac\x6c\x22\x09\x61\x6f\xd4\xb2\x71\x9a\x9b\x17\x7a\x1c\x4b\xab\xf0\x22\x22\xaa\x8e\xe2\x5e\x75\x06\x6c\xae\xb7\x86\x83\xd4\x14\x1a\x98\xa2\xfe\xa0\x19\x23\x14\xbe\x3e\x55\x03\xf5\x28\x33\x8b\xe1\x16\xa5\xbc\x44\x05\x1c\xad\xd0\x7e\x1b\xf2\xa3\x6d\x6d\x81\xa7\xa6\x65\x8a\xaa\x66\x1b\xcf\x08\x53\x47\x40\xad\x67\xae\x46\x2f\x0e\xb0\xb3\x81\xfb\xb9\x63\xbf\xbe\x1b\x9b\xdd\x0b\x3d\x14\xec\x13\x41\xf1\x19\xf1\xce\x8a\x6b\x15\xfc\xea\xc1\xa8\x6d\x71\xec\x71\xf8\x64\xc9\xa1\x68\x25\xe6\xc6\xa0\xdf\x08\x76\x89\x38\xcb\x5f\x4c\xa9\xc9\x80\xd8\xe9\xf6\xba\x68\x80\xf6\xbf\xfc\x73\x19\xf6\xb2\x4c\xab\x8f\x63\xe6\xc6\xbf\x8c\x1e\x7c\x37\x9f\x48\xf4\x3d\x61\x0e\x77\x89\x1b\x9a\x86\xb7\x58\x92\xbf\xff\xed\x87\xae\xc7\x3e\x9a\xad\x65\xb5\xc8\x1f\xe9\x53\xbc\x3e\x5e\xad\xf2\xc6\x00\xe2\x8e\x4b\x45\xd9\x76\x19\x2a\x5e\xc1\xb0\x87\x0a\x7e\xa6\x19\xd4\x0f\x2f\xf3\x5e\x3e\x65\xdf\x85\x24\x22\x76\x0d\x37\xe0\x51\x12\xdd\x5d\x2e\x91\x08\x18\xa3\x6c\x9b\x38\x8e\xbb\xee\x58\x45\xf6\xbe\x97\x7d\xae\xb3\x24\x37\x86\x15\x8b\x08\x37\xef\xdc\x5b\x81\x99\xd3\xbd\x56\xea\x3f\xcc\xe7\xe5\xf7\xaa\x67\x1a\x3c\x1a\xd7\x8c\x1a\x29\xbb\x82\x95\xbc\xe0\xd0\xbe\xef\x42\x85\xe9\x5c\x0e\x5f\xbc\xbf\x3a\x97\x3d\x88\x3b\x28\xb7\x2e\x45\xf7\x13\xf1\xf5\xe5\xfb\xd9\x22\x41\x53\x48\xfd\xa0\x6d\xd6\x4a\x69\x8f\x0d\x95\x81\xf5\x78\x88\x31\xad\x95\x04\xfd\x2c\xc3\xeb\x09\x4d\x73\x62\x5e\x73\x7d\xf2\xd2\xea\xe4\xa6\x5c\x36\x72\xf8\xde\xc7\x4a\x2b\xbb\xbb\xbe\x45\xca\xff\x11\x50\xcf\x1d\x35\x39\xac\x54\x12\x2f\xf0\xbf\xb9\x18\xa1\xb8\xf9\x05\x65\xa3\xf4\xbb\xc6\xaa\x5a\x41\xf5\xe9\x57\x90\xd1\x2a\xd1\x5c\x1e\x3c\x6d\x9d\xe3\xc6\xa4\x45\xdd\x8d\x11\x5d\xb7\xdf\x8d\x95\xbd\x5c\xa2\x3b\xf3\x53\xdb\x27\x7f\x26\xb3\x8a\xa3\x00\xc2\x3e\x5f\x0f\xbc\x8e\x11\xd5\xf5\xc4\xcb\x5d\xd5\xd8\xc7\x0e\x55\xf7\x63\x79\x2a\x4f\xe3\xfe\x9f\x64\xc1\xec\xaa\x5c\xe5\x97\xfa\x3a\xdc\x14\x5e\xe5\xf2\xf2\xd6\x13\x64\x87\xd7\x15\xca\x96\x8a\x0b\xbc\xed\xfb\x0e\x82\xe9\xa4\x9a\x6f\xf0\x04\xcf\x91\x2d\x25\x0d\x99\x4d\x06\x6e\xa6\x55\x3e\xb1\xc6\x28\x93\xbe\x19\xa3\x51\x2f\x63\xb1\x29\x52\x8e\x76\x2c\x8a\xa6\x3a\x2b\x26\x11\x78\x08\x1f\x1e\xc2\x9f\xf7\x43\xf8\x83\x3f\x0c\x7d\x0c\x57\x52\x5e\xa6\x9b\x5e\x81\x46\x1d\x5f\xf6\x27\x42\x9c\xbb\xf0\xb2\x7f\x9b\xa5\xce\x00\xe1\xb9\x3b\xa7\x27\xfb\x7b\xbc\x5f\x3f\xc8\x99\xa3\xdf\x8e\x95\xd4\x62\x42\xed\x5e\xab\xb7\xdb\xaf\x1d\xc8\x2e\xdf\x46\x5b\xc2\x96\x1e\x37\x9a\x24\xa9\x19\x29\x17\x43\xec\xb5\x16\xd5\x6a\x2c\xd2\xe5\xba\xd6\x6d\x29\x94\xff\x69\xb5\x0a\xa5\xd3\x69\x2e\x00\x33\xe2\x6c\xf2\xc5\x74\x3a\x4c\xe6\x55\xd4\x7c\xe1\x63\x81\xf7\x44\x65\x8d\xfb\xc5\x29\x67\x8c\x38\xe1\x77\x26\x5e\x69\xf3\x9c\x58\x2a\x63\xc5\xbf\x64\xb2\xaa\xa3\x5e\xd3\xbf\x51\x13\x9e\x33\xa3\x94\xc5\x74\x94\x08\x48\xb1\xe6\xd6\x1f\xe4\xf6\x8a\x3b\x5f\x88\x7a\x47\xee\x87\x24\xf0\x8a\x38\xcb\xa4\xef\xe5\x3b\x72\xdf\x9f\x4e\xa9\xfb\x8a\xfc\x5b\xd5\xb4\xd6\x24\x04\x51\xa6\x2a\x29\x35\xbd\x2f\x0f\x1e\x80\x6c\x4f\xed\xb5\xbf\x15\xd8\x25\x43\xae\x66\xd4\x65\x2f\xaa\x24\x11\xe7\x67\x5d\x88\x0a\x82\x4c\xcc\x3a\x21\x29\x0a\x94\x67\x75\x57\x37\xa2\x2e\xb9\x57\xb2\x56\x84\xe9\x77\x30\xff\x95\x3f\x34\xe7\x22\xe5\xee\x3e\x75\x5d\x7c\x6e\xac\x68\x50\xa0\x3b\x1c\xb5\x13\xe5\x89\xb0\xc7\x15\xa9\x33\xb2\x6e\xb4\x4e\xb5\x25\x60\x10\xbd\xa0\x9c\x16\xd2\xd9\x91\x3d\xb6\x41\xaa\xf0\xc8\xf5\x2a\xab\x7a\x34\x35\xaf\x1e\xff\x3f\x00\x00\xff\xff\xb4\x36\xcf\x6c\x63\xc4\x01\x00")

func init() {
	if CTX.Err() != nil {
		panic(CTX.Err())
	}

	var err error

	var f webdav.File

	var rb *bytes.Reader
	var r *gzip.Reader

	rb = bytes.NewReader(FileSwaggerJSON)
	r, err = gzip.NewReader(rb)
	if err != nil {
		panic(err)
	}

	err = r.Close()
	if err != nil {
		panic(err)
	}

	f, err = FS.OpenFile(CTX, "swagger.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(f, r)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	Handler = &webdav.Handler{
		FileSystem: FS,
		LockSystem: webdav.NewMemLS(),
	}

}

// Open a file
func (hfs *HTTPFS) Open(path string) (http.File, error) {

	f, err := FS.OpenFile(CTX, path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// ReadFile is adapTed from ioutil
func ReadFile(path string) ([]byte, error) {
	f, err := FS.OpenFile(CTX, path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))

	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(f)
	return buf.Bytes(), err
}

// WriteFile is adapTed from ioutil
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	f, err := FS.OpenFile(CTX, filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

// WalkDirs looks for files in the given dir and returns a list of files in it
// usage for all files in the b0x: WalkDirs("", false)
func WalkDirs(name string, includeDirsInList bool, files ...string) ([]string, error) {
	f, err := FS.OpenFile(CTX, name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	fileInfos, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		filename := path.Join(name, info.Name())

		if includeDirsInList || !info.IsDir() {
			files = append(files, filename)
		}

		if info.IsDir() {
			files, err = WalkDirs(filename, includeDirsInList, files...)
			if err != nil {
				return nil, err
			}
		}
	}

	return files, nil
}
