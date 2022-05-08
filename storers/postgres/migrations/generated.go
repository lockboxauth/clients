// Code generated by go-bindata.
// sources:
// sql/clients_20181208_1_init.sql
// sql/clients_20190816_1_add_name.sql
// sql/clients_20190920_1_unique_uris.sql
// DO NOT EDIT!

package migrations

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _sqlClients_20181208_1_initSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xc4\x92\x4f\x4f\xc2\x30\x18\xc6\xcf\xeb\xa7\x78\x6f\xb0\x08\x17\x35\x5c\x76\x2a\xac\x46\xe2\xfe\xa5\x76\x46\xbc\x34\x65\x7b\x71\x4d\xc6\x20\x6d\x89\xf1\xdb\x1b\x3c\x8c\x55\x21\x7a\xe3\xda\xfc\xf2\x3e\x79\x7e\x7d\xa6\x53\xb8\xd9\xea\x77\xa3\x1c\x42\xb9\x27\x0b\xce\xa8\x60\x20\xe8\x3c\x61\x50\xb5\x1a\x3b\x67\x61\x4c\x02\x5d\xc3\x0b\xe5\x8b\x47\xca\xc7\x77\xb3\x10\x0a\xbe\x4c\x29\x5f\xc1\x13\x5b\x4d\x48\x60\xb1\x32\xe8\x64\xa3\x6c\x03\x82\xbd\x0a\xc8\x72\x01\x59\x99\x24\x10\xb3\x07\x5a\x26\x02\x46\xa3\x13\x66\xab\x06\xb7\x78\x3a\x77\x1b\x5e\xe0\xab\x5d\xb7\xd1\x35\x76\x4e\xab\x16\xe6\x79\x9e\x30\x9a\xfd\x46\x37\xaa\xb5\x78\xa4\x0d\x2a\x87\xb5\x54\x0e\xc4\x32\x65\xcf\x82\xa6\x85\x78\xeb\xf9\x01\xb1\xfe\xec\xc3\x67\xf7\x17\xc3\x7b\x58\xea\xbd\xd7\xfd\x0c\x4f\xc2\x88\xf8\xea\x0c\xd6\xda\x60\xe5\xe4\xc1\xe8\xbf\x05\x1e\x8c\xf6\xc5\x4d\x48\xa0\xad\x5c\x2b\x8b\xc7\x03\xff\x28\xff\xfd\x53\xf2\x47\xc8\x99\xee\x57\xb3\x33\xdc\x59\xbc\xfb\xe8\x48\xcc\xf3\xc2\xdf\x59\x34\x7c\xf3\x04\x46\xe4\x2b\x00\x00\xff\xff\x67\xff\x64\xc5\xa7\x02\x00\x00")

func sqlClients_20181208_1_initSqlBytes() ([]byte, error) {
	return bindataRead(
		_sqlClients_20181208_1_initSql,
		"sql/clients_20181208_1_init.sql",
	)
}

func sqlClients_20181208_1_initSql() (*asset, error) {
	bytes, err := sqlClients_20181208_1_initSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "sql/clients_20181208_1_init.sql", size: 679, mode: os.FileMode(436), modTime: time.Unix(1557644414, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _sqlClients_20190816_1_add_nameSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xd2\xd5\x55\xd0\xce\xcd\x4c\x2f\x4a\x2c\x49\x55\x08\x2d\xe0\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x48\xce\xc9\x4c\xcd\x2b\x29\x56\x70\x74\x71\x51\x70\xf6\xf7\x09\xf5\xf5\x53\xc8\x4b\xcc\x4d\x55\x08\x71\x8d\x08\x51\xf0\xf3\x0f\x51\xf0\x0b\xf5\xf1\xb1\xe6\xe2\x42\x36\xc2\x25\xbf\x3c\x0f\xab\x21\x2e\x41\xfe\x01\xc8\xa6\x58\x73\x01\x02\x00\x00\xff\xff\x31\xf9\x4a\x4e\x7a\x00\x00\x00")

func sqlClients_20190816_1_add_nameSqlBytes() ([]byte, error) {
	return bindataRead(
		_sqlClients_20190816_1_add_nameSql,
		"sql/clients_20190816_1_add_name.sql",
	)
}

func sqlClients_20190816_1_add_nameSql() (*asset, error) {
	bytes, err := sqlClients_20190816_1_add_nameSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "sql/clients_20190816_1_add_name.sql", size: 122, mode: os.FileMode(436), modTime: time.Unix(1566013510, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _sqlClients_20190920_1_unique_urisSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xd2\xd5\x55\xd0\xce\xcd\x4c\x2f\x4a\x2c\x49\x55\x08\x2d\xe0\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x28\x4a\x4d\xc9\x2c\x4a\x4d\x2e\x89\x2f\x2d\xca\x2c\x56\x70\x74\x71\x51\x70\xf6\xf7\x0b\x0e\x09\x72\xf4\xf4\x0b\x41\x95\x8c\x2f\xcd\xcb\x2c\x2c\x4d\x05\xb1\x15\x42\xfd\x3c\x03\x43\x5d\x35\x4a\x8b\x32\x35\xad\xb9\xb8\x90\x8d\x77\xc9\x2f\xcf\xc3\x63\x81\x4b\x90\x7f\x00\x31\x36\x58\x73\x01\x02\x00\x00\xff\xff\x6e\xb5\xc3\x82\xb4\x00\x00\x00")

func sqlClients_20190920_1_unique_urisSqlBytes() ([]byte, error) {
	return bindataRead(
		_sqlClients_20190920_1_unique_urisSql,
		"sql/clients_20190920_1_unique_uris.sql",
	)
}

func sqlClients_20190920_1_unique_urisSql() (*asset, error) {
	bytes, err := sqlClients_20190920_1_unique_urisSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "sql/clients_20190920_1_unique_uris.sql", size: 180, mode: os.FileMode(436), modTime: time.Unix(1568966914, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"sql/clients_20181208_1_init.sql":        sqlClients_20181208_1_initSql,
	"sql/clients_20190816_1_add_name.sql":    sqlClients_20190816_1_add_nameSql,
	"sql/clients_20190920_1_unique_uris.sql": sqlClients_20190920_1_unique_urisSql,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"sql": &bintree{nil, map[string]*bintree{
		"clients_20181208_1_init.sql":        &bintree{sqlClients_20181208_1_initSql, map[string]*bintree{}},
		"clients_20190816_1_add_name.sql":    &bintree{sqlClients_20190816_1_add_nameSql, map[string]*bintree{}},
		"clients_20190920_1_unique_uris.sql": &bintree{sqlClients_20190920_1_unique_urisSql, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
