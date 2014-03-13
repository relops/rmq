package work

import (
	"errors"
	"github.com/streadway/amqp"
)

var (
	ErrInvalidOptions = errors.New("invalid options")
)

type Options struct {
	Direction string `short:"d" long:"direction" description:"Use rmq to send (-d in) or receive (-d out) messages" required:"true"`
	Exchange  string `short:"x" long:"exchange" description:"The exchange to send to (-d in) or bind a queue to when receiving (-d out)"`
	Queue     string `short:"q" long:"queue" description:"The queue to receive from (when used with -d in)"`
	Key       string `short:"k" long:"key" description:"The key to use for routing (-d in) or for queue binding (-d out)"`
	Count     int    `short:"c" long:"count" description:"The number of messages to send" default:"10"`
	Interval  int    `short:"i" long:"interval" description:"The delay (in ms) between sending messages" default:"10"`
	Size      int    `short:"z" long:"size" description:"Message size in bytes" default:"64"`
	Username  string `short:"u" long:"user" description:"The user to connect as" default:"guest"`
	Password  string `short:"P" long:"pass" description:"The user's password" default:"guest"`
	Host      string `short:"H" long:"host" description:"The Rabbit host to connect to" default:"localhost"`
	Port      int    `short:"p" long:"port" description:"The Rabbit port to connect on" default:"5672"`
	Entropy   bool   `short:"e" long:"entropy" description:"Display message level entropy information" default:"false"`
	Version   func() `short:"V" long:"version" description:"Print rmq version and exit"`
}

func (o *Options) Validate() error {
	if o.Direction != "in" && o.Direction != "out" {
		return ErrInvalidOptions
	}
	return nil
}

func (o *Options) IsSender() bool {
	return o.Direction == "in"
}

func (o *Options) uri() string {
	u := &amqp.URI{
		Username: o.Username,
		Password: o.Password,
		Host:     o.Host,
		Port:     o.Port,
		Scheme:   "amqp",
	}
	return u.String()
}
