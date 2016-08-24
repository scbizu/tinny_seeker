package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/garyburd/redigo/redis"
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

const domain = "zafu.edu.cn"
const targetURL = "http://www.zafu.edu.cn/"
const regularHref = `href="(.*?)"`
const regularAction = `action="(.*?)"`
const regularDataAction = `data-action="(.*?)"`

//crawl layer
var layer int
var urls []string

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

	// IDEA: regexp solution

	body := p.GetBodyStr()
	//href
	patternHref := regexp.MustCompile(regularHref)
	matchURLs := patternHref.FindAllStringSubmatch(body, -1)
	for i := 0; i < len(matchURLs); i++ {
		//handle the single URL
		value := matchURLs[i][1]
		value = RemoveSlashFilter(value)
		urls = append(urls, value)
		//ATOMIC OP
		// lock := &sync.Mutex{}
		// lock.Lock()
		// err := SQLmapcall(value)
		// lock.Unlock()
		// if err != nil {
		// 	color.Red("%s", err)
		// }
	}

	//action
	patternAction := regexp.MustCompile(regularAction)
	matchAction := patternAction.FindAllStringSubmatch(body, -1)

	patternDataAction := regexp.MustCompile(regularDataAction)
	matchDataAction := patternDataAction.FindAllStringSubmatch(body, -1)
	//data-action
	for _, v := range matchDataAction {
		matchAction = append(matchAction, v)
	}

	for i := 0; i < len(matchAction); i++ {
		//handle the single URL
		value := matchAction[i][1]
		//ATOMIC OP
		lock := &sync.Mutex{}
		value = AutofillURL(targetURL, value)

		//redis  GET

		conn, err := redis.Dial("tcp", ":6379")
		if err != nil {
			color.Red("%s", err)
		}

		data, err := conn.Do("GET", value)
		if err != nil {
			color.Red("%s", err)
		}
		if data == nil {
			color.Blue("待测URL:" + value)
			lock.Lock()
			err := SQLmapcall(value)
			lock.Unlock()
			if err != nil {
				color.Red("%s", err)
			}
		}

		// lock.Lock()
		// err := SQLmapcall(value)
		// lock.Unlock()
		// if err != nil {
		// 	color.Red("%s", err)
		// }

	}

	/*

		// IDEA: goquery	 solution
			query := p.GetHtmlParser()

			//CASE:<a href="" />
			query.Find("a").Each(func(i int, s *goquery.Selection) {
				value, _ := s.Attr("href")
				value = RemoveSlashFilter(value)
				// name := s.Text()
				urls = append(urls, value)

				// p.AddField(name, value)
				// err := SQLmapcall(value)
				// if err != nil {
				// 	color.Red("%s", err)
				// }
			})

		//CASE:<form action=""></form>
		query.Find("form").Each(func(i int, s *goquery.Selection) {
			lock := &sync.Mutex{}
			value, _ := s.Attr("action")
			if value == "" {
				value, _ = s.Attr("data-action")
			}

			value = AutofillURL(targetURL, value)

			color.Blue("待测URL:" + value)
			lock.Lock()
			err := SQLmapcall(value)
			lock.Unlock()
			if err != nil {
				color.Red("%s", err)
			}

		})
	*/
	//URLS  Filter
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

//SQLmapcall call on the sqlmap srcipt to connect with the sqlmap api
func SQLmapcall(URL string) error {
	//sqlmap middleware

	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		return err
	}

	taskerid, err := scanner.NewTasker()
	if err != nil {
		return err
	}

	isStart, err := scanner.StartTasker(taskerid, URL)
	if err != nil {
		return err
	}

	if isStart {
		color.Green("一个新的Tasker开启了....")
	}

	Res, err := scanner.GetResultFromTasker(taskerid)
	if err != nil {
		return err
	}
	if len(Res) > 0 {
		for _, v := range Res {
			color.Green(v + " ")
		}

	} else {
		color.Cyan("并没有什么卵的漏洞....")
	}
	//store into redis
	_, err = conn.Do("SET", URL, "DUPLICATE")
	if err != nil {
		return err
	}

	return nil
}

//AutofillURL is designed for the need that change relative path to absolute path
func AutofillURL(targetURL string, originURL string) string {
	originURL = RemoveSlashFilter(originURL)
	if !strings.Contains(originURL, "http") {
		originURL = targetURL + originURL
	}
	return originURL
}

//RemoveSlashFilter remove the slash in front of the url
func RemoveSlashFilter(targetURL string) string {
	var b bytes.Buffer
	URLbytes := []byte(targetURL)

	for k, v := range URLbytes {
		if k != 0 || v != '/' {
			b.WriteByte(v)
		}
	}
	return b.String()
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
