package main

import (
	"flag"
	"math/rand"
	"runtime"
	"time"

	"github.com/lingfei-zhang/dht-crawler/crawler"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	crawler.NewDHTCrawler(500, 5, 9881).Start()
}
