package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/hu17889/go_spider/core/common/page"
	"github.com/hu17889/go_spider/core/common/request"
	"github.com/hu17889/go_spider/core/scheduler"
	"github.com/hu17889/go_spider/core/spider"
	"github.com/mkideal/cli"
	"github.com/scbizu/tinny_seeker/scanner"
)

type argT struct {
	cli.Helper
	Layer int  `cli:"la" usage:"crawl layer ,a int type,default is 1" dft:"1"`
	Zone  bool `cli:"local" usage:"use this command to crawl the URL under the same domain" dft:"false"`
	// stroe bool `cli:"!stroe" usage:"use this command to stroe the url "`
}

//crawl layer
var layer int
var urls []string
var domain = "nxzx.net"
var targetURL = "http://www.nxzx.net/"

//  seeker  --la 	 default is OFF
var localFlag bool

//SeekerProcessor Construction
type SeekerProcessor struct {
}

//NewSeeker returns a Processor
func NewSeeker() *SeekerProcessor {
	return &SeekerProcessor{}
}

//Process is used to analyse the whole HTML
func (object *SeekerProcessor) Process(p *page.Page) {

	if !p.IsSucc() {
		color.Red(p.Errormsg())
		return
	}
	query := p.GetHtmlParser()
	//CASE:<a href="" />
	query.Find("a").Each(func(i int, s *goquery.Selection) {
		value, _ := s.Attr("href")
		// name := s.Text()
		urls = append(urls, value)

		// p.AddField(name, value)

	})
	//CASE:<form action=""></form>
	query.Find("form").Each(func(i int, s *goquery.Selection) {

		value, _ := s.Attr("action")
		value = AutofillURL(targetURL, value)

		//TODO:sqlmap insert here

		taskerid, err := scanner.NewTasker()
		if err != nil {
			color.Red("%s", err)
		}

		isStart, err := scanner.StartTasker(taskerid, value)
		if err != nil {
			color.Red("%s", err)
		}

		if isStart {
			color.Green("一个新的Tasker开启了....")
		}

		Res, err := scanner.GetResultFromTasker(taskerid)
		if err != nil {
			color.Red("%s", err)
		}
		if len(Res) > 0 {
			for _, v := range Res {
				color.Green(v + " ")
			}
		} else {
			color.Cyan("并没有什么卵的漏洞....")
		}

		// method, _ := s.Attr("method")
		// // NOTE: SQL injection here
		// if method == "get" {
		// 	params := GetChildsWithTag("input", s)
		// 	AttackParams := "4%'and 1=2 and '%' ='"
		// 	ATKURL := GenerateATKURL(AttackParams, params)
		// 	color.Green(value + ATKURL)
		// 	//REQUEST
		// 	resp, err := http.Get(value + ATKURL)
		// 	resp1, err := http.Get(value + ATKURL)
		// 	if err != nil {
		// 		color.Red("%s", err)
		// 	}
		// 	if resp.StatusCode == 200 && resp1.StatusCode == 200 {
		// 		p.AddField(value+ATKURL, ": searchbar SQL injection test pass!")
		// 	} else {
		// 		p.AddField(value+ATKURL, resp.Status+"\t"+resp1.Status)
		// 	}
		// }
		//add form object to scheduler
		urls = append(urls, value)
	})

	//过滤URLS
	var filteredURL []string
	//add urls to scheduler

	if localFlag {
		//only crawl the local domain
		for i := 0; i < len(urls); i++ {
			urls[i] = AutofillURL(targetURL, urls[i])
			if strings.Contains(urls[i], domain) && !strings.Contains(urls[i], "javascript") {
				filteredURL = append(filteredURL, urls[i])
				// color.Red("%s", urls[i])
			}
		}
		//	p.AddTargetRequests(filteredUrl, "html")
	} else {
		for i := 0; i < len(urls); i++ {
			urls[i] = AutofillURL(targetURL, urls[i])
			if !strings.Contains(urls[i], "javascript") {
				filteredURL = append(filteredURL, urls[i])
			}
		}
		//	p.AddTargetRequests(filteredUrl, "html")
	}

	if layer-1 > 0 {
		p.AddTargetRequests(filteredURL, "html")
		layer--
	}

}

//Finish  will be called in the end
func (object *SeekerProcessor) Finish() {
	color.Yellow("store data to the db...")
}

//AutofillURL is designed for the need that change relative path to absolute path
func AutofillURL(targetURL string, originURL string) string {
	if !strings.Contains(originURL, "http") {
		originURL = targetURL + originURL
	}
	return originURL
}

//GetChildsWithTag Get Children tag with specific name
func GetChildsWithTag(tagname string, s *goquery.Selection) []string {
	var tags []string
	//Children numbers

	s.Find(tagname).Each(func(i int, child *goquery.Selection) {
		len := child.Children().Length()
		tag, _ := child.Attr("name")
		tags = append(tags, tag)
		if len != 0 {
			GetChildsWithTag(tagname, child.Children())
		}
	})

	return tags
}

//GenerateATKURL generate the  ATK request URL(temp)
//TODO:need FIX
func GenerateATKURL(ATKdata string, urlparams []string) string {
	var buf bytes.Buffer
	buf.WriteString("?" + urlparams[0] + "=" + ATKdata)
	for i := 1; i < len(urlparams); i++ {
		buf.WriteString("&")
		buf.WriteString(urlparams[i] + "=" + ATKdata)
	}
	return buf.String()
}

func main() {

	sp := spider.NewSpider(NewSeeker(), "DoubanSeeker").AddUrl(domain, "html")
	//domain's ROOT
	req := request.NewRequest(targetURL, "html", "", "GET", "", nil, nil, nil, nil)

	//Run Client
	cli.Run(new(argT), func(ctx *cli.Context) error {

		argv := ctx.Argv().(*argT)
		layer = argv.Layer
		localFlag = argv.Zone

		// layer = argv.Layer
		pageItems := sp.SetThreadnum(2).SetScheduler(scheduler.NewQueueScheduler(true)).SetSleepTime("rand", 1500, 3500).GetByRequest(req)
		// sp.SetThreadnum(4).SetScheduler(scheduler.NewQueueScheduler(true)).AddPipeline(pipeline.NewPipelineConsole()).Run()
		//	color.Red("%d", layer)

		fmt.Println("-----------------------------------Result---------------------------------")
		for name, value := range pageItems.GetAll() {
			color.Green(name + ":" + value + "\n")
		}
		return nil
	}, "cli for the tinny_seeker")
}
