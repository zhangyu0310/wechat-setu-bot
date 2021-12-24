package main

import (
	"flag"
	"fmt"
	"os"
	"setuServer/config"
	"setuServer/server"
)

var (
	setuApiUrl = flag.String("setu-api-url", "https://api.lolicon.app/setu/v1", "Api Url of setu")
	wechatUrl  = flag.String("wechat-url", "", "Wechat Web Hook Url")
	intervals  = flag.Uint("intervals", 60, "Intervals of post setu.(second) [Minimum is 10s]")
	r18        = flag.Bool("r18", false, "Post R18 picture")
	atAll      = flag.Bool("at-all", false, "@all group member")
	picMsg     = flag.Bool("pic-msg", false, "Download picture & send picture msg")
	newsMsg    = flag.Bool("news-msg", true, "Send picture use news message")
	dlDir      = flag.String("dl-dir", "./", "Dir of download picture")
	picDump    = flag.Bool("pic-dump", false, "Dump setu pictures to local server")
	dumpServer = flag.String("dump-server", "", "Server info to dump pictures")
	dumpUrl    = flag.String("dump-url", "", "Url for user get pictures")
)

// cmdConfigSetToGlobal store command config to global config.
func cmdConfigSetToGlobal(cfg *config.Config) {
	cfg.SetuApiUrl = *setuApiUrl
	cfg.WeChatUrl = *wechatUrl
	cfg.Intervals = *intervals
	cfg.R18 = *r18
	cfg.AtAll = *atAll
	cfg.PicMsg = *picMsg
	cfg.NewsMsg = *newsMsg
	cfg.PicDownloadDir = *dlDir
	cfg.PicDump = *picDump
	cfg.DumpServer = *dumpServer
	cfg.DumpUrl = *dumpUrl
}

func main() {
	help := flag.Bool("help", false, "show the usage")
	version := flag.Bool("version", false, "version of server")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *version {
		// TODO: Print version info.(Version/GitCommit/CompileTime...)
		fmt.Println("Version: v0.1.0")
		os.Exit(0)
	}
	if *wechatUrl == "" {
		fmt.Println("WeChat Url is required.")
		flag.Usage()
		os.Exit(-1)
	}
	config.InitializeConfig(cmdConfigSetToGlobal)
	server.Run()
}
