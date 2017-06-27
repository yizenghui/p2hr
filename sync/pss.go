package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

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
		Rank       float64        // 排序
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
	db, _ = gorm.Open("postgres", "host=localhost user=postgres dbname=spider password=123456 sslmode=disable")
	// db, _ = gorm.Open("postgres", "host=192.157.192.118 user=xiaoyi dbname=spider sslmode=disable password=123456")

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

	// TODO 获取票数
	vote := 1   // 支持
	devote := 0 // 反对
	level := 0  //级别
	// 获取排行分数
	job.Rank = GetRank(vote, devote, time.Now().Unix(), level)
	db.Save(&job)

	fmt.Println(job.ID, job.Category, job.Company, job.Title, job.Rank)
	jobString, _ := json.Marshal(job)
	return string(jobString)
}

// 获取更新时间界限 (如果更新时间小于界限，就去更新职位)
func getUpdateTime() int64 {

	// nTime := time.Now()
	// yesTime := nTime.AddDate(0, 0, -1).Unix()

	timeStr := time.Now().Format("2006-01-02")
	// fmt.Println("timeStr:", timeStr)
	today, _ := time.Parse("2006-01-02", timeStr)
	todayTime := today.Unix() - 8*3600
	// fmt.Println("timeNumber:", todayTime)
	return todayTime
}

//GetRank 获取排名
func GetRank(vote int, devote int, timestamp int64, level int) float64 {

	// 等级加成  积分*(1+等级%) + 等级
	vote = vote*(100+level)/100 + level

	// 赞成与否定差
	voteDiff := vote - devote

	//争议度(赞成/否定)
	var voteDispute float64
	if voteDiff != 0 {
		voteDispute = math.Abs(float64(voteDiff))
	} else {
		voteDispute = 1
	}

	// 项目开始时间 2017-06-01
	projectStartTime, _ := time.Parse("2006-01-02", "2017-06-01")
	fund := projectStartTime.Unix() - 8*3600
	survivalTime := timestamp - fund

	// 投票方向与时间造成的系数差
	var timeMagin int64
	if voteDiff > 0 {
		timeMagin = survivalTime / 45000
	} else if voteDiff < 0 {
		timeMagin = -1 * survivalTime / 45000
	} else {
		timeMagin = 0
	}

	vateMagin := math.Log10(voteDispute)

	//详细算法
	socre := vateMagin + float64(timeMagin)
	return socre
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
