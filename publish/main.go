package main

import (
	"encoding/json"
	"time"

	"fmt"

	"github.com/hprose/hprose-golang/rpc"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/yizenghui/spider/tongcheng"
)

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open("sqlite3", "gz_job.db")
	if err != nil {
		panic("连接数据库失败")
	}
	// 自动迁移模式
	db.AutoMigrate(&tongcheng.Job{})
	defer db.Close()
	PostTask()
}

//PostTask 同步任务
func PostTask() {
	ticker := time.NewTicker(time.Second * 2)
	for _ = range ticker.C {
		go Publish()
	}
}

//Stub rpc 服务器提供接口
type Stub struct {
	Save      func(string) (string, error)
	AsyncSave func(func(string, error), string) `name:"save"`
}

// Publish 发布
func Publish() {
	var job tongcheng.Job
	db.Where("publish_at = 0").First(&job)
	if job.ID > 0 {
		postJob := tongcheng.TransformJob(job)
		job.PublishAt = time.Now().Unix()
		db.Save(&job)
		fmt.Println(job.ID, job.Title, job.Company)
		client := rpc.NewClient("http://127.0.0.1:818/")
		var stub *Stub
		client.UseService(&stub)
		if jsonStr, err := json.Marshal(postJob); err == nil {
			_, err := stub.Save(string(jsonStr))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
