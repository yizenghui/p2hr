package main

import (
	"net/http"
	"strconv"
	"time"

	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/labstack/echo"
	"github.com/lib/pq"
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

// ResJob 响应结构
type ResJob struct {
	ID          uint      `json:"id" `
	CreatedAt   time.Time `json:"created_at" `
	UpdatedAt   time.Time `json:"updated_at" `
	Title       string    `json:"title" `
	Position    string    `json:"position" `
	Company     string    `json:"company" `
	Category    string    `json:"category" `
	Area        string    `json:"area" `
	Salary      string    `json:"salary" `
	Education   string    `json:"education" `
	Experience  string    `json:"experience" `
	Intro       string    `json:"intro"` // 职位介绍
	Tags        []string  `json:"tags"`  // 职位介绍
	Linkman     string    `json:"linkman" `
	Telephone   string    `json:"telephone" `
	Email       string    `json:"email" `
	Address     string    `json:"address" `
	UpdatedDate string    `json:"updated_date" `
}

func job(c echo.Context) error {
	var response ResJob
	// db, err := gorm.Open("postgres", "host=localhost user=postgres dbname=spider sslmode=disable password=123456")
	db, err := gorm.Open("postgres", "host=192.157.192.118 user=xiaoyi dbname=spider sslmode=disable password=123456")
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
	// fmt.Println(job.Tags)
	response.Tags = job.Tags
	// response.Tags = strings.Split(Substr(job.Tags, 1, -1), ",")
	response.Linkman = job.Linkman
	response.Telephone = job.Telephone
	response.Email = job.Email
	response.Address = job.Address
	response.UpdatedDate = GetUpdateTime(job.UpdatedAt)
	return c.JSON(http.StatusOK, response)
}

func jobs(c echo.Context) error {

	// 检查用户坐标
	// lat, lng := 23.1156237010336, 113.412643600147

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	// 坐标查询
	lat, _ := strconv.ParseFloat(c.QueryParam("lat"), 64)
	lng, _ := strconv.ParseFloat(c.QueryParam("lng"), 64)
	distance, _ := strconv.Atoi(c.QueryParam("distance"))
	category, _ := strconv.Atoi(c.QueryParam("category"))
	tag, _ := strconv.Atoi(c.QueryParam("tag"))
	minPay, _ := strconv.Atoi(c.QueryParam("min_pay"))
	maxPay, _ := strconv.Atoi(c.QueryParam("max_pay"))
	if distance <= 0 || distance > 20000 {
		distance = 5000
	}

	query := c.QueryParam("query")

	var mapWhere string
	if query != "" {

		mapWhere = fmt.Sprintf("earth_box (ll_to_earth (%f, %f),%d) @> ll_to_earth (lat, lng) and intro ~'%v'", lat, lng, distance, query)

	} else {

		mapWhere = fmt.Sprintf("earth_box (ll_to_earth (%f, %f),%d) @> ll_to_earth (lat, lng)", lat, lng, distance)

	}
	if category != 0 {
		mapWhere = fmt.Sprintf("%v and category = %d", mapWhere, category)
	}
	if tag != 0 {
		mapWhere = fmt.Sprintf("%v and param @> array[%d]", mapWhere, tag)
	}
	if minPay == -1 && maxPay == -1 {
		// 面议
		mapWhere = fmt.Sprintf("%v and min_pay = %d and max_pay=%d", mapWhere, 0, 0)
	} else if minPay != 0 && maxPay != 0 {
		// n1~n2 多少至多少
		mapWhere = fmt.Sprintf("%v and min_pay >= %d and min_pay<%d", mapWhere, minPay, maxPay)
	} else if minPay == 0 && maxPay != 0 {
		// 0~n1 多少以下不包括面议
		mapWhere = fmt.Sprintf("%v and min_pay >= %d and max_pay <= %d and max_pay > 0", mapWhere, minPay, maxPay)
	} else if minPay != 0 && maxPay == 0 {
		// n1~0 多少以上
		mapWhere = fmt.Sprintf("%v and min_pay >= %d", mapWhere, minPay)
	}

	// fmt.Println(mapWhere)

	if limit <= 0 || limit > 100 {
		limit = 10
	}
	// limit = 10
	if offset < 0 || offset > 500 {
		offset = 0
	}
	// offset = 0

	// fmt.Println(limit, offset)
	//db, err := gorm.Open("postgres", "host=localhost user=postgres dbname=spider sslmode=disable password=123456")
	db, err := gorm.Open("postgres", "host=192.157.192.118 user=xiaoyi dbname=spider sslmode=disable password=123456")

	if err != nil {
		panic("连接数据库失败")
	}

	// 自动迁移模式
	// db.AutoMigrate(&Job{})
	defer db.Close()

	var jobs Jobs
	// db.Offset(offset).Limit(limit).Where("earth_box (ll_to_earth (23.1156237010336, 113.412643600147),100) @> ll_to_earth (lat, lng)").Find(&jobs)
	db.Offset(offset).Limit(limit).Where(mapWhere).Order("rank desc").Find(&jobs)

	type ListJob struct {
		ID          uint     `json:"id" `
		Title       string   `json:"title" `
		Position    string   `json:"position" `
		Company     string   `json:"company" `
		Salary      string   `json:"salary" `
		Welfare     []string `json:"welfare" `
		Address     string   `json:"address" `
		Distance    string   `json:"distance" `
		Area        string   `json:"area" `
		Education   string   `json:"education" `
		Experience  string   `json:"experience" `
		ShowInfo    bool     `json:"showinfo" `
		UpdatedDate string   `json:"updated_date" `
	}
	type ListJobs []ListJob

	var list ListJobs
	for _, j := range jobs {
		// tags := strings.Split(Substr(j.Tags, 1, -1), ",")
		edu := conf.Education[j.Education]
		if edu == "" {
			edu = "学历不限"
		}
		exp := conf.Experience[j.Experience]
		if exp == "" {
			exp = "经验不限"
		} else {
			exp = exp + "经验"
		}
		lj := ListJob{
			ID:          j.ID,
			Title:       j.Title,
			Company:     j.Company,
			Address:     j.Address,
			Salary:      GetSalary(j.MinPay, j.MaxPay),
			Welfare:     j.Tags,
			Distance:    GetDistace(lat, lng, j.Lat, j.Lng),
			Area:        location.GetNameByAdcode(strconv.Itoa(j.Area)),
			Education:   edu,
			Experience:  exp,
			ShowInfo:    false,
			Position:    j.Position,
			UpdatedDate: GetUpdateTime(j.UpdatedAt),
		}
		list = append(list, lj)
	}
	return c.JSON(http.StatusOK, list)
}

// GetUpdateTime 格式化时间
func GetUpdateTime(updateAt time.Time) string {
	// timestamp :=updateAt.Unix()
	str := updateAt.Format("2006-01-02")
	today := time.Now().Format("2006-01-02")
	if str == today {
		str = "今天"
	}
	return str
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
			distance = "0.1km"
		case d < 500:
			distance = "0.5km"
		case d < 1000:
			distance = "1km"
		case d < 2000:
			distance = "2km"
		case d < 3000:
			distance = "3km"
		case d < 4000:
			distance = "4km"
		case d < 5000:
			distance = "5km"
		case d < 10000:
			distance = "10km"
		default:
			distance = "10km+"
		}
	} else {
		distance = ""
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
	//db, err = gorm.Open("postgres", "host=localhost user=postgres dbname=spider sslmode=disable password=123456")

	db, err = gorm.Open("postgres", "host=192.157.192.118 user=xiaoyi dbname=spider sslmode=disable password=123456")
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
