package internal

import (
	"bufio"
	"github.com/Rehtt/v2ray-tools/v2log/internal/database"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"gorm.io/gorm"
	"io"
	"log"
	"net"
	"net/http"
	"os"
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
	// 2022/09/30 17:52:56 112.103.143.64:0 accepted tcp:grpc.biliapi.net:443 email: qm@ws.com
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
	ip := strings.Split(s[2], ":")
	info.Ip = strings.Join(ip[:len(ip)-1],":")
	u := strings.Split(s[4], ":")
	info.Target = u[1]
	info.Port = u[2]
	info.Email = s[6]
	success = true
	return
}

// 完善数据库
func Followup() {
	database.DB.Transaction(func(tx *gorm.DB) error {
		var emails []*database.EmailSheet
		database.DB.Where("user IS NULL").Find(&emails)
		for i := range emails {
			emails[i].User = &strings.Split(emails[i].Email, "@")[0]
			database.DB.Model(emails[i]).Updates(emails[i])
		}

		var ips []*database.IpSheet
		database.DB.Where("nation IS NULL").Find(&ips)
		var ok bool
		var nation, region, province, city, isp string
		for i := range ips {
			nation, region, province, city, isp, ok = GetIpAddr(ips[i].Ip)
			ips[i].Nation = &nation
			ips[i].Region = &region
			ips[i].Province = &province
			ips[i].City = &city
			ips[i].ISP = &isp
			if ok {
				database.DB.Model(ips[i]).Updates(ips[i])
			}
		}

		var urls []*database.UrlSheet
		database.DB.Where("type IS NULL").Find(&urls)
		for i := range urls {
			break
			nation, _, _, _, _, ok = GetIpAddr(ips[i].Ip)
			urls[i].Nation = &nation
			database.DB.Model(urls[i]).Updates(urls[i])
		}
		return nil
	})
}

func GetIpAddr(ipStr string) (nation, region, province, city, isp string, ok bool) {
	// 国家|区域|省份|城市|ISP
	var dbPath = "ip2region.xdb"
	if _, err := os.Stat(dbPath); err != nil {
		resp, err := http.Get("https://github.com/lionsoul2014/ip2region/raw/master/data/ip2region.xdb")
		if err != nil {
			log.Println("下载ip2region.xdb失败：", err.Error())
		}
		defer resp.Body.Close()
		f, _ := os.Create(dbPath)
		f.ReadFrom(resp.Body)
		defer f.Close()
	}
	searcher, _ := xdb.NewWithFileOnly(dbPath)
	defer searcher.Close()

	ip := net.ParseIP(ipStr)
	if ip == nil {
		i, err := net.ResolveIPAddr("ip", ipStr)
		if err != nil {
			log.Println("找不到ip：", ipStr)
			return
		}
		ip = i.IP
	}

	data, err := searcher.SearchByStr(ip.To16().String())
	if err != nil {
		log.Println("找不到ip记录：", ipStr, err.Error())
		return
	}
	s := strings.Split(data, "|")

	nation = s[0]
	region = s[1]
	province = s[2]
	city = s[3]
	isp = s[4]
	ok = true
	return
}
