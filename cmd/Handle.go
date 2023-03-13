package cmd

import (
	"errors"
	"fmt"
	"github.com/gookit/gcli/v3"
	"markdown-upload-images-to-juejin/lib"
	"os"
	"regexp"
	"strings"
)

var HandleOpts = struct {
	path    string
	service string
}{}

func HandleMarkDownImageCommand() *gcli.Command {
	c := gcli.NewCommand(
		"upload",
		"upload and Parsing pictures in Markdown",
	)
	c.Aliases = []string{"u", "U"}
	c.StrOpt(&HandleOpts.path, "path", "p", "", "要处理的 Markdown 路径，可为目录")
	//c.StrOpt(&HandleOpts.service, "service", "s", "", "上传服务")
	c.StrVar(&UploadOpts.Session, &gcli.FlagMeta{
		Name:     "sessionid",
		Desc:     "掘金Cookie `Sessionid`",
		Shorts:   []string{"s", "S"},
		Required: true,
		Validator: func(val string) error {
			pattern := "^[a-zA-Z0-9]{32}$"
			matched, err := regexp.MatchString(pattern, val)
			if err != nil {
				fmt.Println("正则匹配出错：", err)
			}
			if !matched {
				return errors.New("参数 `sessionid` 验证失败，由数字字母组成的32位字符")
			}

			return nil
		},
	})
	c.SetFunc(func(c *gcli.Command, args []string) error {
		//fmt.Printf("%+v\n", opts)
		//fmt.Println(args)
		HandleMarkDownImage()
		return nil
	})
	return c
}

func HandleMarkDownImage() bool {
	if lib.IsDir(HandleOpts.path) {
		fmt.Printf("%s 是一个目录 🐱", HandleOpts.path)
		files, _ := lib.GetFileName(HandleOpts.path)
		//fmt.Printf("%+v\n", files)
		fmt.Printf("共找到 %d 个 md 文件 \n", len(files))
		for _, file := range files {
			HandleMarkDownSingleFile(file)
		}
	} else {
		HandleMarkDownSingleFile()
	}

	return true
}

func HandleMarkDownSingleFile(path ...string) {
	UploadOpts.fail = 0
	ParseOpts.path = HandleOpts.path
	if len(path) > 0 {
		ParseOpts.path = path[0]
	}
	fmt.Println(fmt.Sprintf("-------------处理文件 : %s", ParseOpts.path))
	rawImageRes, absImageRes := ParseMarkDownImage()
	if len(rawImageRes) == 0 {
		return
	}
	var newImages []string
	for num, ImagePath := range absImageRes {
		fmt.Println(fmt.Sprintf("📤 正在上传 : %d/%d", num+1, len(absImageRes)))

		image, err := uploadTempImage(ImagePath)
		if err != nil {
			newImages = append(newImages, rawImageRes[num])
		} else {
			newImages = append(newImages, image)
		}

	}
	replaceMdImage(rawImageRes, newImages)
}

/*
修改markdown中上传图片的地址
*/
func replaceMdImage(oldImages []string, newImages []string) {
	if len(oldImages) == len(newImages) {
		res, _ := os.ReadFile(ParseOpts.path)
		source := string(res)
		_ = os.WriteFile(ParseOpts.path+"x", []byte(source), 0666)
		for i := 0; i < len(oldImages); i++ {
			source = strings.Replace(source, oldImages[i], newImages[i], -1)
		}
		if UploadOpts.fail > 0 {
			fmt.Printf("❌上传失败 %d 个\n", UploadOpts.fail)
		}
		_ = os.WriteFile(ParseOpts.path, []byte(source), 0666)
		fmt.Printf("👌 备份 : %sx\n", ParseOpts.path)

	} else {
		panic("replaceMdImage 图片长度不对等！！！")
	}
}
