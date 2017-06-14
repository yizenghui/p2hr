package serve_test

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/astaxie/beego/validation"
	"github.com/hprose/hprose-golang/rpc"
)

func Test_Server(t *testing.T) {

	type Stub struct {
		Save      func(string) (string, error)
		AsyncSave func(func(string, error), string) `name:"save"`
	}

	type Job struct {
		Title      string   `json:"title"`
		Position   string   `json:"position"`
		Category   int      `json:"category"`
		Area       int      `json:"area"`
		Education  int      `json:"education"`
		Experience int      `json:"experience"`
		MinPay     int      `json:"min_pay"`
		MaxPay     int      `json:"max_pay"`
		Welfare    []int    `json:"welfare"`
		Tags       []string `json:"tags"`
		Company    string   `json:"company"`
		SourceFrom string   `json:"source_from"`
		CompanyUrl string   `json:"company_url"`
		LinkMan    string   `json:"link_man"`
		Telephone  string   `json:"telephone"`
		Email      string   `json:"email"`
	}

	client := rpc.NewClient("http://127.0.0.1:818/")
	var stub *Stub
	client.UseService(&stub)
	// stub.AsyncSave(func(result string, err error) {
	// 	fmt.Println(result, err)
	// }, `{"title":"异步标题"}`)

	var job Job
	job.Title = "职位的标题"
	job.Position = "PHP程序员"
	job.Category = 101
	job.Area = 454000

	welfare := []int{
		402,
		403,
	}

	job.Welfare = welfare
	job.Tags = []string{"包吃", "包住"}

	// var json_str []string

	// json.Unmarshal(job, &json_str)

	// fmt.Println(stub.Save(json_str))

	if json_str, err := json.Marshal(job); err == nil {
		// fmt.Println("================struct 到json str==")
		// fmt.Println(string(json_str))
		fmt.Println(stub.Save(string(json_str)))

	}

	// fmt.Println(stub.Save(`{"title":"同步标题","position":"同步标题","category":123,"area":225,"LinkMan":"易增辉","telephone":"13620455072","email":"245561237@qq.com","welfare":["好的","有前途的"]}`))
}

func Test_Job(t *testing.T) {
	type RequstJob struct {
		Title    string `json:"title" valid:"Required; MaxSize(12);"`
		Position string `json:"position" valid:"Required; MaxSize(12);"`
		Category int    `json:"category"  valid:"Range(101,129);"`
		Welfare  []int  `json:"welfare" ` // 暂时不清楚怎么使用子集验证
		// Welfare  []int  `json:"welfare"  valid:"Range(400,411);"`
	}

	job := RequstJob{}
	job.Title = "职位的标题"
	job.Position = "PHP程序员"
	job.Category = 101

	welfare := []int{
		402,
		403,
	}

	job.Welfare = welfare

	valid := validation.Validation{}

	b, err := valid.Valid(&job)
	if err != nil {
		// handle error
	}
	if !b {
		// validation does not pass
		// blabla...
		for _, err := range valid.Errors {
			log.Println(err.Key, err.Message)
		}
	}
}
