package cmd

import (
	"fmt"
	"github.com/gookit/gcli/v3"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ParseOpts = struct {
	path string
}{}

func ParseMarkDownImageCommand() *gcli.Command {
	c := gcli.NewCommand(
		"parse",
		"Parsing pictures in Markdown",
	)
	c.Aliases = []string{"p", "P"}
	c.StrOpt(&ParseOpts.path, "path", "p", "", "the option message")
	c.SetFunc(func(c *gcli.Command, args []string) error {
		//fmt.Printf("%+v\n", opts)
		//fmt.Println(args)
		rawImageRes, absImageRes := ParseMarkDownImage()
		fmt.Printf("共找到 %d 个图片  %+v\n", len(rawImageRes), rawImageRes)
		fmt.Printf("共找到 %d 个图片  %+v\n", len(absImageRes), absImageRes)
		return nil
	})
	return c
}

func ParseMarkDownImage() ([]string, []string) {
	mdPath := strings.Join(strings.Split(ParseOpts.path, "/")[:len(strings.Split(ParseOpts.path, "/"))-1], "/")
	var rawImageRes []string
	var absImageRes []string
	//fmt.Printf("%+v\n", opts.path)
	//fmt.Println(c.Args())

	mdImageRegex, _ := regexp.Compile("!\\[.*?\\]\\((.*?)\\)")
	res, _ := os.ReadFile(ParseOpts.path)
	source := string(res)
	result := mdImageRegex.FindAllStringSubmatch(source, -1)
	//fmt.Printf("%+v\n", result)
	for _, i := range result {
		if !strings.HasPrefix(i[1], "http") && i[1] != "" { // 避免网络图片和空路径图片
			if strings.Count(i[1], "%") > 5 {
				uni, _ := url.QueryUnescape(i[1])
				rawImageRes = append(rawImageRes, i[1])
				fpt, err := filepath.Abs(mdPath + "/" + uni)
				if err != nil {

				}
				absImageRes = append(absImageRes, fpt)
			} else {
				rawImageRes = append(rawImageRes, i[1])
				fpt, err := filepath.Abs(mdPath + "/" + i[1])
				if err != nil {

				}
				absImageRes = append(absImageRes, fpt)
			}
		}
	}
	return rawImageRes, absImageRes
}
