package config

type Config struct {
	Api               Api `yaml:"api"`
	DefaultPartitions int `yaml:"default_partitions"`
}

type Api struct {
	Port int `yaml:"port"`
}
