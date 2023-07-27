package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/gorilla/websocket"
	"github.com/mmcdole/gofeed"
)

type Config struct {
	Values []string `json:"values"`
}

var (
	db       *badger.DB
	rssUrls  Config
	upgrader = websocket.Upgrader{}
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
	initDB()
}

func initDB() error {
	var err error
	options := badger.DefaultOptions("db").WithTruncate(false)
	db, err = badger.Open(options)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	go updateFeeds()
	http.HandleFunc("/feeds", getFeedsHandler)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", serveHome)

	//加载静态文件
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	for _, url := range rssUrls.Values {
		feedJSON, err := get(url)
		if err != nil {
			log.Printf("Error getting feed from Redis: %v", err)
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(feedJSON))
		if err != nil {
			log.Printf("Error sending message: %v", err)
			continue
		}
	}
}

func updateFeeds() {
	for {
		for _, url := range rssUrls.Values {
			fp := gofeed.NewParser()
			feed, err := fp.ParseURL(url)
			if err != nil {
				log.Printf("Error fetching feed: %v | %v", url, err)
				continue
			}

			feedJSON, err := json.Marshal(feed)
			if err != nil {
				log.Printf("Error marshaling feed: %v", err)
				continue
			}

			err = update(url, string(feedJSON))
			if err != nil {
				log.Printf("Error saving feed to Redis: %v", err)
			}
		}
		time.Sleep(5 * time.Minute)
	}
}

func getFeedsHandler(w http.ResponseWriter, r *http.Request) {

	feeds := make([]gofeed.Feed, 0, len(rssUrls.Values))
	for _, url := range rssUrls.Values {
		feedJSON, err := get(url)
		if err != nil {
			log.Printf("Error getting feed from Redis: %v | %v", url, err)
			continue
		}

		var feed gofeed.Feed
		err = json.Unmarshal([]byte(feedJSON), &feed)
		if err != nil {
			log.Printf("Error unmarshaling feed: %v", err)
			continue
		}

		feeds = append(feeds, feed)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feeds)
}

func update(key, value string) error {
	// 写入数据
	if err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(value))
		return err
	}); err != nil {
		return err
	}
	return nil
}

func get(key string) (string, error) {
	value := make([]byte, 0)
	// 读取数据
	if err := db.View(func(txn *badger.Txn) error {
		if item, err := txn.Get([]byte(key)); err != nil {
			return err
		} else {
			if value, err = item.ValueCopy(nil); err != nil {
				return err
			} else {
				return nil
			}
		}
	}); err != nil {
		return "", err
	}
	if string(value) == "" {
		return "", fmt.Errorf("person not found")
	}
	return string(value), nil
}
