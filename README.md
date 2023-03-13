## Usage

> When should I use this tool?
> > This tool is used to help publish articles on Juejin.

To use this tool, simply run the following command:

```shell
This tool is designed to help you host markdown images (Version: 1.0.0)
Usage:
  main [global options...] COMMAND [--options ...] [arguments ...]

Global Options:
  -h, --help                      Display the help information
  --nc, --no-color                Disable color when outputting message
  --ni, --no-interactive          Disable interactive confirmation operation
  --np, --no-progress             Disable display progress message
  --verb, --verbose               Set logs reporting level(quiet 0 - 5 crazy) (default 1=error)
  -V, --version                   Display app version information

Available Commands:
  genac        Generate auto complete scripts for current application (alias: gen-ac)
  parse        Parsing pictures in Markdown (alias: p,P)
  upload       Upload and Parsing pictures in Markdown (alias: u,U)
  help         Display help information

```

Examples:

```shell
./markdown-upload-images-to-juejin u --p=./test.md -s="your sessionid"
```






## FAQ
> How do I get my session ID?
> > To get your session ID, find the `sessionid` from the cookie in the login state.

> Q: How do I upload images to Juejin using this tool?
> >A: To upload images to Juejin using this tool, download the binary file for your platform from the Releases section on GitHub. The tool will automatically process the image in Markdown, upload it to Juejin, and generate the URL.

> Q: What image formats are supported by this tool?
> >A: This tool supports PNG, JPEG, and GIF image formats.

>Q: How do I report a bug or suggest a feature?
> >A: To report a bug or suggest a feature, please open an issue on the tool's GitHub repository.



## Thank

[Go 实现 AWS4 请求认证 | Go主题月](https://juejin.cn/post/6950300506946273294)