package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mmcdole/gofeed"
)

var ctx = context.Background()
var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "xx.xx.xx.xx:6399",
		Password: "xxxxxx",
		DB:       0,
	})
}

func main() {
	go updateFeeds()
	http.HandleFunc("/feeds", getFeedsHandler)
	http.HandleFunc("/", serveHome)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func updateFeeds() {
	rssUrls := []string{
		// 添加您的RSS订阅链接
		"https://www.zhihu.com/rss",
		"https://tech.meituan.com/feed/",
		"http://www.ruanyifeng.com/blog/atom.xml",
		"https://cn.wsj.com/zh-hans/rss",
		"https://feeds.appinn.com/appinns/",
		"https://v2ex.com/feed/tab/tech.xml",
                "https://hostloc.com/forum.php?mod=rss&fid=45&auth=389ec3vtQanmEuRoghE%2FpZPWnYCPmvwWgSa7RsfjbQ%2BJpA%2F6y6eHAx%2FKqtmPOg",
	}

	for {
		for _, url := range rssUrls {
			fp := gofeed.NewParser()
			feed, err := fp.ParseURL(url)
			if err != nil {
				log.Printf("Error fetching feed: %v", err)
				continue
			}

			feedJSON, err := json.Marshal(feed)
			if err != nil {
				log.Printf("Error marshaling feed: %v", err)
				continue
			}

			err = rdb.Set(ctx, url, feedJSON, 0).Err()
			if err != nil {
				log.Printf("Error saving feed to Redis: %v", err)
			}
		}
		time.Sleep(10 * time.Minute)
	}
}

func getFeedsHandler(w http.ResponseWriter, r *http.Request) {
	rssUrls := []string{
		// 添加您的RSS订阅链接
		"https://www.zhihu.com/rss",
		"https://tech.meituan.com/feed/",
		"http://www.ruanyifeng.com/blog/atom.xml",
		"https://cn.wsj.com/zh-hans/rss",
		"https://feeds.appinn.com/appinns/",
		"https://v2ex.com/feed/tab/tech.xml",
                "https://hostloc.com/forum.php?mod=rss&fid=45&auth=389ec3vtQanmEuRoghE%2FpZPWnYCPmvwWgSa7RsfjbQ%2BJpA%2F6y6eHAx%2FKqtmPOg",
	}

	feeds := make([]gofeed.Feed, 0, len(rssUrls))
	for _, url := range rssUrls {
		feedJSON, err := rdb.Get(ctx, url).Result()
		if err != nil {
			log.Printf("Error getting feed from Redis: %v", err)
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

