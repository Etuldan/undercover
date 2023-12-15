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
}
