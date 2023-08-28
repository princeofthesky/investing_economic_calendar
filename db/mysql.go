package db

import (
	"context"
	"github.com/go-pg/pg/v10"
	"go-pinterest/config"
	"go-pinterest/date"
	"strconv"
)

var mysqlDb *pg.DB

func Init(postgres config.Postgres) error {
	mysqlDb = pg.Connect(&pg.Options{
		Addr:     postgres.Addr,
		User:     postgres.User,
		Password: postgres.Password,
		Database: postgres.Database,
	})
	return mysqlDb.Ping(context.Background())
}
func Close() {
	mysqlDb.Close()
}
func GetDb() *pg.DB {
	return mysqlDb
}

func GetAllCountries() ([]Country, error) {
	countries := make([]Country, 0)
	err := mysqlDb.Model((*Country)(nil)).Column("*").Order("id DESC").ForEach(
		func(c *Country) error {
			countries = append(countries, *c)
			return nil
		})

	return countries, err
}

func GetAllCategories() ([]EconomicCategory, error) {
	categories := make([]EconomicCategory, 0)
	err := mysqlDb.Model((*EconomicCategory)(nil)).Column("*").Order("id DESC").ForEach(
		func(c *EconomicCategory) error {
			categories = append(categories, *c)
			return nil
		})

	return categories, err
}

func GetEventIdByInvestingId(investingId int) (EventInfo, error) {
	info := EventInfo{Id: 0, InvestingId: investingId}
	err := mysqlDb.Model(&info).Where("\"investing_id\"=?", info.InvestingId).Select()
	return info, err
}

func GetEventInfoById(Id int) (EventInfo, error) {
	info := EventInfo{Id: Id}
	err := mysqlDb.Model(&info).WherePK().Select()
	return info, err
}

func InsertEventInfo(info EventInfo) (EventInfo, error) {
	_, err := mysqlDb.Model(&info).Insert()
	return info, err
}

func UpdateCurrency(countryId int ,curr string) (Country, error) {
	country:=Country{Id: countryId,Currency: curr}
	_, err := mysqlDb.Model(&country).WherePK().Column("currency").Update()
	return country,err
}

func InsertEventToList(info EventList) (EventList, error) {
	_, err := mysqlDb.Model(&info).Insert()
	return info, err
}
func InsertHoliday(info Holiday) (Holiday, error) {
	_, err := mysqlDb.Model(&info).Insert()
	return info, err
}


func GetAllHolidays(date int64) ([]Holiday, error) {
	data := make([]Holiday, 0)
	err := mysqlDb.Model((*Holiday)(nil)).Column("*").Where("time = ?",date).Order("time DESC").ForEach(
		func(c *Holiday) error {
			data = append(data, *c)
			return nil
		})

	return data, err
}

func GetEventList(countries []int,categories []int, day int64) ([]EventList, error) {
	WhereCondition := "event_time >= " + strconv.FormatInt(day,10) +" AND event_time < " + strconv.FormatInt(day+date.SECOND_PER_DAY,10)

	if len(categories) > 0 {
		addedQuery := " AND ( category_id = " + strconv.Itoa(categories[0])
		for i := 1; i < len(categories); i++ {
			addedQuery = addedQuery + " OR category_id = " + strconv.Itoa(categories[i])
		}
		WhereCondition = WhereCondition + addedQuery + " ) "
	}

	if len(countries) > 0 {
		addedQuery := " AND ( country_id = " + strconv.Itoa(countries[0])
		for i := 1; i < len(countries); i++ {
			addedQuery = addedQuery + " OR country_id = " + strconv.Itoa(countries[i])
		}
		WhereCondition = WhereCondition + addedQuery + " ) "
	}
	data := []EventList{}
	err := mysqlDb.Model().Table("event_lists").Where(WhereCondition).Order("event_time DESC").ForEach(
		func(c *EventList) error {
			data = append(data, *c)
			return nil
		})
	return data, err
}
