package server

type Query struct {
	R18 int `json:"r18"`
	// KeyWord string `json:"keyword"`
	Num int `json:"num"`
	// Proxy string `json:"proxy"`
	SmallSize bool `json:"size1200"`
}

type Result struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Count    int    `json:"count"`
	Setus    []Setu `json:"data"`
	picPaths []string
}

func (result *Result) setPicPath(path string) {
	result.picPaths = append(result.picPaths, path)
}

func (result *Result) getPicPath(index uint) string {
	return result.picPaths[index]
}

type Setu struct {
	Pid    int      `json:"pid"`
	P      int      `json:"p"`
	Uid    int      `json:"uid"`
	Title  string   `json:"title"`
	Author string   `json:"author"`
	Url    string   `json:"url"`
	R18    bool     `json:"r18"`
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Tags   []string `json:"tags"`
	DumpUrl string
}

type PostWeChatNews struct {
	MsgType string `json:"msgtype"`
	News    News   `json:"news"`
}

type News struct {
	Articles []Article `json:"articles"`
}

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Picurl      string `json:"picurl"`
}

type PostWeChatPic struct {
	MsgType string `json:"msgtype"`
	Image   Image  `json:"image"`
}

type Image struct {
	Base64 string `json:"base64"`
	Md5    string `json:"md5"`
}

type PostWeChatText struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
}

type Text struct {
	Content       string   `json:"content"`
	MentionedList []string `json:"mentioned_list"`
}
