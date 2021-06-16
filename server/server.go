package server

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"setuServer/config"
	"strconv"
	"strings"
	"time"
)

func Run() {
	first := true
	cfg := config.GetGlobalConfig()
	for true {
		if !first {
			intervals := cfg.Intervals
			if intervals < 10 {
				intervals = 10
			}
			time.Sleep(time.Duration(intervals) * time.Second)
		}
		first = false
		// Get setu info & download setu picture
		result, err := getSetuFromApi()
		if err != nil {
			fmt.Println("Get setu failed.")
			continue
		}
		// Post setu news
		var articles []Article
		for i := 0; i < result.Count; i++ {
			setu := result.Setus[i]
			desc := fmt.Sprintf("Author: %s, Tags: ", setu.Author)
			for _, tag := range setu.Tags {
				desc += tag + " | "
			}
			article := Article{Title: setu.Title,
				Description: desc,
				Url:         setu.Url,
				Picurl:      setu.Url}
			articles = append(articles, article)
		}
		postNews := PostWeChatNews{MsgType: "news", News: News{Articles: articles}}
		err = postSetuToWeChat(postNews)
		if err != nil {
			fmt.Println("Post setu news failed.")
			continue
		}
		// Post setu text
		for i := 0; i < result.Count; i++ {
			setu := result.Setus[i]
			var MentionedList []string
			if cfg.AtAll {
				MentionedList = append(MentionedList, "@all")
			}
			postText := PostWeChatText{MsgType: "text",
				Text: Text{Content: setu.Url,
					MentionedList: MentionedList}}
			err = postSetuToWeChat(postText)
			if err != nil {
				fmt.Println("Post setu text failed.")
			}
		}
		// Post setu pic
		if !cfg.PicMsg {
			continue
		}
		for i := 0; i < result.Count; i++ {
			picPath := result.getPicPath(uint(i))
			compress := false
			for round := 0; round < 5; round++ {
				fileInfo, err := os.Stat(picPath)
				if err != nil {
					fmt.Println(err)
					break
				}
				if fileInfo.Size() > 2*1024*1024 {
					picPath, err = picCompress(picPath)
					if err != nil {
						fmt.Println(err)
						break
					}
				} else {
					compress = true
					break
				}
			}
			if compress {
				picFile, err := os.OpenFile(picPath, os.O_RDONLY, 0666)
				if err != nil {
					fmt.Println(err)
					continue
				}
				picData, err := ioutil.ReadAll(picFile)
				if err != nil {
					fmt.Println(err)
					_ = picFile.Close()
					continue
				}
				picBase64 := base64.StdEncoding.EncodeToString(picData)
				md5Hash := md5.New()
				md5Hash.Write(picData)
				md5Str := hex.EncodeToString(md5Hash.Sum(nil))
				postPic := PostWeChatPic{MsgType: "image", Image: Image{Base64: picBase64, Md5: md5Str}}
				err = postSetuToWeChat(postPic)
				if err != nil {
					fmt.Println(err)
				}
				_ = picFile.Close()
			}
		}
	}
}

// getSetuFromApi get setu info & download setu
func getSetuFromApi() (result Result, err error) {
	cfg := config.GetGlobalConfig()
	// TODO: Set Key Word & Num (WeChat not support interact)
	// Get setu from setu api
	r18 := 0
	if cfg.R18 {
		r18 = 2
	}
	query := Query{R18: r18, Num: 1, SmallSize: false}
	jsonStr, err := json.Marshal(query)
	if err != nil {
		fmt.Println("Marshal json failed.", err)
		return
	}
	req, err := http.NewRequest("GET", cfg.SetuApiUrl, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println("Http request failed.", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Http Do failed.", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Println("Http Get status is", resp.StatusCode, "not 200")
		return
	}
	bodyStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read Http Get body failed.", err)
		return
	}
	err = json.Unmarshal(bodyStr, &result)
	if err != nil {
		fmt.Println("Json unmarshal failed.", err)
		return
	}
	if result.Code != 0 {
		fmt.Println("Result error, Code is", result.Code, "Mes:", result.Msg)
		return
	}
	// If don't need picture message, return
	if !cfg.PicMsg {
		return
	}
	// Download setu from pic url.
	for i := 0; i < result.Count; i++ {
		setu := result.Setus[i]
		var dlReq *http.Request
		var dlResp *http.Response
		dlReq, err = http.NewRequest("GET", setu.Url, bytes.NewBuffer([]byte("")))
		if err != nil {
			fmt.Println(err)
			return
		}
		dlReq.Header.Set("Referer", "https://i.pixiv.cat/")
		dlResp, err = http.DefaultClient.Do(dlReq)
		if err != nil {
			fmt.Println("Download picture failed.", err)
			return
		}
		index := strings.LastIndex(setu.Url, ".")
		if index == -1 {
			_ = dlResp.Body.Close()
			continue
		}
		format := setu.Url[index:]
		now := time.Now().Format("2006-01-02-15-04-05")
		path := cfg.PicDownloadDir + "/img" + strconv.Itoa(i) + "-" + now + format
		var imgFile *os.File
		imgFile, err = os.Create(path)
		if err != nil {
			return
		}
		_, err = io.Copy(imgFile, dlResp.Body)
		if err != nil {
			return
		}
		result.setPicPath(path)
		_ = imgFile.Close()
		_ = dlResp.Body.Close()
	}
	return
}

// postSetuToWeChat post setu to wechat
func postSetuToWeChat(post interface{}) (err error) {
	cfg := config.GetGlobalConfig()
	postStr, err := json.Marshal(post)
	if err != nil {
		fmt.Println("Json marshal post failed.", err)
		return
	}
	respPost, err := http.Post(cfg.WeChatUrl, "application/json", bytes.NewBuffer(postStr))
	if err != nil {
		fmt.Println("Post to wechat failed", err)
		return
	}
	msg, err := ioutil.ReadAll(respPost.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(msg))
	_ = respPost.Body.Close()
	return
}

func picCompress(picPath string) (newPicPath string, err error) {
	picFile, err := os.OpenFile(picPath, os.O_RDONLY, 0666)
	if err != nil {
		return
	}
	defer func(picFile *os.File) {
		_ = picFile.Close()
	}(picFile)
	pic, _, err := image.Decode(picFile)
	if err != nil {
		return
	}
	newPicPath = picPath + "_" + time.Now().Format("15-04-05")
	newPic := resize.Resize(uint(pic.Bounds().Dx()/2), 0, pic, resize.Lanczos3)
	newPicFile, err := os.OpenFile(newPicPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer func(newPicFile *os.File) {
		_ = newPicFile.Close()
	}(newPicFile)
	err = png.Encode(newPicFile, newPic)
	if err != nil {
		return
	}
	return
}
