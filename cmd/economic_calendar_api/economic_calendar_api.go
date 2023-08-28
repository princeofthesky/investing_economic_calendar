package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml"
	"go-pinterest/config"
	"go-pinterest/date"
	"go-pinterest/db"
	"go-pinterest/model"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	httpPort = flag.String("http_port", "9080", "http_port listen")
	conf     = flag.String("conf", "./investing_economic_calender.toml", "config run file *.toml")
	c        = config.CrawlConfig{}
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
		fmt.Println("err", err)
	}
	defer db.Close()
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/economic-calendar/countries", GetAllCountries)
	r.GET("/economic-calendar/categories", GetAllCategories)
	r.POST("/economic-calendar/events", GetAllEvents)
	r.Run(":" + *httpPort)
}

func GetAllCountries(c *gin.Context) {
	countries, _ := db.GetAllCountries()
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: countries})
	c.Data(200, "text/html; charset=UTF-8", data)
}

func GetAllCategories(c *gin.Context) {
	categories, _ := db.GetAllCategories()
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: categories})
	c.Data(200, "text/html; charset=UTF-8", data)
}

func GetAllEvents(c *gin.Context) {
	request, _ := io.ReadAll(c.Request.Body)
	eventQuery := model.EventQuery{}
	json.Unmarshal(request, &eventQuery)
	fmt.Println(eventQuery.From)
	startDate := date.ParseDate(eventQuery.From + " 0:0:0")
	endDate := date.ParseDate(eventQuery.To + " 0:0:0")
	dayQuery := startDate
	var responses []model.EventResponse
	for dayQuery <= endDate {
		response := model.EventResponse{}
		response.Date = date.FormatDate(time.Unix(dayQuery, 0))
		eventList, err := db.GetEventList(eventQuery.Countries, eventQuery.Categories, dayQuery)
		if err != nil {
			fmt.Println("err when get event list ", err.Error())
		}
		response.Holidays, err = db.GetAllHolidays(dayQuery)
		if err != nil {
			fmt.Println("err when get all holiday list ", err.Error())
		}
		response.Events = []db.EventInfo{}
		for i := 0; i < len(eventList); i++ {
			event := eventList[i]
			eventInfo, _ := db.GetEventInfoById(event.EventId)
			response.Events = append(response.Events, eventInfo)
		}
		dayQuery = dayQuery + date.SECOND_PER_DAY
		responses = append(responses, response)
	}
	data, _ := json.Marshal(model.HttpResponse{Code: 0, Message: "", Data: responses})
	c.Data(200, "text/html; charset=UTF-8", data)
}
