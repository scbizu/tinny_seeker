package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/hu17889/go_spider/core/common/page"
	"github.com/hu17889/go_spider/core/common/request"
	"github.com/hu17889/go_spider/core/scheduler"
	"github.com/hu17889/go_spider/core/spider"
	"github.com/mkideal/cli"
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
var domain = "dirtytao.com"
var targetURL = "http://www.dirtytao.com/"

//  seeker  --la 	 default is OFF
var localFlag bool

//Processor
type SeekerProcessor struct {
}

//Construct function
func NewSeeker() *SeekerProcessor {
	return &SeekerProcessor{}
}

//Process
func (object *SeekerProcessor) Process(p *page.Page) {

	if !p.IsSucc() {
		color.Red(p.Errormsg())
		return
	}
	query := p.GetHtmlParser()
	//CASE:<a href="" />
	query.Find("a").Each(func(i int, s *goquery.Selection) {
		value, _ := s.Attr("href")
		//	name := s.Text()
		urls = append(urls, value)
		req := p.GetRequest()
		//p.AddField(name, value)
		//	resp, _ := downloader.ConnectByHttp(p, req)
		//	p.AddField(value, resp.Status)
		p.AddField(value, req.GetMethod())
	})
	//CASE:<form action=""></form>
	query.Find("form").Each(func(i int, s *goquery.Selection) {
		value, _ := s.Attr("action")
		value = AutofillUrl(targetURL, value)
		method, _ := s.Attr("method")
		urls = append(urls, value)
		p.AddField(value, method)
	})
	//过滤URLS
	var filteredUrl []string
	//add urls to scheduler

	if localFlag {
		//only crawl the local domain
		for i := 0; i < len(urls); i++ {
			urls[i] = AutofillUrl(targetURL, urls[i])
			if strings.Contains(urls[i], domain) && !strings.Contains(urls[i], "javascript") {
				filteredUrl = append(filteredUrl, urls[i])
				// color.Red("%s", urls[i])
			}
		}
		//	p.AddTargetRequests(filteredUrl, "html")
	} else {
		for i := 0; i < len(urls); i++ {
			urls[i] = AutofillUrl(targetURL, urls[i])
			if !strings.Contains(urls[i], "javascript") {
				filteredUrl = append(filteredUrl, urls[i])
			}
		}
		//	p.AddTargetRequests(filteredUrl, "html")
	}

	if layer-1 > 0 {
		p.AddTargetRequests(filteredUrl, "html")
		layer--
	}

}

func (object *SeekerProcessor) Finish() {
	color.Yellow("store data to the db...")
}

//design for the need that change relative path to absolute path
func AutofillUrl(targetUrl string, originUrl string) string {
	if !strings.Contains(originUrl, "http") {
		originUrl = targetUrl + originUrl
	}
	return originUrl
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
		pageItems := sp.SetThreadnum(1).SetScheduler(scheduler.NewQueueScheduler(true)).SetSleepTime("rand", 800, 1500).GetByRequest(req)
		//	sp.SetThreadnum(4).SetScheduler(scheduler.NewQueueScheduler(true)).AddPipeline(pipeline.NewPipelineConsole()).Run()
		// color.Red("%d", layer)

		fmt.Println("-----------------------------------Result---------------------------------")
		for name, value := range pageItems.GetAll() {
			color.Green(name + ":" + value + "\n")
		}

		return nil
	}, "cli for the tinny_seeker")
}
