package main

import (
	"embed"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mmcdole/gofeed"
)

type Config struct {
	Values         []string `json:"values"`
	ReFresh        int      `json:"refresh"`
	AutoUpdatePush int      `json:"autoUpdatePush"`
}

var (
	dbMap    sync.Map
	rssUrls  Config
	upgrader = websocket.Upgrader{}

	//go:embed static
	dirStatic embed.FS
	//go:embed index.html
	fileIndex embed.FS

	htmlContent []byte
)

func init() {
	// 读取配置文件
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	// 解析JSON数据到Config结构体
	err = json.Unmarshal(data, &rssUrls)
	if err != nil {
		panic(err)
	}
	// 读取 index.html 内容
	htmlContent, err = fileIndex.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
}

func main() {
	go updateFeeds()
	http.HandleFunc("/feeds", getFeedsHandler)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", serveHome)

	//加载静态文件
	fs := http.FileServer(http.FS(dirStatic))
	http.Handle("/static/", fs)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlContent)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %v", err)
		return
	}
	defer conn.Close()
	for {
		for _, url := range rssUrls.Values {
			feedJSON, ok := dbMap.Load(url)
			if !ok {
				log.Printf("Error getting feed from db is null %v", url)
				continue
			}
			err = conn.WriteMessage(websocket.TextMessage, []byte(feedJSON.(string)))
			//错误直接关闭更新
			if err != nil {
				log.Printf("Error sending message or Connection closed: %v", err)
				return
			}
		}
		//如果未配置则不自动更新
		if rssUrls.AutoUpdatePush == 0 {
			return
		}
		time.Sleep(time.Duration(rssUrls.AutoUpdatePush) * time.Minute)
	}
}

func updateFeeds() {
	var (
		tick          = time.Tick(time.Duration(rssUrls.ReFresh) * time.Minute)
		fp            = gofeed.NewParser()
		formattedTime = time.Now().Format("2006-01-02 15:04:05")
	)
	for {
		for _, url := range rssUrls.Values {
			go updateFeed(fp, url, formattedTime)
		}
		<-tick
	}
}

func updateFeed(fp *gofeed.Parser, url, formattedTime string) {
	feed, err := fp.ParseURL(url)
	if err != nil {
		log.Printf("Error fetching feed: %v | %v", url, err)
		return
	}

	feed.Custom = map[string]string{"lastupdate": formattedTime}

	feedJSON, err := json.Marshal(feed)
	if err != nil {
		log.Printf("Error marshaling feed: %v", err)
		return
	}
	dbMap.Store(url, string(feedJSON))
}

func getFeedsHandler(w http.ResponseWriter, r *http.Request) {

	feeds := make([]gofeed.Feed, 0, len(rssUrls.Values))
	for _, url := range rssUrls.Values {
		feedJSON, ok := dbMap.Load(url)
		if !ok {
			log.Printf("Error getting feed from db is null %v", url)
			continue
		}

		var feed gofeed.Feed
		if err := json.Unmarshal([]byte(feedJSON.(string)), &feed); err != nil {
			log.Printf("Error unmarshaling feed: %v", err)
			continue
		}

		feeds = append(feeds, feed)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feeds)
}
