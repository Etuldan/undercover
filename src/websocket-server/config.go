package main

type Config struct {
	Server struct {
		Port     string `yaml:"port"`
		Host     string `yaml:"host"`
		Endpoint string `yaml:"endpoint"`
	} `yaml:"server"`
	Seq struct {
		Url    string `yaml:"url"`
		ApiKey string `yaml:"apikey"`
	} `yaml:"seq"`
	Game struct {
		NbUndercover int `yaml:"nbUndercover"`
		NbWhite      int `yaml:"nbWhite"`
		NbTurn       int `yaml:"nbTurn"`
	} `yaml:"game"`
	Debug struct {
		GameUuid string `yaml:"gameUuid"`
	} `yaml:"debug"`
}
