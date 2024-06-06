package internal

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/Rehtt/Kit/util"
	"github.com/Rehtt/v2ray-tools/v2log/internal/database"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"gorm.io/gorm"
)

func Tail(filePath string, f func(text string)) {
	c := exec.Command("tail", "-f", "-n", "+0", filePath)
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
	Time             time.Time
	Ip               string
	Target           string
	Port             string
	TransferProtocol string
	Email            string
}

var (
	InfoPool = sync.Pool{New: func() interface{} {
		return new(Info)
	}}
)

func Split(str string) (info *Info, success bool) {
	info = InfoPool.Get().(*Info)
	// 2022/09/30 17:52:56 12.103.143.64:0 accepted tcp:grpc.biliapi.net:443 email: qm@ws.com
	// 2022/09/29 11:36:22 13.89.232.233:0 accepted tcp:play.google.com:443 email: rehtt@vless.com
	// 2022/09/30 23:02:54 tcp:1.155.158.187:0 accepted tcp:edge.microsoft.com:443 email: rehtt@vless.com
	var err error
	s := strings.Split(str, " ")
	if len(s) < 7 {
		return
	}
	info.Time, err = time.ParseInLocation("2006/01/02 15:04:05", strings.Join(s[:2], " "), time.Local)
	if err != nil {
		return
	}
	ip := strings.Split(s[2], ":")
	if ip[0] == "tcp" {
		info.Ip = strings.Join(ip[1:len(ip)-1], ":")
	} else {
		info.Ip = strings.Join(ip[:len(ip)-1], ":")
	}

	u := strings.Split(s[4], ":")
	if len(u) != 2 {
		return
	}
	info.TransferProtocol = u[0]
	info.Target = u[1]
	info.Port = u[2]
	for i, v := range s {
		if v == "email:" {
			info.Email = s[i+1]
			break
		}
	}
	if info.Email == "" {
		return
	}
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
	if ipStr[0] == '[' {
		return
	}
	// 国家|区域|省份|城市|ISP
	var dbPath = "ip2region.xdb"

	info, err := util.GetGitHubFileInfo("lionsoul2014", "ip2region", "data/ip2region.xdb")
	if err != nil {
		log.Println("get hash error:", err)
	}
	if GetHash(info.Name) != info.Sha {
		SetHash(info.Name, info.Sha)
		resp, err := http.Get(info.DownloadUrl)
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

type hash struct {
	Hash string `json:"hash"`
	Time string `json:"time"`
}

func GetHash(key string) string {
	return openHashFile()[key].Hash
}
func SetHash(key, value string) {
	tmp := openHashFile()
	tmp[key] = hash{
		Hash: value,
		Time: time.Now().Format("2006-01-02"),
	}
	f, err := os.Create("hash")
	if err != nil {
		log.Println("save hash file error:", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(tmp)
}
func openHashFile() (out map[string]hash) {
	out = make(map[string]hash)
	f, err := os.Open("hash")
	if err != nil {
		return
	}
	defer f.Close()
	json.NewDecoder(f).Decode(&out)
	return out
}
