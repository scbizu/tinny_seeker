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
var domain = "http://www.dirtytao.com/"

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
	query.Find("a").Each(func(i int, s *goquery.Selection) {
		value, _ := s.Attr("href")
		name := s.Text()
		urls = append(urls, value)
		p.AddField(name, value)
	})
	//过滤URLS
	var filteredUrl []string
	//add urls to scheduler

	if localFlag {
		//only crawl the local domain
		for i := 0; i < len(urls); i++ {
			if !strings.Contains(urls[i], domain) {
				filteredUrl = append(filteredUrl, urls[i])
				// color.Red("%s", urls[i])
			}
		}
		p.AddTargetRequests(filteredUrl, "html")
	} else {
		filteredUrl = urls
	}

	if layer-1 > 0 {
		p.AddTargetRequests(filteredUrl, "html")
		layer--
	}

}

func (object *SeekerProcessor) Finish() {
	color.Yellow("store data to the db...")
}

func main() {

	sp := spider.NewSpider(NewSeeker(), "DoubanSeeker").AddUrl(domain, "html")
	//domain's ROOT
	req := request.NewRequest(domain, "html", "", "GET", "", nil, nil, nil, nil)

	//Run Client
	cli.Run(new(argT), func(ctx *cli.Context) error {

		argv := ctx.Argv().(*argT)
		layer = argv.Layer
		localFlag = argv.Zone

		// layer = argv.Layer
		pageItems := sp.SetThreadnum(4).SetScheduler(scheduler.NewQueueScheduler(true)).GetByRequest(req)
		//	sp.SetThreadnum(4).SetScheduler(scheduler.NewQueueScheduler(true)).AddPipeline(pipeline.NewPipelineConsole()).Run()
		// color.Red("%d", layer)

		fmt.Println("-----------------------------------Result---------------------------------")
		for name, value := range pageItems.GetAll() {
			color.Green(name + "\t:\t" + value + "\n")
		}

		return nil
	}, "cli for the tinny_seeker")
}
