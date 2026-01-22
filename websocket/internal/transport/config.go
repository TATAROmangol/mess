package transport

import "time"

type ClientConfig struct {
	MessageBuffer int           `yaml:"message_buffer"`
	WriteTimeout  time.Duration `yaml:"write_timeout"`
	ReadTimeout   time.Duration `yaml:"read_timeout"`
	PingPeriod    time.Duration `yaml:"ping_timeout"`
}

type WSHandlerConfig struct {
	ReadBufferSizeBytes  int          `yaml:"read_buffer_size_bytes"`
	WriteBufferSizeBytes int          `yaml:"write_buffer_size_bytes"`
	ClientConfig         ClientConfig `yaml:"client"`
}

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}
