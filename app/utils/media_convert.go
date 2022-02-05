package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/golang/glog"

	"github.com/spf13/viper"
)

// ConvertM4aMp4 converts an m4a file to mp4 in place by using temp folder to create new file then copies to original
// ffmpeg -y -i input.m4a -c:a copy output.mp4
func ConvertM4aMp4(inputPath string) (error) {

	var err error

	_, filename := filepath.Split(inputPath)

	// name new mp4 temp file by adding mp4 suffix
	tempPath := path.Join(viper.GetString("media.uploadTemp.path"), filename + ".mp4")

	// convert and save in temp
	s := viper.GetStringSlice("commands.exec.m4a2mp4.args")
	for i, arg := range s {
		if arg == "m4afile" {
			s[i] = inputPath
		} else if arg == "mp4file" {
			s[i] = tempPath
		}
	}
	cmd := exec.Command(viper.GetString("commands.exec.m4a2mp4.command"), s...)
	glog.V(2).Infof("running ffmpeg %v", cmd.Args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("cannot run ffmpeg convert %s %v", inputPath, err)
		goto done
	}
	glog.V(2).Infof("ffmpeg output %s", string(out))

	// copy back to original and delete temp
	_, err = CopyFile(tempPath, inputPath)
	if err != nil {
		err = fmt.Errorf("cannot copy converted mp4 to original %s %v", inputPath, err)
	}

	done:
	err = os.Remove(tempPath) 
	if err != nil {
		err = fmt.Errorf("cannot delete temp converted mp4 %s %v", tempPath, err)
	}
	return err
}


// ConvertHeicJpg converts an heic file to jpeg in place by using temp folder to create new file then copies to original
// convert -format jpg input.heic output.jpg
func ConvertHeicJpg(inputPath string) (error) {

	var err error

	_, filename := filepath.Split(inputPath)

	// name new jpg temp file by adding jpg suffix
	tempPath := path.Join(viper.GetString("media.uploadTemp.path"), filename + ".jpg")

	// convert and save in temp
	s := viper.GetStringSlice("commands.exec.heic2jpg.args")
	for i, arg := range s {
		if arg == "heicfile" {
			s[i] = inputPath
		} else if arg == "jpgfile" {
			s[i] = tempPath
		}
	}
	cmd := exec.Command(viper.GetString("commands.exec.heic2jpg.command"), s...)
	glog.V(2).Infof("running convert %v", cmd.Args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("cannot run convert %s %v", inputPath, err)
		goto done
	}
	glog.V(2).Infof("convert output %s", string(out))

	// copy back to original and delete temp
	_, err = CopyFile(tempPath, inputPath)
	if err != nil {
		err = fmt.Errorf("cannot copy converted jpg to original %s %v", inputPath, err)
	}

	done:
	err = os.Remove(tempPath) 
	if err != nil {
		err = fmt.Errorf("cannot delete temp converted jpg %s %v", tempPath, err)
	}
	return err
}
