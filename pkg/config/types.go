package config

type Config struct {
	Api        Api        `yaml:"api"`
	Message    Message    `yaml:"message"`
	Subscriber Subscriber `yaml:"subscriber"`
	Topic      Topic      `yaml:"topic"`
}

type Api struct {
	Port int `yaml:"port"`
}

type Subscriber struct {
	CleanupTime   int `yaml:"cleanup_time"`
	RetryCount    int `yaml:"retry_count"`
	RetryInterval int `yaml:"retry_interval"`
	Timeout       int `yaml:"timeout"`
	InactiveTime  int `yaml:"inactive_time"`
}

type Topic struct {
	Buffer int `yaml:"buffer"`
}

type Message struct {
	CleanupTime int `yaml:"cleanup_time"`
	TTL         int `yaml:"ttl"`
}
