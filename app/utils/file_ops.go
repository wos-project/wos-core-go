package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"archive/tar"

	"github.com/golang/glog"
)

// CopyFile copies a file from src to dest full path
func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// DeleteFile deletes a file
func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		glog.Errorf("cannot delete file %s", path)
	} else {
		glog.Infof("deleted file %s", path)
	}
	return err
}

// CopyDir is a recursive file copy
func CopyDir(source, destination string) error {
	var err error = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			var data, err1 = ioutil.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})
	return err
}

// ExpandTarReader expands a tar stream to expandDirPath
func ExpandTarReader(r io.Reader, expandDirPath string) error {

	tarReader := tar.NewReader(r)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(path.Join(expandDirPath, header.Name), 0755); err != nil {
				return fmt.Errorf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			// create directory if it doesn't exist
			p := path.Join(expandDirPath, header.Name)
			d := filepath.Dir(p)
			_, err = os.Stat(d)
			if err != nil {
				err = os.MkdirAll(d, 0777)
				if err != nil {
					return fmt.Errorf("Cannot mkdirall %s %s", d, err.Error())
				}
			}
			outFile, err := os.Create(p)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
		default:
			return fmt.Errorf(
				"ExtractTarGz: uknown type: %v in %s",
				header.Typeflag,
				header.Name)
		}
	}
	return nil
}

// ExpandTarFile expands a tar file tarFilePath to expandDirPath
func ExpandTarFile(tarFilePath string, expandDirPath string) error {

	r, err := os.Open(tarFilePath)
	if err != nil {
		return fmt.Errorf("Cannot open tar file %s", tarFilePath)
	}
	defer r.Close()

	tarReader := tar.NewReader(r)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(path.Join(expandDirPath, header.Name), 0755); err != nil {
				return fmt.Errorf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			// create directory if it doesn't exist
			p := path.Join(expandDirPath, header.Name)
			d := filepath.Dir(p)
			_, err = os.Stat(d)
			if err != nil {
				err = os.MkdirAll(d, 0777)
				if err != nil {
					return fmt.Errorf("Cannot mkdirall %s %s", d, err.Error())
				}
			}
			outFile, err := os.Create(p)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
		default:
			return fmt.Errorf(
				"ExtractTarGz: uknown type: %v in %s",
				header.Typeflag,
				header.Name)
		}
	}
	return nil
}

// AddFileToTarWriter adds a file to a tar file
func AddFileToTarWriter(pathOnDisk string, pathInTar string, tarWriter *tar.Writer) error {
	file, err := os.Open(pathOnDisk)
	if err != nil {
		return fmt.Errorf("Could not open file '%s', got error '%v'", pathOnDisk, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("Could not get stat for file '%s', got error '%v'", pathOnDisk, err)
	}

	// TODO: remove root from filepath

	header := &tar.Header{
		Name:    pathInTar,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	err = tarWriter.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("Could not write header for file '%s', got error '%v'", pathOnDisk, err)
	}

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return fmt.Errorf("Could not copy the file '%s' data to the tarball, got error '%v'", pathOnDisk, err)
	}

	return nil
}
