package util

import (
	"io/ioutil"
	"os"
	"log"
	"errors"
	"io"
	"path/filepath"
	"sort"
	"os/exec"
	"fmt"
)

const (
	DefaultFileMode = 0600
	FileTimeFormat = "2006-01-02_15.04.05.000"
)

func SortFilesByDate(files []os.FileInfo, asc bool) {
	if asc {
		sort.SliceStable(files, func(i,j int) bool {
			return files[i].ModTime().After(files[j].ModTime())
		})
	} else {
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

func RsyncSSH(src, dest string, delete bool, excludes ...string) (err error) {

	command := "rsync -avz -e 'ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null' "

	if delete {
		command += "--delete "
	}

	if len(excludes) > 0 {
		for _, param := range excludes {
			command += fmt.Sprintf("--exclude %v ", param)
		}
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("%v %v %v", command, src, dest))
	err = cmd.Run()
	return
}
