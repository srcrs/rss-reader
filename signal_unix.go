//go:build unix
// +build unix

package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func handleSignal() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGUSR1)
	defer close(ch)

	for {
		<-ch
		conf, err := parseConf()
		if err != nil {
			continue
		}
		increment := rssUrls.getIncrement(conf)
		rssUrls = conf
		now := time.Now().Format("2006-01-02 15:04:05")
		for _, item := range increment {
			go updateFeed(item, now)
		}
	}
}
