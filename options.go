package mgoc

import (
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type dialOption struct {
	Debug          bool                     // enable debug mode
	Max            int                      // max active connections
	Idle           int                      // max idle connections
	SSH            *SSH                     // SSH tunnel server config
	ConnectTimeout int                      // connect timeout
	WriteTimeout   int                      // write timeout seconds
	ReadTimeout    int                      // read timeout seconds
	DatabaseOpt    *options.DatabaseOptions // database options
}

type Option func(*dialOption)

var defaultDialOption = &dialOption{
	ConnectTimeout: defaultConnectTimeoutSeconds,
	WriteTimeout:   defaultWriteTimeoutSeconds,
	ReadTimeout:    defaultReadTimeoutSeconds,
}

func makeOption(opts ...Option) *dialOption {
	for _, opt := range opts {
		opt(defaultDialOption)
	}
	log.Json(defaultDialOption)
	return defaultDialOption
}

func WithDebug() Option {
	return func(opt *dialOption) {
		opt.Debug = true
	}
}

func WithMaxConn(max int) Option {
	return func(opt *dialOption) {
		opt.Max = max
	}
}

func WithIdleConn(idle int) Option {
	return func(opt *dialOption) {
		opt.Idle = idle
	}
}

func WithConnectTimeout(timeout int) Option {
	return func(opt *dialOption) {
		opt.ConnectTimeout = timeout
	}
}

func WithWriteTimeout(timeout int) Option {
	return func(opt *dialOption) {
		opt.WriteTimeout = timeout
	}
}

func WithReadTimeout(timeout int) Option {
	return func(opt *dialOption) {
		opt.ReadTimeout = timeout
	}
}

func WithDatabaseOpt(opt *options.DatabaseOptions) Option {
	return func(d *dialOption) {
		d.DatabaseOpt = opt
	}
}

func WithSSH(ssh *SSH) Option {
	return func(opt *dialOption) {
		opt.SSH = ssh
	}
}
