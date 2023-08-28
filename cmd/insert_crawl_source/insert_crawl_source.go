package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/db"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	source = flag.String("source", "GIF_source.csv", "source category Pinterest")
	conf   = flag.String("conf", "investing_economic_calender.toml", "config run file *.toml")
	c      = config.CrawlConfig{}
)

func UpdateCountryAndCategory() ([]db.Country, []db.EconomicCategory) {
	countries := []db.Country{}
	categories := []db.EconomicCategory{}
	req, err := http.NewRequest("GET", "https://www.investing.com/economic-calendar/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return countries, categories
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return countries, categories
	}
	doc.Find("ul[class=countryOption] li label[for*=country]").Each(
		func(i int, selection *goquery.Selection) {
			title := selection.Text()
			title = strings.TrimSpace(title)
			investingIdText, _ := selection.Attr("for")
			investingIdText = strings.ReplaceAll(investingIdText, "country", "")
			investingId,_ := strconv.Atoi(investingIdText)
			country:=db.Country{
				Title: title,
				InvestingId: investingId,
			}
			countries=append(countries,country)
		})

	doc.Find("div[id*=_category] label[for*=category]").Each(
		func(i int, selection *goquery.Selection) {
			title := selection.Text()
			title = strings.TrimSpace(title)
			valueQuery,_:=selection.Attr("for")
			valueQuery = strings.ReplaceAll(valueQuery, "category", "")
			category:=db.EconomicCategory{
				Title: title,
				ValueQuery:valueQuery,
			}
			categories=append(categories,category)
		})
	return countries, categories
}
func main() {
	flag.Parse()
	configBytes, err := ioutil.ReadFile(*conf)
	if err != nil {
		fmt.Println("err when read config file ", err, "file ", *conf)
	}
	err = toml.Unmarshal(configBytes, &c)
	if err != nil {
		fmt.Println("err when pass toml file ", err)
	}
	text, err := json.Marshal(c)
	fmt.Println("Success read config from toml file ", string(text))
	err = db.Init(c.Postgres)
	if err != nil {
		fmt.Println("err", err)
	}

	// Open CSV file
	fileContent, err := os.Open(*source)
	if err != nil {
		fmt.Println("err", err)
	}
	defer fileContent.Close()

	// Read File into a Variable
	mysql := db.GetDb()

	countries,categories:=UpdateCountryAndCategory()
	for _, country := range countries {
		_,err=mysql.Model(&country).Insert()
		if err !=nil{
			fmt.Println("err when insert countries",err)
		}
	}

	for _, category := range categories {
		mysql.Model(&category).Insert()
		if err !=nil{
			fmt.Println("err when insert category",err)
		}
	}
}
