package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/astaxie/beego/validation"
	"github.com/hprose/hprose-golang/rpc"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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
		Title      string  `gorm:"size:255"`   // 职位标题
		Position   string  `gorm:"size:255"`   // 原职位分类
		Company    string  `gorm:"size:255"`   // 公司名
		Param      string  `gorm:"type:int[]"` // 标签
		Category   int     // 分类
		Area       int     // 地区
		MinPay     int     // 最小月薪
		MaxPay     int     // 最大月薪
		Education  int     // 学历
		Experience int     // 工作经验
		Intro      string  `gorm:"type:text"` // 职位介绍
		Rank       float32 // 排序
		Tags       string  `gorm:"type:text[]"` // 标签
		SourceFrom string  `gorm:"size:255"`    // string默认长度为255, 使用这种tag重设。
		CompanyURL string  `gorm:"size:255"`    // string默认长度为255, 使用这种tag重设。
		Linkman    string  `gorm:"size:255"`
		Telephone  string  `gorm:"size:255"`
		Email      string  `gorm:"size:255"`
		Lng        float64
		Lat        float64
		Address    string
	}
	// Jobs struct{
	// 	job[] *Job
	// }
)

// Category 大类
var Category = map[int]string{
	101: "经营管理类",
	102: "公关/市场营销类",
	103: "贸易/销售/业务类",
	104: "财务类",
	105: "行政/人力资源管理类",
	106: "文职类",
	107: "客户服务类",
	108: "工厂类",
	109: "计算机/互联网类",
	110: "电子/通讯类",
	111: "机械类",
	112: "规划/建筑/建材类",
	113: "房地产/物业管理类",
	114: "金融/经济",
	115: "设计类",
	116: "法律类",
	117: "酒店/餐饮类",
	118: "物流/交通运输类",
	119: "商场类",
	120: "电气/电力类",
	121: "咨询/顾问类",
	122: "化工/生物类",
	123: "文化/教育/体育/艺术类",
	124: "医疗卫生/护理/保健类",
	125: "新闻/出版/传媒类",
	126: "公众服务类",
	127: "印刷/染织类",
	128: "技工类",
	129: "其他专业类",
}

//Welfare 标签
var Welfare = map[int]string{
	401: "五险一金",
	402: "包住",
	403: "包吃",
	404: "年底双薪",
	405: "周末双休",
	406: "交通补助",
	407: "加班补助",
	408: "饭补",
	409: "话补",
	410: "房补",
}

// 数据库对象

var db *gorm.DB

func init() {
	db, _ = gorm.Open("postgres", "host=localhost user=postgres dbname=spider password=123456 sslmode=disable")

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

	db.Where(Job{SourceFrom: job.SourceFrom}).FirstOrCreate(&job)
	err = RequstJobSaveData(&job, j)
	if err != nil {
		return "err: " + err.Error() + "!"
	}

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
	param := []string{}
	// var param map[int]int

	var b bytes.Buffer
	b.WriteString("{")
	if rj.Category != 0 {
		param = append(param, strconv.Itoa(rj.Category))
	}
	// if err := checkCategory(rj.Category); err != nil {

	// 	return job, err
	// }

	if rj.Area != 0 {
		param = append(param, strconv.Itoa(rj.Area))
	}
	if rj.Education != 0 {
		param = append(param, strconv.Itoa(rj.Education))
	}
	if rj.Experience != 0 {
		param = append(param, strconv.Itoa(rj.Experience))
	}

	for _, tag := range rj.Welfare {
		param = append(param, strconv.Itoa(tag))
	}

	b.WriteString(strings.Join(param, ","))

	b.WriteString("}")

	job.Param = b.String()
	job.Tags = "{" + strings.Join(rj.Tags, ",") + "}"

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
