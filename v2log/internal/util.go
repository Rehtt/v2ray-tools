package internal

import (
	"bufio"
	"github.com/Rehtt/v2ray-tools/v2log/internal/database"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func Tail(filePath string, f func(text string)) {
	c := exec.Command("tail", "-f", filePath)
	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}
	if err = c.Start(); err != nil {
		log.Fatalln(err)
	}
	go func(stdout io.ReadCloser) {
		buf := bufio.NewScanner(stdout)
		buf.Split(bufio.ScanLines)
		for buf.Scan() {
			f(buf.Text())
		}
	}(stdout)
}

func CleanFile(filePath string) error {
	return exec.Command("bash", "-c", "> "+filePath).Run()
}

type Info struct {
	Time   time.Time
	Ip     string
	Target string
	Port   string
	Email  string
}

var (
	InfoPool = sync.Pool{New: func() interface{} {
		return new(Info)
	}}
)

func Split(str string) (info *Info, success bool) {
	info = InfoPool.Get().(*Info)
	// 2022/09/29 11:36:22 113.89.232.233:0 accepted tcp:play.google.com:443 email: rehtt@vless.com
	var err error
	s := strings.Split(str, " ")
	if len(s) != 7 {
		return
	}
	info.Time, err = time.Parse("2006/01/02 15:04:05", strings.Join(s[:2], " "))
	if err != nil {
		return
	}
	info.Ip = strings.Split(s[2], ":")[0]
	u := strings.Split(s[4], ":")
	info.Target = u[1]
	info.Port = u[2]
	info.Email = s[6]
	success = true
	return
}

// 完善数据库
func Followup() {
	var emails []*database.EmailSheet
	database.DB.Where("user IS NULL").Find(&emails)
	for i := range emails {
		emails[i].User = &strings.Split(emails[i].Email, "@")[0]
		database.DB.Model(emails[i]).Updates(emails[i])
	}

	var ips []*database.IpSheet
	database.DB.Where("nation IS NULL").Find(&ips)
	for i := range ips {
		// todo
		database.DB.Model(ips[i]).Updates(ips[i])
	}

	var urls []*database.UrlSheet
	database.DB.Where("type IS NULL").Find(&urls)
	for i := range urls {
		// todo
		database.DB.Model(urls[i]).Updates(urls[i])
	}
}
