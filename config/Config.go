package config

type Postgres struct {
	Addr     string
	User     string
	Password string
	Database string
}

type CrawlConfig struct {
	Postgres  Postgres
	StartDate int
	EndDate   int
}
