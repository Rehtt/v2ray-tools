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
	Email     string    `json:"email" gorm:"uniqueIndex"`
	User      *string   `json:"user" gorm:"index"`
}

func (EmailSheet) TableName() string {
	return "email_sheet"
}
func FirstOrCreateEmail(db *gorm.DB, email string) (id uint) {
	e := EmailSheetPool.Get().(*EmailSheet)
	defer EmailSheetPool.Put(e)
	db.Where("email = ?", email).Find(e)
	if e.ID != 0 {
		return e.ID
	}
	e.Email = email
	db.Create(e)
	return e.ID
}

type IpSheet struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	Ip        string    `json:"ip" gorm:"uniqueIndex"`
	Nation    *string   `json:"nation" gorm:"index"`
	Province  *string   `json:"province"`
	City      *string   `json:"city"`
	Region    *string   `json:"region"`
	Street    *string   `json:"street"`
}

func (IpSheet) TableName() string {
	return "ip_sheet"
}
func FirstOrCreateIp(db *gorm.DB, ip string) (id uint) {
	e := IpSheetPool.Get().(*IpSheet)
	defer IpSheetPool.Put(e)
	db.Where("ip = ?", ip).Find(e)
	if e.ID != 0 {
		return e.ID
	}
	e.Ip = ip
	db.Create(e)
	return e.ID
}

type UrlSheet struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	Url       string    `json:"url" gorm:"index:url_port"`
	Port      string    `json:"port" gorm:"index:url_port"`
	Type      *string   `json:"type" gorm:"index"`
	Company   *string   `json:"company"`
	Nation    *string   `json:"nation"`
	NSFW      *bool     `json:"nsfw"`
}

func (UrlSheet) TableName() string {
	return "url_sheet"
}

func FirstOrCreateUrl(db *gorm.DB, url, port string) (id uint) {
	e := UrlSheetPool.Get().(*UrlSheet)
	defer UrlSheetPool.Put(e)
	db.Where("url = ? AND port = ?", url, port).Find(e)
	if e.ID != 0 {
		return e.ID
	}
	e.Url = url
	e.Port = port
	db.Create(e)
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
