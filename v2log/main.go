package main

import (
	"flag"
	"fmt"
	"github.com/Rehtt/v2ray-tools/v2log/internal"
	"github.com/Rehtt/v2ray-tools/v2log/internal/database"
	"sync"
	"time"
)

var (
	filePath = flag.String("log_file", "/var/log/v2ray/access.log", "log文件")
	dbPath   string
	lock     sync.Mutex // sqlite并发会报错
)

func init() {
	go func() {
		for {
			now := time.Now()
			lock.Lock()
			database.InitDB(fmt.Sprintf("v2log_%d-%s.db", now.Year(), now.Month()))
			lock.Unlock()
			time.Sleep(time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.Local).Sub(now))
		}
	}()
}
func main() {
	flag.Parse()

	internal.Tail(*filePath, func(text string) {
		lock.Lock()
		defer lock.Unlock()
		if text == "" {
			return
		}
		info, ok := internal.Split(text)
		if !ok {
			return
		}
		defer internal.InfoPool.Put(info)
		r := database.NewRecordSheet()
		r.VisitDate = info.Time
		r.EmailId = database.FirstOrCreateEmail(info.Email)
		r.IpId = database.FirstOrCreateIp(info.Ip)
		r.UrlId = database.FirstOrCreateUrl(info.Target, info.Port)
		database.SaveRecord(r)
	})

	// 清理日志
	t := time.NewTicker(time.Hour)
	for {
		internal.CleanFile(*filePath)

		lock.Lock()
		internal.Followup()
		lock.Unlock()
		<-t.C
	}
}
