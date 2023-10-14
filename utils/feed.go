package utils

import (
	"log"
	"os"
	"rss-reader/globals"
	"rss-reader/models"
	"time"
)

func UpdateFeeds() {
	var (
		tick = time.Tick(time.Duration(globals.RssUrls.ReFresh) * time.Minute)
	)
	for {
		formattedTime := time.Now().Format("2006-01-02 15:04:05")
		for _, url := range globals.RssUrls.Values {
			go UpdateFeed(url, formattedTime)
		}
		<-tick
	}
}

func UpdateFeed(url, formattedTime string) {
	result, err := globals.Fp.ParseURL(url)
	if err != nil {
		log.Printf("Error fetching feed: %v | %v", url, err)
		return
	}
	//feed内容无更新时无需更新缓存
	if cache, ok := globals.DbMap[url]; ok &&
		len(result.Items) > 0 &&
		len(cache.Items) > 0 &&
		result.Items[0].Link == cache.Items[0].Link {
		return
	}
	customFeed := models.Feed{
		Title:  result.Title,
		Link:   result.Link,
		Custom: map[string]string{"lastupdate": formattedTime},
		Items:  make([]models.Item, 0, len(result.Items)),
	}
	for _, v := range result.Items {
		customFeed.Items = append(customFeed.Items, models.Item{
			Link:        v.Link,
			Title:       v.Title,
			Description: v.Description,
		})
	}
	globals.Lock.Lock()
	defer globals.Lock.Unlock()
	globals.DbMap[url] = customFeed
}

//获取feeds列表
func GetFeeds() []models.Feed {
	feeds := make([]models.Feed, 0, len(globals.RssUrls.Values))
	for _, url := range globals.RssUrls.Values {
		globals.Lock.RLock()
		cache, ok := globals.DbMap[url]
		globals.Lock.RUnlock()
		if !ok {
			log.Printf("Error getting feed from db is null %v", url)
			continue
		}

		feeds = append(feeds, cache)
	}
	return feeds
}

func WatchConfigFileChanges(filePath string) {
	// 获取初始文件信息
	initialFileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Println("无法获取文件信息:", err)
		return
	}

	for {
		// 每隔一段时间检查文件是否有变化
		time.Sleep(7 * time.Second)

		// 获取最新的文件信息
		currentFileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Println("无法获取文件信息:", err)
			return
		}

		// 检查文件的修改时间是否有变化
		if currentFileInfo.ModTime() != initialFileInfo.ModTime() {
			log.Println("文件已修改")
			initialFileInfo = currentFileInfo
			globals.Init()
			formattedTime := time.Now().Format("2006-01-02 15:04:05")
			for _, url := range globals.RssUrls.Values {
				go UpdateFeed(url, formattedTime)
			}
		}
	}
}
