package main

import (
	"flag"
	"fmt"
	"github.com/Rehtt/v2ray-tools/v2log/internal"
	"github.com/Rehtt/v2ray-tools/v2log/internal/database"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	filePath = flag.String("log_file", "/var/log/v2ray/access.log", "log文件")
	dbPath   string
	lock     sync.Mutex // sqlite并发会报错
)

func init() {
	var waitDB sync.WaitGroup
	waitDB.Add(1)
	go func() {
		var w sync.Once
		for {
			now := time.Now()
			fileName := fmt.Sprintf("v2log_%d-%d.db", now.Year(), now.Month())
			lock.Lock()
			fmt.Println(database.InitDB(fileName))
			lock.Unlock()
			w.Do(func() {
				waitDB.Done()
			})
			time.Sleep(time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0).Sub(now))
		}
	}()
	waitDB.Wait()
}
func main() {
	flag.Parse()

	var last time.Time
	lock.Lock()
	database.DB.Model(&database.RecordingSheet{}).Select("visit_date").Limit(1).Order("id desc").Find(&last)
	lock.Unlock()

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
		if info.Time.Before(last) {
			return
		}
		// 使用事务
		database.DB.Transaction(func(tx *gorm.DB) error {
			r := database.NewRecordSheet()
			r.VisitDate = info.Time
			last = info.Time
			r.EmailId = database.FirstOrCreateEmail(tx, info.Email)
			r.IpId = database.FirstOrCreateIp(tx, info.Ip)
			r.UrlId = database.FirstOrCreateUrl(tx, info.Target, info.Port, info.TransferProtocol)
			fmt.Println(database.SaveRecord(tx, r), info.Time.Format("2006/01/02-15:04:05"), info.Email, info.Target)
			return nil
		})
	})

	// 清理日志

	go func() {
		t := time.NewTicker(10 * time.Minute)
		for {
			if lock.TryLock() {
				internal.CleanFile(*filePath)
				internal.Followup()
				lock.Unlock()
			}
			<-t.C
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigs

	// 通过加锁阻止跑新的任务
	lock.Lock()
	fmt.Println("exiting")
}
