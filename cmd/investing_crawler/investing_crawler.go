package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/date"
	"go-pinterest/db"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	//SessionKey    = "TWc9PSZ5azdObjM3eEJJTVlJQUNjTStlSzU3ajhGWC8veUI3T1VycUVrc3Y4dmkzVlgvei9ZTW91T1JiTzJycTZHMHRJVjBDZ1hCSWMwSERSREpFUkdvNy9wQStZRGF2VktkVVhKN2V1LzU1WGpKa0hpYzFidG96UkdZQXdITWFDenZJVVZ2dVl3aVBHT1IyK3NiZnJJR0hWb1plRER5Z2hPM1RHc1BhWWVhdUlLbjg1MlkzTnVDTTVZNjFtbjBYQndQVmF0THhWVk85b29hMFV2OUx6NVRFMkFpZlAyZStaU0dCQUovQWVtQTVuR2xDNm9NaU0yL3RTdXV5aGhXVXprN0dOOU1EU3lVcy9VMUk2SVpzK0FHYUE5TExLZGFUVmY5aEdJQVJmcjJVUnBlc3ZhVENxRzFoY3A1UHFOSDlIYXdZM1RpQ2sxb2FKdXJTZ2pGMlRvNzdUZFFOK2ZlTXZVdGx0cnFFUWpuSzVtTU1GTzhWejZKcWw3all0VlM1aVRQcFdZdHpQK0Q4MGdGZUhEMnN0d0xURThOY2gwMUt6QXpDRVZJcllSQXBVZmhKMzROUTUyU21EclE1VE5Lb1E3TVlPSU9zZE5xRlNIcnJ3eXhWc2xIbTBvcnBnSHlLMlBYQnEwYm5kQSt4dXVMK0VTTmVqMWpGQThzUjdiTDdNU1lXVFhWaVpxQ3hEK3E0R2ZWZlo3ek1selBWbDJiN3VnMlJrRE1DTXRYeXJDeGhWK0dqeGJqRWVWNTdBNW9ZcWkvdmRQU0NNMmszZlNKaGtBSG1icFExdkF6Mjh4TDFydXFyNk5GZmZTSTkyaVhYaE1XeTZ6a1JDaEFzdlViTWVyajRpcWo5Z2xKOTBBOU9HNEdxVGw2MGR5STlFdGVXYnRtaVJ6ZTY2OGdqNVNiMkhpUHZrL0VldnlnMlVhV0pKQkZnenE1Q3RXMkI2Qi9WTVUxK2VTdUoyam1hWVB2Um9Obyt0S2cvQUxrM3Y4dWxaMHJmMmdLVlJMLytHUWFYNVoxWElqSk1kRnlnOEpHQ0JscVZWN0FjNHJ1aDBTK1JlenRUODdVSUx1RmtZOERlUEFaYllETld4eit0ZU8vL3RGcFpST0lSYWxKcHJTSFI5WmRab1hRR2JHM0c4dHNmdTlmVFk0Y2FVeDkwbm9xWk1KSTNDb28rWTMxSjVaZnZCN0kxdWIzWGEwVCswZDVwV3BIVnh4eEpyRUdRVEovT3dTRW5WdW05U20vd1hTTCtld0piRERRZGN0ckRoa0NKS212K1gmeHlYTkcvdFJNZmtBNzJDTUQ1SUt3cmJFUFlFPQ=="
	//Csrftoken     = "1d7df6fbd5e80a9b140abd4196ccf0d6"
	//maxRelated    = flag.Int("max_related_pin", 50, "max items per query related")
	conf          = flag.String("conf", "./investing_economic_calender.toml", "config run file *.toml")
	c             = config.CrawlConfig{}
	countriesName = map[string]int{}
)

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
		fmt.Println("err when connect to db", err)
	}
	defer db.Close()
	categories, _ := db.GetAllCategories()
	countries, _ := db.GetAllCountries()
	countryIds := []int{}
	for _, country := range countries {
		countryIds = append(countryIds, country.InvestingId)
		countriesName[country.Title] = country.Id
	}
	now := time.Now().Unix()
	now = (now / date.SECOND_PER_DAY) * date.SECOND_PER_DAY
	for i := 0; i < 4000; i++ {
		crawlDate := time.Unix(now, 0)
		dateTime := date.FormatDate(crawlDate)
		fmt.Println(dateTime)
		for _, category := range categories {
			err = GetEventEconomic(dateTime, countryIds, category)
			if err != nil {
				fmt.Println("err when GetEventEconomic", err, category)
				return
			}
		}
		now = now - date.SECOND_PER_DAY
	}
}
func GetDataFromUrl(url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch bodyBytes: %d %s", resp.StatusCode, resp.Status)
	}
	return bodyBytes, nil
}

func GetEventEconomic(day string, countryIds []int, category db.EconomicCategory) error {
	bodyPost := "country%5B%5D=" + strconv.Itoa(countryIds[0])
	for i := 1; i < len(countryIds); i++ {
		bodyPost = bodyPost + "&country%5B%5D=" + strconv.Itoa(countryIds[i])
	}
	bodyPost = bodyPost + "&category%5B%5D=" +
		category.ValueQuery +
		"&importance%5B%5D=1&importance%5B%5D=2&importance%5B%5D=3" +
		"&dateFrom=" +
		day +
		"&dateTo=" +
		day +
		"&timeZone=55&timeFilter=timeOnly&currentTab=custom&submitFilters=1&limit_from=0"
	reponseByte, err := GetDataFromUrl("https://www.investing.com/economic-calendar/Service/getCalendarFilteredData", []byte(bodyPost))
	if err != nil {
		return err
	}
	response := map[string]interface{}{}
	err = json.Unmarshal(reponseByte, &response)
	if err != nil {
		return err
	}
	data := "<html><body><table>" + response["data"].(string) + "</table></body></html>"
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return err
	}
	count := 0
	doc.Find("tr[id*=eventRowId]").Each(
		func(i int, tr *goquery.Selection) {
			count++
			eventInfo := db.EventInfo{}
			country, _ := tr.Find("td[class*=flagCur] span[title]").Attr("title")
			countryId := countriesName[country]
			cur := tr.Find("td[class*=flagCur]").Text()
			cur = strings.TrimSpace(cur)
			if len(cur) > 0 {
				_, err = db.UpdateCurrency(countryId, cur)
				if err != nil {
					fmt.Println("err when UpdateCurrency ", countryId, cur)
				}
			}
			firstLeft := tr.Find("td[class=\"first left\"]").Text()
			firstLeft = strings.ToLower(firstLeft)
			textNum := tr.Find("td[class*=\"textNum\"]").Text()
			textNum = strings.ToLower(textNum)
			if strings.Contains(firstLeft, "all day") || strings.Contains(textNum, "holiday") {
				title := tr.Find("td[class*=\"left event\"]").Text()
				time := date.ParseDate(day + " 0:0:0")
				holiday := db.Holiday{Title: title, CountryId: countryId, Time: time}
				if strings.Contains(firstLeft, "all day") {
					holiday.AllDay = true
				} else {
					holiday.AllDay = false
				}
				db.InsertHoliday(holiday)
				return
			}
			investingIdText, _ := tr.Attr("id")
			investingIdText = strings.ReplaceAll(investingIdText, "eventRowId_", "")
			investingId, _ := strconv.Atoi(investingIdText)
			eventInfo, _ = db.GetEventIdByInvestingId(investingId)
			importance := tr.Find("td[title*=Expected] i[class=grayFullBullishIcon]").Size()
			if eventInfo.Id == 0 {
				eventTime, _ := tr.Attr("data-event-datetime")
				timeStamp := date.ParseDate(eventTime)
				titleElement := tr.Find("td[class*=\"left event\"] a[href]")
				title := titleElement.Text()
				url, _ := titleElement.Attr("href")
				actual := tr.Find("td[id*=eventActual]").Text()
				forecast := tr.Find("td[id*=eventForecast]").Text()
				previous := tr.Find("td[id*=eventPrevious]").Text()
				eventInfo.Actual = actual
				eventInfo.ForeCast = forecast
				eventInfo.Title = title
				eventInfo.Previous = previous
				eventInfo.Importance = importance
				eventInfo.Url = "https://www.investing.com/" + url
				eventInfo.Time = timeStamp
				eventInfo.InvestingId = investingId
				eventInfo.CountryId = countryId
				eventInfo, err = db.InsertEventInfo(eventInfo)
				if err != nil {
					fmt.Println("err when insert event info", err, firstLeft)
					fmt.Println(tr.Html())
					return
				}
			}
			if eventInfo.Id > 0 {
				_, err = db.InsertEventToList(db.EventList{EventId: eventInfo.Id, CountryId: countryId, CategoryId: category.Id, Importance: importance, EventTime: eventInfo.Time})
				if err != nil {
					fmt.Println("err when insert event list", err)
					return
				}
			}
		})
	if count > 50 {
		fmt.Println(count)
		fmt.Println(day, category.Title)
	}
	return nil
}

//curl 'https://www.investing.com/economic-calendar/Service/getCalendarFilteredData' \
//-H 'content-type: application/x-www-form-urlencoded' \
//-H 'user-agent: Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Mobile Safari/537.36' \
//-H 'x-requested-with: XMLHttpRequest' \
//--data-raw 'country%5B%5D=75&country%5B%5D=84&country%5B%5D=178&country%5B%5D=138&country%5B%5D=168&country%5B%5D=180&country%5B%5D=5&country%5B%5D=4&country%5B%5D=143&country%5B%5D=61&country%5B%5D=123&country%5B%5D=63&country%5B%5D=202&country%5B%5D=41&country%5B%5D=85&country%5B%5D=46&country%5B%5D=12&country%5B%5D=9&country%5B%5D=162&country%5B%5D=26&country%5B%5D=11&country%5B%5D=110&country%5B%5D=112&country%5B%5D=90&country%5B%5D=36&country%5B%5D=238&country%5B%5D=52&country%5B%5D=80&country%5B%5D=56&country%5B%5D=100&country%5B%5D=170&country%5B%5D=38&country%5B%5D=53&country%5B%5D=45&country%5B%5D=125&country%5B%5D=148&country%5B%5D=193&country%5B%5D=44&country%5B%5D=87&country%5B%5D=60&country%5B%5D=20&country%5B%5D=43&country%5B%5D=21&country%5B%5D=172&country%5B%5D=82&country%5B%5D=105&country%5B%5D=247&country%5B%5D=139&country%5B%5D=7&country%5B%5D=188&country%5B%5D=109&country%5B%5D=42&country%5B%5D=111&country%5B%5D=103&country%5B%5D=96&country%5B%5D=68&country%5B%5D=97&country%5B%5D=204&country%5B%5D=94&country%5B%5D=57&country%5B%5D=102&country%5B%5D=92&country%5B%5D=35&country%5B%5D=119&country%5B%5D=10&country%5B%5D=23&country%5B%5D=33&country%5B%5D=66&country%5B%5D=48&country%5B%5D=14&country%5B%5D=106&country%5B%5D=93&country%5B%5D=39&country%5B%5D=51&country%5B%5D=74&country%5B%5D=17&country%5B%5D=22&country%5B%5D=71&country%5B%5D=72&country%5B%5D=89&country%5B%5D=59&country%5B%5D=121&country%5B%5D=24&country%5B%5D=55&country%5B%5D=107&country%5B%5D=113&country%5B%5D=78&country%5B%5D=15&country%5B%5D=122&country%5B%5D=37&country%5B%5D=27&country%5B%5D=232&country%5B%5D=6&country%5B%5D=70&country%5B%5D=32&country%5B%5D=163&country%5B%5D=174&country%5B%5D=8&country%5B%5D=34&country%5B%5D=47&country%5B%5D=145&country%5B%5D=114&country%5B%5D=54&country%5B%5D=25&country%5B%5D=29&country%5B%5D=86&country%5B%5D=95&category%5B%5D=_centralBanks&importance%5B%5D=1&importance%5B%5D=2&importance%5B%5D=3&dateFrom=2023-5-10&dateTo=2023-5-10&timeZone=55&timeFilter=timeOnly&currentTab=custom&submitFilters=1&limit_from=0
//' \
//--compressed
//
//country%5B%5D=75&country%5B%5D=84&country%5B%5D=178&country%5B%5D=138&country%5B%5D=168&country%5B%5D=180&country%5B%5D=5&country%5B%5D=4&country%5B%5D=143&country%5B%5D=61&country%5B%5D=123&country%5B%5D=63&country%5B%5D=202&country%5B%5D=41&country%5B%5D=85&country%5B%5D=46&country%5B%5D=12&country%5B%5D=9&country%5B%5D=162&country%5B%5D=26&country%5B%5D=11&country%5B%5D=110&country%5B%5D=112&country%5B%5D=90&country%5B%5D=36&country%5B%5D=238&country%5B%5D=52&country%5B%5D=80&country%5B%5D=56&country%5B%5D=100&country%5B%5D=170&country%5B%5D=38&country%5B%5D=53&country%5B%5D=45&country%5B%5D=125&country%5B%5D=148&country%5B%5D=193&country%5B%5D=44&country%5B%5D=87&country%5B%5D=60&country%5B%5D=20&country%5B%5D=43&country%5B%5D=21&country%5B%5D=172&country%5B%5D=82&country%5B%5D=105&country%5B%5D=247&country%5B%5D=139&country%5B%5D=7&country%5B%5D=188&country%5B%5D=109&country%5B%5D=42&country%5B%5D=111&country%5B%5D=103&country%5B%5D=96&country%5B%5D=68&country%5B%5D=97&country%5B%5D=204&country%5B%5D=94&country%5B%5D=57&country%5B%5D=102&country%5B%5D=92&country%5B%5D=35&country%5B%5D=119&country%5B%5D=10&country%5B%5D=23&country%5B%5D=33&country%5B%5D=66&country%5B%5D=48&country%5B%5D=14&country%5B%5D=106&country%5B%5D=93&country%5B%5D=39&country%5B%5D=51&country%5B%5D=74&country%5B%5D=17&country%5B%5D=22&country%5B%5D=71&country%5B%5D=72&country%5B%5D=89&country%5B%5D=59&country%5B%5D=121&country%5B%5D=24&country%5B%5D=55&country%5B%5D=107&country%5B%5D=113&country%5B%5D=78&country%5B%5D=15&country%5B%5D=122&country%5B%5D=37&country%5B%5D=27&country%5B%5D=232&country%5B%5D=6&country%5B%5D=70&country%5B%5D=32&country%5B%5D=163&country%5B%5D=174&country%5B%5D=8&country%5B%5D=34&country%5B%5D=47&country%5B%5D=145&country%5B%5D=114&country%5B%5D=54&country%5B%5D=25&country%5B%5D=29&country%5B%5D=86&country%5B%5D=95&category%5B%5D=_centralBanks&importance%5B%5D=1&importance%5B%5D=2&importance%5B%5D=3&dateFrom=2023-5-10&dateTo=2023-5-10&timeZone=55&timeFilter=timeOnly&currentTab=custom&submitFilters=1&limit_from=0
