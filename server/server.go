package server

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"setuServer/config"
	"setuServer/picdump"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"google.golang.org/grpc"
)

func dumpPictureToLocalServer(result *Result, dumpClient picdump.CourierClient, dumpUrl string) {
	for index, setu := range result.Setus {
		name, err := getPictureName(setu.Url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		picFile, err := os.OpenFile(result.picPaths[index], os.O_RDONLY, 0666)
		if err != nil {
			continue
		}
		pic, err := ioutil.ReadAll(picFile)
		if err != nil {
			fmt.Println(err)
			_ = picFile.Close()
			continue
		}
		reply, err := dumpClient.SendPic(context.Background(), &picdump.PicRequest{Pic: pic, PicName: name})
		if err != nil {
			fmt.Println("Call dump pictures rpc failed.", err)
		} else {
			fmt.Println("Dump pictures success!", reply.Message)
			result.Setus[index].DumpUrl = dumpUrl + name
		}
		_ = picFile.Close()
	}
}

func postSetuNews(result Result) (err error) {
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
	postNews := BotMsgReq{MsgType: BotMsgNews, News: &News{Articles: articles}}
	err = postSetuToWeChat(postNews)
	if err != nil {
		fmt.Println("Post setu news failed.")
		return
	}
	return
}

func postSetuText(result Result, atAll bool) {
	for i := 0; i < result.Count; i++ {
		setu := result.Setus[i]
		var MentionedList []string
		if atAll {
			MentionedList = append(MentionedList, "@all")
		}
		content := setu.Url
		if setu.DumpUrl != "" {
			content = setu.DumpUrl
		}
		postText := BotMsgReq{
			MsgType: BotMsgText,
			Text: &Text{
				Content:       content,
				MentionedList: MentionedList,
			},
		}
		err := postSetuToWeChat(postText)
		if err != nil {
			fmt.Println("Post setu text failed.")
		}
	}
}

func postSetuPic(result Result) {
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
			postPic := BotMsgReq{MsgType: BotMsgImage, Image: &Image{Base64: picBase64, Md5: md5Str}}
			err = postSetuToWeChat(postPic)
			if err != nil {
				fmt.Println(err)
			}
			_ = picFile.Close()
		}
	}
}

// Run The main loop to send setu on time.
func Run() {
	first := true
	cfg := config.GetGlobalConfig()
	var dumpClient picdump.CourierClient
	if cfg.PicDump {
		conn, err := grpc.Dial(cfg.DumpServer, grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallSendMsgSize(31457280),
				grpc.MaxCallRecvMsgSize(31457280)))
		if err != nil {
			fmt.Println("Connect dump server failed.", err)
			os.Exit(-1)
		}
		dumpClient = picdump.NewCourierClient(conn)
	}
	intervals := cfg.Intervals
	if intervals < 10 {
		intervals = 10
	}
	for true {
		if first {
			first = false
		} else {
			time.Sleep(time.Duration(intervals) * time.Second)
		}
		// Get setu info & download setu picture
		result, err := getSetuFromApi()
		if err != nil {
			fmt.Println("Get setu failed.")
			continue
		}
		if cfg.PicDump {
			dumpPictureToLocalServer(&result, dumpClient, cfg.DumpUrl)
		}
		postSetuText(result, cfg.AtAll)
		// Post setu by different way
		if cfg.NewsMsg {
			if err := postSetuNews(result); err != nil {
				fmt.Println(err)
				continue
			}
		}
		// Post setu pic
		if cfg.PicMsg {
			postSetuPic(result)
		}
	}
}

func getPictureName(url string) (string, error) {
	index := strings.LastIndex(url, "/img/")
	if index == -1 {
		return "", errors.New("can't find index in url")
	}
	name := url[index+1:]
	name = strings.ReplaceAll(name, "/", "-")
	return name, nil
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
	// Don't need to get picture message, return
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
		var name string
		name, err = getPictureName(setu.Url)
		if err != nil {
			fmt.Println(err)
			_ = dlResp.Body.Close()
			continue
		}
		path := cfg.PicDownloadDir + "/" + name
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

// postSetuToWeChat post setu to WeChat
func postSetuToWeChat(post BotMsgReq) (err error) {
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

// picCompress Modify size to compress pictures.
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
	newPicPath = picPath + "_" + time.Now().Format("15-04-05") + ".tmp.png"
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
