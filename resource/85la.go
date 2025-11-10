package resource

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/injoyai/logs"
	"net/http"
)

func Download85la() ([]*Node, error) {

	req, err := http.NewRequest("GET", "https://www.85la.com", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", "www.85la.com")
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")

	//打开85la网站
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	logs.Debug(doc.Text())

	url := ""
	doc.Find("div").EachWithBreak(func(i int, selection *goquery.Selection) bool {
		logs.Debug(selection.Text())
		//href, exist := selection.Attr("href")
		//if exist {
		//	url = href
		//	return false
		//}
		return true
	})

	logs.Debug("url", url)

	return nil, nil
}

type Node struct {
	Protocol string
	User     string
	Host     string
	Port     string
	Params   map[string]string
	Tag      string
}
