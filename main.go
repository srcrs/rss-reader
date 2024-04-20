package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"rss-reader/globals"
	"rss-reader/models"

	"rss-reader/utils"
	"time"

	"github.com/gorilla/websocket"
)

func init() {
	globals.Init()
}

func main() {
	go utils.UpdateFeeds()
	go utils.WatchConfigFileChanges("config.json")
	http.HandleFunc("/feeds", getFeedsHandler)
	http.HandleFunc("/ws", wsHandler)
	// http.HandleFunc("/", serveHome)
	http.HandleFunc("/", tplHandler)

	//加载静态文件
	fs := http.FileServer(http.FS(globals.DirStatic))
	http.Handle("/static/", fs)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Write(globals.HtmlContent)
}

func tplHandler(w http.ResponseWriter, r *http.Request) {
	// 创建一个新的模板，并设置自定义分隔符为<< >>，避免与Vue的语法冲突
	tmplInstance := template.New("index.html").Delims("<<", ">>")
	//添加加法函数计数
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
	// 加载模板文件
	tmpl, err := tmplInstance.Funcs(funcMap).ParseFS(globals.DirStatic, "static/index.html")
	if err != nil {
		log.Println("模板加载错误:", err)
		return
	}

	//判断现在是否是夜间
	formattedTime := time.Now().Format("15:04:05")
	darkMode := false
	if globals.RssUrls.NightStartTime != "" && globals.RssUrls.NightEndTime != "" {
		if globals.RssUrls.NightStartTime > formattedTime || formattedTime > globals.RssUrls.NightEndTime {
			darkMode = true
		}
	}

	// 定义一个数据对象
	data := struct {
		Keywords       string
		RssDataList    []models.Feed
		DarkMode       bool
		AutoUpdatePush int
	}{
		Keywords:       getKeywords(),
		RssDataList:    utils.GetFeeds(),
		DarkMode:       darkMode,
		AutoUpdatePush: globals.RssUrls.AutoUpdatePush,
	}

	// 渲染模板并将结果写入响应
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("模板渲染错误:", err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := globals.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %v", err)
		return
	}

	defer conn.Close()
	for {
		for _, url := range globals.RssUrls.Values {
			globals.Lock.RLock()
			cache, ok := globals.DbMap[url]
			globals.Lock.RUnlock()
			if !ok {
				log.Printf("Error getting feed from db is null %v", url)
				continue
			}
			data, err := json.Marshal(cache)
			if err != nil {
				log.Printf("json marshal failure: %s", err.Error())
				continue
			}

			err = conn.WriteMessage(websocket.TextMessage, data)
			//错误直接关闭更新
			if err != nil {
				log.Printf("Error sending message or Connection closed: %v", err)
				return
			}
		}
		//如果未配置则不自动更新
		if globals.RssUrls.AutoUpdatePush == 0 {
			return
		}
		time.Sleep(time.Duration(globals.RssUrls.AutoUpdatePush) * time.Minute)
	}
}

//获取关键词也就是title
//获取feeds列表
func getKeywords() string {
	words := ""
	for _, url := range globals.RssUrls.Values {
		globals.Lock.RLock()
		cache, ok := globals.DbMap[url]
		globals.Lock.RUnlock()
		if !ok {
			log.Printf("Error getting feed from db is null %v", url)
			continue
		}
		if cache.Title != "" {
			words += cache.Title + ","
		}
	}
	return words
}

func getFeedsHandler(w http.ResponseWriter, r *http.Request) {
	feeds := utils.GetFeeds()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feeds)
}
