package util

import (
	"compress/gzip"
	"archive/tar"
	"strings"
	"os"
	"log"
	"io"
)

func tarGzWrite( _path string, tw *tar.Writer, fi os.FileInfo) (err error) {
	fr, er := os.Open( _path )

	if er != nil {
		err = er
		log.Println("TarGz Error", err)
		return
	}

	defer fr.Close()

	h := new( tar.Header )
	h.Name = strings.Replace(_path, "dest/", "", 1)

	h.Size = fi.Size()
	h.Mode = int64( fi.Mode() )
	h.ModTime = fi.ModTime()

	err = tw.WriteHeader( h )
	if err != nil {
		log.Println("TarGz Error", err)
		return
	}

	_, err = io.Copy( tw, fr )
	if err != nil {
		log.Println("TarGz Error", err)
		return
	}
	return
}

func iterateDirectory( dirPath string, tw *tar.Writer ) (err error) {
	dir, er := os.Open( dirPath )
	if er != nil {
		err = er
		log.Println("TarGz Error", err)
		return
	}
	defer dir.Close()
	fis, er := dir.Readdir( 0 )

	if er != nil {
		err = er
		log.Println("TarGz Error", err)
		return
	}

	for _, fi := range fis {
		curPath := dirPath + "/" + fi.Name()
		if fi.IsDir() {
			//TarGzWrite( curPath, tw, fi )
			iterateDirectory( curPath, tw )
		} else {
			//fmt.Printf( "adding... %s\n", curPath )
			tarGzWrite( curPath, tw, fi )
		}
	}
	return
}

func TarGz( outFilePath string, inPath string ) (err error) {
	// file write
	fw, er := os.Create( outFilePath )

	if er != nil {
		err = er
		log.Println("TarGz Error", err)
		return
	}

	defer fw.Close()

	// gzip write
	gw := gzip.NewWriter( fw )
	defer gw.Close()

	// tar write
	tw := tar.NewWriter( gw )
	defer tw.Close()

	iterateDirectory( inPath, tw )

	return
}
