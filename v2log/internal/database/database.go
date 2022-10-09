package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sync"
	"time"
)

var (
	DB             *gorm.DB
	EmailSheetPool = sync.Pool{New: func() interface{} {
		return new(EmailSheet)
	}}
	IpSheetPool = sync.Pool{New: func() interface{} {
		return new(IpSheet)
	}}
	UrlSheetPool = sync.Pool{New: func() interface{} {
		return new(UrlSheet)
	}}
	RecordingSheetPool = sync.Pool{New: func() interface{} {
		return new(RecordingSheet)
	}}
)

func InitDB(dbFile string) (err error) {
	DB, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	DB.AutoMigrate(
		&EmailSheet{},
		&IpSheet{},
		&UrlSheet{},
		&RecordingSheet{},
	)

	return
}

type EmailSheet struct {
	ID    uint    `json:"id" gorm:"primaryKey"`
	Email string  `json:"email" gorm:"uniqueIndex"`
	User  *string `json:"user" gorm:"index"`
}

func (e *EmailSheet) Zero() {
	e.ID = 0
	e.Email = ""
	e.User = nil
}

func (EmailSheet) TableName() string {
	return "email_sheet"
}
func FirstOrCreateEmail(db *gorm.DB, email string) (id uint) {
	e := EmailSheetPool.Get().(*EmailSheet)
	defer EmailSheetPool.Put(e)
	e.Zero()
	e.Email = email
	db.Where("email = ?", email).FirstOrCreate(e)
	return e.ID
}

type IpSheet struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	Ip       string  `json:"ip" gorm:"uniqueIndex"`
	Nation   *string `json:"nation" gorm:"index"`
	Region   *string `json:"region"`
	Province *string `json:"province"`
	City     *string `json:"city"`
	ISP      *string `json:"isp"`
}

func (IpSheet) TableName() string {
	return "ip_sheet"
}
func (i *IpSheet) Zero() {
	i.ID = 0
	i.Ip = ""
	i.Nation = nil
	i.Region = nil
	i.Province = nil
	i.City = nil
	i.ISP = nil
}
func FirstOrCreateIp(db *gorm.DB, ip string) (id uint) {
	e := IpSheetPool.Get().(*IpSheet)
	defer IpSheetPool.Put(e)
	e.Zero()
	e.Ip = ip
	db.Where("ip = ?", ip).FirstOrCreate(e)
	return e.ID
}

type UrlSheet struct {
	ID               uint    `json:"id" gorm:"primaryKey"`
	Url              string  `json:"url" gorm:"uniqueIndex:url_port_tp"`
	Port             string  `json:"port" gorm:"uniqueIndex:url_port_tp"`
	TransferProtocol string  `json:"transfer_protocol" gorm:"uniqueIndex:url_port_tp"`
	Type             *string `json:"type" gorm:"index"`
	Company          *string `json:"company"`
	Nation           *string `json:"nation"`
	NSFW             *bool   `json:"nsfw"`
}

func (UrlSheet) TableName() string {
	return "url_sheet"
}
func (u *UrlSheet) Zero() {
	u.ID = 0
	u.Url = ""
	u.Port = ""
	u.TransferProtocol = ""
	u.Type = nil
	u.Company = nil
	u.Nation = nil
	u.NSFW = nil
}

func FirstOrCreateUrl(db *gorm.DB, url, port, tp string) (id uint) {
	e := UrlSheetPool.Get().(*UrlSheet)
	defer UrlSheetPool.Put(e)
	e.Zero()
	e.Url = url
	e.Port = port
	e.TransferProtocol = tp
	db.Where("url = ? AND port = ? AND transfer_protocol = ?", url, port, tp).FirstOrCreate(e)
	return e.ID
}

type RecordingSheet struct {
	ID        uint        `gorm:"primarykey" json:"id"`
	VisitDate time.Time   `json:"visit_date"`
	EmailId   uint        `json:"email_id" gorm:"index"`
	Email     *EmailSheet `json:"email" gorm:"-"`
	IpId      uint        `json:"ip_id" gorm:"index"`
	IP        *IpSheet    `json:"ip" gorm:"-"`
	UrlId     uint        `json:"url_id" gorm:"index"`
	Url       *UrlSheet   `json:"url" gorm:"-"`
}

func (RecordingSheet) TableName() string {
	return "recording_sheet"
}
func SaveRecord(db *gorm.DB, sheet *RecordingSheet) (id uint) {
	defer RecordingSheetPool.Put(sheet)
	db.Create(sheet)
	return sheet.ID
}
func NewRecordSheet() (s *RecordingSheet) {
	s = RecordingSheetPool.Get().(*RecordingSheet)
	s.ID = 0
	return
}
