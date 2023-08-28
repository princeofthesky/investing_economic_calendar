package db

type Country struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Cover       string `json:"cover"`
	InvestingId int    `json:"investing_id"`
	Currency    string `json:"currency"`
}
type EconomicCategory struct {
	Id         int    `json:"id"`
	Title      string `json:"title"`
	ValueQuery string `json:"value_query"`
}
type EventInfo struct {
	Id          int    `json:"id"`
	CountryId   int    `json:"country_id"`
	InvestingId int    `json:"investing_id"`
	Importance  int    `json:"importance"`
	Title       string `json:"title"`
	Actual      string `json:"actual"`
	ForeCast    string `json:"fore_cast"`
	Previous    string `json:"previous"`
	Url         string `json:"url"`
	Time        int64  `json:"time"`
}

type Holiday struct {
	Title     string `json:"title"`
	CountryId int    `json:"country_id"`
	Time      int64  `json:"time"`
	AllDay    bool   `json:"all_day"`
}
type EventList struct {
	EventId    int   `json:"event_id"`
	CategoryId int   `json:"category_id"`
	CountryId  int   `json:"country_id"`
	Importance int   `json:"importance"`
	EventTime  int64 `json:"event_time"`
}
