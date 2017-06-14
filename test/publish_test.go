package serve_test

import (
	"encoding/json"
	"testing"
	"time"

	"fmt"

	"github.com/hprose/hprose-golang/rpc"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/yizenghui/spider/tongcheng"
)

func Test_Publish(t *testing.T) {

	type Stub struct {
		Save      func(string) (string, error)
		AsyncSave func(func(string, error), string) `name:"save"`
	}

	db, err := gorm.Open("sqlite3", "gz_job.db")
	// db, err := gorm.Open("postgres", "host=localhost user=postgres dbname=spider sslmode=disable password=123456")

	if err != nil {
		panic("连接数据库失败")
	}

	// 自动迁移模式
	db.AutoMigrate(&tongcheng.Job{})
	defer db.Close()
	var job tongcheng.Job
	db.Where("publish_at = 0").First(&job)
	// fmt.Println(job)
	postJob := tongcheng.TransformJob(job)
	// fmt.Println(postJob)

	job.PublishAt = time.Now().Unix()
	db.Save(&job)
	client := rpc.NewClient("http://127.0.0.1:818/")
	var stub *Stub
	client.UseService(&stub)
	// stub.AsyncSave(func(result string, err error) {
	// 	fmt.Println(result, err)
	// }, `{"title":"异步标题"}`)

	if json_str, err := json.Marshal(postJob); err == nil {
		// fmt.Println("================struct 到json str==")
		// fmt.Println(string(json_str))
		fmt.Println(stub.Save(string(json_str)))

	}
}
