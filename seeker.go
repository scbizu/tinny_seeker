package main

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/hu17889/go_spider/core/common/page"
	"github.com/hu17889/go_spider/core/common/request"
	"github.com/hu17889/go_spider/core/spider"
)

//Processor
type SeekerProcessor struct {
}

//Construct function
func NewSeeker() *SeekerProcessor {
	return &SeekerProcessor{}
}

//Process
func (object *SeekerProcessor) Process(p *page.Page) {
	// var urls []string
	if !p.IsSucc() {
		color.Red(p.Errormsg())
		return
	}
	query := p.GetHtmlParser()
	query.Find(".anony-nav-links ul li").Each(func(i int, s *goquery.Selection) {
		value, _ := s.Find("a").Attr("href")
		name := s.Find("a").Text()
		p.AddField(name, value)
	})
	//continue to get links
	// p.AddTargetRequest(urls, "html")

}

func (object *SeekerProcessor) Finish() {
	color.Yellow("store data to the db...")
}

func main() {
	sp := spider.NewSpider(NewSeeker(), "DoubanSeeker")
	//Douban Root diretory
	req := request.NewRequest("https://www.douban.com/", "html", "", "GET", "", nil, nil, nil, nil)
	pageItems := sp.GetByRequest(req)
	//pageItems := sp.Get("http://baike.baidu.com/view/1628025.htm?fromtitle=http&fromid=243074&type=syn", "html")

	// url := pageItems.GetRequest().GetUrl()
	fmt.Println("-----------------------------------Result---------------------------------")
	// fmt.Println("url\t:\t" + url)
	for name, value := range pageItems.GetAll() {
		color.Green(name + "\t:\t" + value + "\n")
	}

}
