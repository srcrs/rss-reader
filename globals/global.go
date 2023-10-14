package globals

import (
	"embed"
	"rss-reader/models"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mmcdole/gofeed"
)

var (
	DbMap    map[string]models.Feed
	RssUrls  models.Config
	Upgrader = websocket.Upgrader{}
	Lock     sync.RWMutex

	//go:embed static
	DirStatic embed.FS

	HtmlContent []byte

	Fp = gofeed.NewParser()
)

func Init() {
	conf, err := models.ParseConf()
	if err != nil {
		panic(err)
	}
	RssUrls = conf
	// 读取 index.html 内容
	HtmlContent, err = DirStatic.ReadFile("static/index.html")
	if err != nil {
		panic(err)
	}

	DbMap = make(map[string]models.Feed)
}
