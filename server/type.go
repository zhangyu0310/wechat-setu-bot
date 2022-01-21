package server

// Query get setu info. TODO: Annotate field can be assigned
type Query struct {
	R18 int `json:"r18"`
	Num int `json:"num"`
	// Uid int `json:"uid"`
	// KeyWord string `json:"keyword"`
	Tag  []string `json:"tag"`
	Size []string `json:"size"`
	// Proxy string `json:"proxy"`
	// DateAfter int `json:"dateAfter"`
	// DateBefore int `json:"dateBefore"`
	// Dsc bool `json:"dsc"`
}

type Result struct {
	Error    string `json:"error"`
	Setus    []Setu `json:"data"`
	picPaths []string
}

func (result *Result) setPicPath(path string) {
	result.picPaths = append(result.picPaths, path)
}

func (result *Result) getPicPath(index uint) string {
	return result.picPaths[index]
}

type PicUrl struct {
	Original string `json:"original"`
	Regular  string `json:"regular"`
	Small    string `json:"small"`
	Thumb    string `json:"thumb"`
	Mini     string `json:"mini"`
}

type Setu struct {
	Pid     int      `json:"pid"`
	P       int      `json:"p"`
	Uid     int      `json:"uid"`
	Title   string   `json:"title"`
	Author  string   `json:"author"`
	R18     bool     `json:"r18"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
	Tags    []string `json:"tags"`
	Ext     string   `json:"ext"`
	Date    int      `json:"uploadDate"`
	Urls    PicUrl   `json:"urls"`
	DumpUrl string
}

type BotMsgType string

const (
	BotMsgText  BotMsgType = "text"
	BotMsgNews  BotMsgType = "news"
	BotMsgImage BotMsgType = "image"
)

type BotMsgReq struct {
	MsgType BotMsgType `json:"msgtype"`
	News    *News      `json:"news,omitempty"`
	Image   *Image     `json:"image,omitempty"`
	Text    *Text      `json:"text,omitempty"`
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

type Image struct {
	Base64 string `json:"base64"`
	Md5    string `json:"md5"`
}

type Text struct {
	Content       string   `json:"content"`
	MentionedList []string `json:"mentioned_list"`
}
