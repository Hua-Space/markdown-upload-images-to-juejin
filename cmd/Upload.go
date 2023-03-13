package cmd

import (
	"fmt"
	"os"
)

var UploadOpts = struct {
	path    string
	fail    int
	Session string
}{}

func uploadTempImage(ImagePath string) (string, error) {
	_, err := os.Stat(ImagePath)
	if err != nil {
		fmt.Printf("无法读取 : %s \n", ImagePath)
		UploadOpts.fail = UploadOpts.fail + 1
		return "", err
	}
	res, err := JuejinUploadImage(ImagePath)
	if err != nil {
		fmt.Println(err.Error())
		UploadOpts.fail = UploadOpts.fail + 1
		return "", err
	}

	return res, err
}
