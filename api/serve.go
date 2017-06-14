package main

import (
	"net/http"
	"strconv"
	"time"

	"strings"

	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/labstack/echo"
	"github.com/xuebing1110/location"
	"github.com/yizenghui/gps"

	"github.com/yizenghui/spider/conf"
)

var db *gorm.DB

func api(c echo.Context) error {
	return c.JSON(http.StatusOK, "ok!")
}

// Jobs jobs
type Jobs []Job

// Job job
type Job struct {
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

// ResJob 响应结构
type ResJob struct {
	ID         uint      `json:"id" `
	CreatedAt  time.Time `json:"created_at" `
	UpdatedAt  time.Time `json:"updated_at" `
	Title      string    `json:"title" `
	Position   string    `json:"position" `
	Company    string    `json:"company" `
	Category   string    `json:"category" `
	Area       string    `json:"area" `
	Salary     string    `json:"salary" `
	Education  string    `json:"education" `
	Experience string    `json:"experience" `
	Intro      string    `json:"intro"` // 职位介绍
	Tags       []string  `json:"tags"`  // 职位介绍
	Linkman    string    `json:"linkman" `
	Telephone  string    `json:"telephone" `
	Email      string    `json:"email" `
	Address    string    `json:"address" `
}

func job(c echo.Context) error {
	var response ResJob
	db, err := gorm.Open("postgres", "host=localhost user=postgres dbname=spider sslmode=disable password=123456")
	if err != nil {
		panic("连接数据库失败")
	}
	// 自动迁移模式
	// db.AutoMigrate(&Job{})
	defer db.Close()

	var job Job
	id, _ := strconv.Atoi(c.Param("id"))
	db.First(&job, id)
	response.ID = job.ID
	response.CreatedAt = job.CreatedAt
	response.UpdatedAt = job.UpdatedAt
	response.Title = job.Title
	response.Position = job.Position
	response.Company = job.Company
	response.Category = conf.Category[job.Category]
	response.Area = location.GetNameByAdcode(strconv.Itoa(job.Area))
	response.Salary = GetSalary(job.MinPay, job.MaxPay)
	response.Education = conf.Education[job.Education]
	response.Experience = conf.Experience[job.Experience]
	response.Intro = job.Intro
	response.Tags = strings.Split(Substr(job.Tags, 1, -1), ",")
	response.Linkman = job.Linkman
	response.Linkman = job.Linkman
	response.Telephone = job.Telephone
	response.Email = job.Email
	response.Address = job.Address
	return c.JSON(http.StatusOK, response)
}

func jobs(c echo.Context) error {

	// 检查用户坐标
	lat, lng := 23.1156237010336, 113.412643600147

	db, err := gorm.Open("postgres", "host=localhost user=postgres dbname=spider sslmode=disable password=123456")

	if err != nil {
		panic("连接数据库失败")
	}

	// 自动迁移模式
	// db.AutoMigrate(&Job{})
	defer db.Close()

	var jobs Jobs
	db.Find(&jobs)

	type ListJob struct {
		ID       uint     `json:"id" `
		Title    string   `json:"title" `
		Company  string   `json:"company" `
		Salary   string   `json:"salary" `
		Welfare  []string `json:"welfare" `
		Address  string   `json:"address" `
		Distance string   `json:"distance" `
	}
	type ListJobs []ListJob

	var list ListJobs
	for _, j := range jobs {
		tags := strings.Split(Substr(j.Tags, 1, -1), ",")
		lj := ListJob{
			ID:       j.ID,
			Title:    j.Title,
			Company:  j.Company,
			Address:  j.Address,
			Salary:   GetSalary(j.MinPay, j.MaxPay),
			Welfare:  tags,
			Distance: GetDistace(lat, lng, j.Lat, j.Lng),
		}
		list = append(list, lj)
	}
	return c.JSON(http.StatusOK, list)
}

// GetSalary 待遇文字
func GetSalary(MinPay, MaxPay int) string {
	if MinPay == 0 && MaxPay == 0 {
		return "面议"
	}
	if MinPay > 0 && MaxPay <= 25000 {
		return fmt.Sprintf("%d-%d", MinPay, MaxPay)
	}
	return fmt.Sprintf("%d以上", MaxPay)
}

//GetDistace 获取2点间的距离
func GetDistace(latA, lngA, LatB, LngB float64) string {
	var distance string
	if LatB > 0 && LngB > 0 {
		d := gps.Distance(latA, lngA, LatB, LngB)
		switch {
		case d < 100:
			distance = "百米内"
		case d < 500:
			distance = "半公里内"
		case d < 1000:
			distance = "1公里内"
		case d < 2000:
			distance = "2公里内"
		case d < 3000:
			distance = "3公里内"
		case d < 4000:
			distance = "4公里内"
		case d < 5000:
			distance = "5公里内"
		case d < 10000:
			distance = "10公里内"
		default:
			distance = "超过10公里"
		}
	} else {
		distance = "未知距离"
	}
	return distance
}

// Substr 截取字符串
func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}

	if length < 0 {
		length = rl + length - start
	}

	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}
func init() {

	var err error
	db, err = gorm.Open("postgres", "host=localhost user=postgres dbname=spider sslmode=disable password=123456")

	if err != nil {
		panic("连接数据库失败")
	}

	// 自动迁移模式
	// db.AutoMigrate(&Job{})
	defer db.Close()
}

func main() {
	e := echo.New()
	e.GET("/api", api)
	e.GET("/job/:id", job)
	e.GET("/job", jobs)
	e.GET("/jobs", jobs)
	e.Logger.Fatal(e.StartTLS(":1323", "cert.pem", "key.pem"))
}
