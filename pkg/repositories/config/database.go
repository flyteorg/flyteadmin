package config

// Database config. Contains values necessary to open a database connection.
type DbConfig struct {
	BaseConfig
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DbName       string `json:"dbname"`
	User         string `json:"user"`
	Password     string `json:"password"`
	RootCA       string `json:"rootCA"`
	ExtraOptions string `json:"options"`

	Region string `json:"region"` // AWS specific
	UseIAM bool   `json:"useIAM"`
}
