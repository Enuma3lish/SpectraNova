package conf

import "time"

type Bootstrap struct {
	Server Server `json:"server"`
	Data   Data   `json:"data"`
	Auth   Auth   `json:"auth"`
}

type Server struct {
	HTTP HTTP `json:"http"`
	GRPC GRPC `json:"grpc"`
}

type HTTP struct {
	Addr    string        `json:"addr"`
	Timeout time.Duration `json:"timeout"`
}

type GRPC struct {
	Addr    string        `json:"addr"`
	Timeout time.Duration `json:"timeout"`
}

type Data struct {
	Database Database `json:"database"`
}

type Database struct {
	Driver string `json:"driver"`
	Source string `json:"source"`
}

type Auth struct {
	JWTSecret     string        `json:"jwt_secret"`
	AccessExpiry  time.Duration `json:"access_expiry"`
	RefreshExpiry time.Duration `json:"refresh_expiry"`
}
