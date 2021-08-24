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
	wechatUrl  = flag.String("wechat-url", "", "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=8f32332c-93a9-4323-896b-8eb3b30b696a")
	intervals  = flag.Uint("intervals", 60, "Intervals of post setu.(second) [Minimum is 10s]")
	r18        = flag.Bool("r18", false, "Post R18 picture")
	atAll      = flag.Bool("at-all", false, "@all group member")
	picMsg     = flag.Bool("pic-msg", false, "Download picture & send picture msg")
	dlDir      = flag.String("dl-dir", "./", "Dir of download picture")
)

// cmdConfigSetToGlobal store command config to global config.
func cmdConfigSetToGlobal(cfg *config.Config) {
	cfg.SetuApiUrl = *setuApiUrl
	cfg.WeChatUrl = *wechatUrl
	cfg.Intervals = *intervals
	cfg.R18 = *r18
	cfg.AtAll = *atAll
	cfg.PicMsg = *picMsg
	cfg.PicDownloadDir = *dlDir
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
