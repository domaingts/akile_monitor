package main

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Data struct {
	Name string `gorm:"primaryKey"`
	Data string
}

type Host struct {
	Name    string `json:"name" gorm:"primaryKey"`
	DueTime int64  `json:"due_time"` // 到期时间
	BuyUrl  string `json:"buy_url"`  // 购买链接
	Seller  string `json:"seller"`   // 卖家
	Price   string `json:"price"`    // 价格
}

var db *gorm.DB
var filedb *gorm.DB

func initDb() {
	var dbfile = "file::memory:?cache=shared"
	Db, err := gorm.Open(sqlite.Open(dbfile), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}

	Db.AutoMigrate(&Data{})
	db = Db
}

func initFileDb() {
	var dbfile = "ak_monitor.db"
	Db, err := gorm.Open(sqlite.Open(dbfile), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}

	Db.AutoMigrate(&Host{})
	filedb = Db
}
