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
	"log"
	"net/http"
	"os"
	"setuServer/config"
	"setuServer/transmit"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"google.golang.org/grpc"
)

func dumpPictureToLocalServer(result *Result, dumpClient transmit.PicCourierClient, dumpUrl string) {
	for index, setu := range result.Setus {
		name, err := getPictureName(setu.Urls.Original)
		if err != nil {
			log.Println(err)
			continue
		}
		picFile, err := os.OpenFile(result.picPaths[index], os.O_RDONLY, 0666)
		if err != nil {
			continue
		}
		pic, err := ioutil.ReadAll(picFile)
		if err != nil {
			log.Println(err)
			_ = picFile.Close()
			continue
		}
		reply, err := dumpClient.SendPic(context.Background(), &transmit.PicRequest{Pic: pic, PicName: name})
		if err != nil {
			log.Println("Call dump pictures rpc failed.", err)
		} else {
			log.Println("Dump pictures success!", reply.Message)
			result.Setus[index].DumpUrl = dumpUrl + name
		}
		_ = picFile.Close()
	}
}

func transmitSetu(courier transmit.SetuCourierClient, messages []BotMsgReq) {
	for _, msg := range messages {
		article := msg.News.Articles[0]
		setuReq := transmit.SeTuRequest{Title: article.Title,
			Desc:        article.Description,
			OriginalUrl: article.Picurl,
			Url:         msg.Text.Content,
			PicBase64:   msg.Image.Base64,
			PicMd5:      msg.Image.Md5}
		reply, err := courier.SendSuTu(context.Background(), &setuReq)
		if err != nil {
			log.Println("Call dump pictures rpc failed.", err)
		} else {
			log.Println("Dump pictures success!", reply.ErrMessage)
		}
	}
}

func postSetuNews(result Result, transmitMsg *[]BotMsgReq) (err error) {
	for i := 0; i < len(result.Setus); i++ {
		setu := result.Setus[i]
		desc := fmt.Sprintf("Author: %s, Tags: ", setu.Author)
		for _, tag := range setu.Tags {
			desc += tag + " | "
		}
		article := Article{Title: setu.Title,
			Description: desc,
			Url:         setu.Urls.Original,
			Picurl:      setu.Urls.Original}
		var articles []Article
		articles = append(articles, article)
		news := &News{Articles: articles}
		postNews := BotMsgReq{MsgType: BotMsgNews, News: news}
		err = postSetuToWeChat(postNews)
		if err != nil {
			log.Println("Post setu news failed.")
			return
		}
		(*transmitMsg)[i].News = news
	}
	return
}

func postSetuText(result Result, atAll bool, transmitMsg *[]BotMsgReq) {
	for i := 0; i < len(result.Setus); i++ {
		setu := result.Setus[i]
		var MentionedList []string
		if atAll {
			MentionedList = append(MentionedList, "@all")
		}
		content := setu.Urls.Original
		if setu.DumpUrl != "" {
			content = setu.DumpUrl
		}
		txt := &Text{
			Content:       content,
			MentionedList: MentionedList,
		}
		postText := BotMsgReq{
			MsgType: BotMsgText,
			Text:    txt,
		}
		err := postSetuToWeChat(postText)
		if err != nil {
			log.Println("Post setu text failed.")
		}
		(*transmitMsg)[i].Text = txt
	}
}

func postSetuPic(result Result, transmitMsg *[]BotMsgReq) {
	for i := 0; i < len(result.Setus); i++ {
		picPath := result.getPicPath(uint(i))
		compress := false
		for round := 0; round < 5; round++ {
			fileInfo, err := os.Stat(picPath)
			if err != nil {
				log.Println(err)
				break
			}
			if fileInfo.Size() > 2*1024*1024 {
				picPath, err = picCompress(picPath)
				if err != nil {
					log.Println(err)
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
				log.Println(err)
				continue
			}
			picData, err := ioutil.ReadAll(picFile)
			if err != nil {
				log.Println(err)
				_ = picFile.Close()
				continue
			}
			picBase64 := base64.StdEncoding.EncodeToString(picData)
			md5Hash := md5.New()
			md5Hash.Write(picData)
			md5Str := hex.EncodeToString(md5Hash.Sum(nil))

			img := &Image{Base64: picBase64, Md5: md5Str}
			postPic := BotMsgReq{MsgType: BotMsgImage, Image: img}
			err = postSetuToWeChat(postPic)
			if err != nil {
				log.Println(err)
			}
			_ = picFile.Close()
			(*transmitMsg)[i].Image = img
		}
	}
}

// Run The main loop to send setu on time.
func Run() {
	first := true
	cfg := config.GetGlobalConfig()
	var dumpClient transmit.PicCourierClient
	var setuClient transmit.SetuCourierClient
	if cfg.PicDump {
		picConn, err := grpc.Dial(cfg.DumpServer, grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallSendMsgSize(31457280),
				grpc.MaxCallRecvMsgSize(31457280)))
		if err != nil {
			log.Println("Connect dump server failed.", err)
			os.Exit(-1)
		}
		setuConn, err := grpc.Dial(cfg.TransmitServer, grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallSendMsgSize(31457280),
				grpc.MaxCallRecvMsgSize(31457280)))
		if err != nil {
			log.Println("Connect transmit server failed.", err)
			os.Exit(-1)
		}
		dumpClient = transmit.NewPicCourierClient(picConn)
		setuClient = transmit.NewSetuCourierClient(setuConn)
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
			log.Println("Get setu failed.")
			continue
		}
		if cfg.PicDump {
			dumpPictureToLocalServer(&result, dumpClient, cfg.DumpUrl)
		}
		messages := make([]BotMsgReq, 1)
		postSetuText(result, cfg.AtAll, &messages)
		// Post setu by different way
		if cfg.NewsMsg {
			if err := postSetuNews(result, &messages); err != nil {
				log.Println(err)
				continue
			}
		}
		// Post setu pic
		if cfg.PicMsg {
			postSetuPic(result, &messages)
		}
		if cfg.SetuTransmit {
			transmitSetu(setuClient, messages)
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

	query := Query{R18: r18, Num: 1, Tag: cfg.Tags, Size: cfg.PicSize}
	jsonStr, err := json.Marshal(query)
	if err != nil {
		log.Println("Marshal json failed.", err)
		return
	}
	req, err := http.NewRequest("POST", cfg.SetuApiUrl, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println("Http request failed.", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Http Do failed.", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != 200 {
		log.Println("Http Get status is", resp.StatusCode, "not 200")
		return
	}
	bodyStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Read Http Get body failed.", err)
		return
	}
	err = json.Unmarshal(bodyStr, &result)
	if err != nil {
		log.Println("Json unmarshal failed.", err)
		return
	}
	if result.Error != "" {
		log.Println("Result error! Error message:", result.Error)
		return
	}
	// Don't need to get picture message, return
	if !cfg.PicMsg {
		return
	}
	// Download setu from pic url.
	for i := 0; i < len(result.Setus); i++ {
		setu := result.Setus[i]
		var dlReq *http.Request
		var dlResp *http.Response
		dlReq, err = http.NewRequest("GET", setu.Urls.Original, bytes.NewBuffer([]byte("")))
		if err != nil {
			log.Println(err)
			return
		}
		dlReq.Header.Set("Referer", "https://i.pixiv.cat/")
		dlResp, err = http.DefaultClient.Do(dlReq)
		if err != nil {
			log.Println("Download picture failed.", err)
			return
		}
		var name string
		name, err = getPictureName(setu.Urls.Original)
		if err != nil {
			log.Println(err)
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
	if cfg.WeChatUrl == "" {
		return nil
	}
	postStr, err := json.Marshal(post)
	if err != nil {
		log.Println("Json marshal post failed.", err)
		return
	}
	respPost, err := http.Post(cfg.WeChatUrl, "application/json", bytes.NewBuffer(postStr))
	if err != nil {
		log.Println("Post to wechat failed", err)
		return
	}
	msg, err := ioutil.ReadAll(respPost.Body)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(msg))
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
