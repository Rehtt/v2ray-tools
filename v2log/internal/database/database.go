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
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
	User      *string   `json:"user"`
}

func (EmailSheet) TableName() string {
	return "email_sheet"
}
func FirstOrCreateEmail(email string) (id uint) {
	e := EmailSheetPool.Get().(*EmailSheet)
	defer EmailSheetPool.Put(e)
	DB.Where(EmailSheet{Email: email}).FirstOrCreate(e)
	return e.ID
}

type IpSheet struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	Ip        string    `json:"ip"`
	Nation    *string   `json:"nation"`
	Province  *string   `json:"province"`
	City      *string   `json:"city"`
	Region    *string   `json:"region"`
	Street    *string   `json:"street"`
}

func (IpSheet) TableName() string {
	return "ip_sheet"
}
func FirstOrCreateIp(ip string) (id uint) {
	e := IpSheetPool.Get().(*IpSheet)
	defer IpSheetPool.Put(e)
	DB.Where(IpSheet{Ip: ip}).FirstOrCreate(e)
	return e.ID
}

type UrlSheet struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	Url       string    `json:"url"`
	Port      string    `json:"port" `
	Type      *string   `json:"type"`
	Company   *string   `json:"company"`
	Nation    *string   `json:"nation"`
	NSFW      *bool     `json:"nsfw"`
}

func (UrlSheet) TableName() string {
	return "url_sheet"
}

func FirstOrCreateUrl(url, port string) (id uint) {
	e := UrlSheetPool.Get().(*UrlSheet)
	defer UrlSheetPool.Put(e)
	DB.Where(UrlSheet{Url: url, Port: port}).FirstOrCreate(e)
	return e.ID
}

type RecordingSheet struct {
	ID        uint        `gorm:"primarykey" json:"id"`
	VisitDate time.Time   `json:"visit_date"`
	EmailId   uint        `json:"email_id"`
	Email     *EmailSheet `json:"email" gorm:"-"`
	IpId      uint        `json:"ip_id"`
	IP        *IpSheet    `json:"ip" gorm:"-"`
	UrlId     uint        `json:"url_id"`
	Url       *UrlSheet   `json:"url" gorm:"-"`
}

func (RecordingSheet) TableName() string {
	return "recording_sheet"
}
func SaveRecord(sheet *RecordingSheet) (id uint) {
	defer RecordingSheetPool.Put(sheet)
	DB.Create(sheet)
	return sheet.ID
}
func NewRecordSheet() (s *RecordingSheet) {
	s = RecordingSheetPool.Get().(*RecordingSheet)
	s.ID = 0
	return s
}
