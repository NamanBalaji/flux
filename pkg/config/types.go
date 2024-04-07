package config

type Config struct {
	Api Api `yaml:"api"`
}

type Api struct {
	Port int `yaml:"port"`
}
