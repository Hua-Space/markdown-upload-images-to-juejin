package main

import (
	"github.com/gookit/gcli/v3"
	"github.com/gookit/gcli/v3/builtin"
	"markdown-upload-images-to-juejin/cmd"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	app := gcli.NewApp(func(app *gcli.App) {
		app.Version = "1.0.0"
		app.Desc = "This tool is designed to help you host markdown images"
	})

	app.Add(
		cmd.ParseMarkDownImageCommand(),
		cmd.HandleMarkDownImageCommand(),
		builtin.GenAutoComplete(),
	)

	app.Run(nil)
}
