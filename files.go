package util

import (
	"io/ioutil"
	"os"
	"log"
	"errors"
	"io"
	"path/filepath"
	"compress/gzip"
	"archive/tar"
	"fmt"
	"strings"
	"os/exec"
	"sort"
)

const (
	DefaultFileMode = 0600
	FileTimeFormat = "2006-01-02_15.04.05.000"

)

type ByDateFilesAsc []os.FileInfo
type ByDateFilesDesc []os.FileInfo

func (s ByDateFilesAsc) Len() int {
	return len(s)
}
func (s ByDateFilesAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByDateFilesAsc) Less(i, j int) bool {
	return s[i].ModTime().After(s[j].ModTime())
}

func (s ByDateFilesDesc) Len() int {
	return len(s)
}
func (s ByDateFilesDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByDateFilesDesc) Less(i, j int) bool {
	return s[i].ModTime().Before(s[j].ModTime())
}

func SortFilesByDate(files []os.FileInfo, asc bool) {
	if asc {
		//sort.Sort(ByDateFilesAsc(files))
		sort.SliceStable(files, func(i,j int) bool {
			return files[i].ModTime().After(files[j].ModTime())
		})
	} else {
		//sort.Sort(ByDateFilesDesc(files))
		sort.SliceStable(files, func(i,j int) bool {
			return files[i].ModTime().Before(files[j].ModTime())
		})
	}
}

//delete dir
func RemoveDirRecursively(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

//copy file with rights
func CopyFile(source string, dest string) (err error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err == nil {
		si, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, si.Mode())
		}

	}

	return
}

func CopyDir(source string, dest string, excludes ...func(os.FileInfo) bool) (err error) {

	// get properties of source dir
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.New("Source is not a directory")
	}

	// ensure dest dir does not already exist

	_, err = os.Open(dest)
	if !os.IsNotExist(err) {
		return errors.New("Destination already exists")
	}

	// create dest dir

	err = os.MkdirAll(dest, fi.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(source)

	filterExcluded := func(f os.FileInfo) (result bool){
		for _,pred := range excludes {
			if pred(f) {
				result = true
				break
			}
		}
		return
	}

	for _, entry := range entries {

		if filterExcluded(entry) {
			continue
		}

		sfp := source + "/" + entry.Name()
		dfp := dest + "/" + entry.Name()

		if entry.IsDir() {
			err = CopyDir(sfp, dfp, excludes...)
			if err != nil {
				log.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(sfp, dfp)
			if err != nil {
				log.Println(err)
			}
		}

	}
	return
}

func MakeDirIfNotExists(dir string, fileMode os.FileMode) (err error) {
	if _, er := os.Stat(dir); er != nil {
		if os.IsNotExist(er) {
			err = os.Mkdir(dir, fileMode)
		} else {
			log.Println(err)
		}
	}
	return
}

func FileExists(filename string) (exists bool) {
	if _, err := os.Stat(filename); err == nil {
		exists = true
	}
	return
}

func Rsync(targetDir, params, dest string, delete bool) (err error)  {
	command := "rsync %v %v %v"
	if delete {
		command = command + " --delete"
	}
	cmd := exec.Command("bash", "-c", fmt.Sprintf(command, params, targetDir, dest))
	err = cmd.Run()
	return
}

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
