package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/astaxie/beego/validation"
	"github.com/hprose/hprose-golang/rpc"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
)

type (

	//RequstJob POST 请求参数获取
	RequstJob struct {
		Title      string   `json:"title" valid:"Required; MaxSize(12)"`     // 职位标题
		Position   string   `json:"position" valid:"Required; MaxSize(12);"` // 原职位分类
		Company    string   `json:"company"`                                 // 公司名
		Category   int      `json:"category"  valid:"Range(101,129);"`       // 分类
		Area       int      `json:"area"`                                    // 地区
		MinPay     int      `json:"min_pay"`                                 // 最小月薪
		MaxPay     int      `json:"max_pay"`                                 // 最大月薪
		Education  int      `json:"education"`                               // 学历
		Experience int      `json:"experience"`                              // 工作经验
		Welfare    []int    `json:"welfare"`                                 //   valid:"Range(400,411);"  未清楚怎么使用子集验证
		Tags       []string `json:"tags"`
		Intro      string   `json:"intro"`
		SourceFrom string   `json:"source_from"` // string默认长度为255, 使用这种tag重设。
		CompanyURL string   `json:"company_url"` // string默认长度为255, 使用这种tag重设。
		Linkman    string   `json:"linkman"`
		Telephone  string   `json:"telephone"`
		Email      string   `json:"email"`
		Address    string   `json:"address"`
		Lng        float64  `json:"lng"`
		Lat        float64  `json:"lat"`
	}
	// Job 数据库
	Job struct {
		gorm.Model
		Title      string         `gorm:"size:255"`   // 职位标题
		Position   string         `gorm:"size:255"`   // 原职位分类
		Company    string         `gorm:"size:255"`   // 公司名
		Param      pq.Int64Array  `gorm:"type:int[]"` // 标签
		Category   int            // 分类
		Area       int            // 地区
		MinPay     int            // 最小月薪
		MaxPay     int            // 最大月薪
		Education  int            // 学历
		Experience int            // 工作经验
		Intro      string         `gorm:"type:text"` // 职位介绍
		Rank       float32        // 排序
		Tags       pq.StringArray `gorm:"type:text[]"` // 标签
		SourceFrom string         `gorm:"size:255"`    // string默认长度为255, 使用这种tag重设。
		CompanyURL string         `gorm:"size:255"`    // string默认长度为255, 使用这种tag重设。
		Linkman    string         `gorm:"size:255"`
		Telephone  string         `gorm:"size:255"`
		Email      string         `gorm:"size:255"`
		Lng        float64
		Lat        float64
		Address    string
	}
	// Jobs struct{
	// 	job[] *Job
	// }
)

// 数据库对象

var db *gorm.DB

func init() {
	//db, _ = gorm.Open("postgres", "host=localhost user=postgres dbname=spider password=123456 sslmode=disable")
	db, _ = gorm.Open("postgres", "host=192.157.192.118 user=xiaoyi dbname=spider sslmode=disable password=123456")

	db.AutoMigrate(&Job{})
}

func main() {
	service := rpc.NewHTTPService()
	service.AddFunction("save", save, rpc.Options{})
	http.ListenAndServe(":818", service)
}

func save(str string) string {
	// fmt.Println(str)
	var j RequstJob
	var err error
	json.Unmarshal([]byte(str), &j)

	valid := validation.Validation{}

	b, err := valid.Valid(&j)
	if err != nil {
		// handle error
	}
	if !b {
		// validation does not pass
		// blabla...
		for _, err := range valid.Errors {
			log.Println(err.Key, err.Message)
		}
		return string("数据异常")
	}

	if j.SourceFrom == "" {
		return string("同步职位失败")
	}

	var job Job

	db.Where(Job{SourceFrom: j.SourceFrom}).FirstOrCreate(&job)
	err = RequstJobSaveData(&job, j)
	if err != nil {
		return "err: " + err.Error() + "!"
	}
	// fmt.Println(job.Param, job.Tags)

	db.Save(&job)

	fmt.Println(job.ID, job.Category, job.Company, job.Title)
	jobString, _ := json.Marshal(job)
	return string(jobString)
}

//RequstJobSaveData 把请求的数据包转成数据模型中的参数
func RequstJobSaveData(job *Job, rj RequstJob) error {

	job.Title = rj.Title
	job.Position = rj.Position
	job.Company = rj.Company
	job.Category = rj.Category
	job.Area = rj.Area
	job.MinPay = rj.MinPay
	job.MaxPay = rj.MaxPay
	job.Education = rj.Education
	job.Experience = rj.Experience
	job.Intro = rj.Intro
	// job.Welfare = "{" + strings.Join(rj.Welfare, ",") + "}"
	job.CompanyURL = rj.CompanyURL
	job.SourceFrom = rj.SourceFrom
	job.Linkman = rj.Linkman
	job.Telephone = rj.Telephone
	job.Email = rj.Email
	job.Address = rj.Address
	job.Lat = rj.Lat
	job.Lng = rj.Lng
	param := []int64{}
	// var param map[int]int

	if rj.Category != 0 {
		param = append(param, int64(rj.Category))
	}
	// if err := checkCategory(rj.Category); err != nil {

	// 	return job, err
	// }

	if rj.Area != 0 {
		param = append(param, int64(rj.Area))
	}
	if rj.Education != 0 {
		param = append(param, int64(rj.Education))
	}
	if rj.Experience != 0 {
		param = append(param, int64(rj.Experience))
	}

	for _, tag := range rj.Welfare {
		param = append(param, int64(tag))
	}

	job.Param = param
	job.Tags = rj.Tags

	return nil
}

// func checkCategory(val int) error {
// 	validate := validator.New()

// 	fn := func(fl validator.FieldLevel) bool {
// 		fmt.Println(fl.Field().String())
// 		return true
// 	}

// 	validate.RegisterValidation("isequaltestfunc", fn)

// 	err2 := validate.Var(val, "isequaltestfunc")
// 	fmt.Println(err2)
// 	return err2
// }
