package work

import (
	"fmt"
	"github.com/streadway/amqp"
)

type Options struct {
	QueueInfo         string  `short:"Q" long:"queue-info" description:"List queues" optional:"true" optional-value:"*"`
	Direction         string  `short:"d" long:"direction" description:"Use rmq to send (-d in) or receive (-d out) messages"`
	Exchange          string  `short:"x" long:"exchange" description:"The exchange to send to (-d in) or bind a queue to when receiving (-d out)"`
	Queue             string  `short:"q" long:"queue" description:"The queue to receive from (when used with -d in)"`
	Persistent        bool    `short:"P" long:"persistent" description:"Use persistent messaging" default:"false"`
	NoDeclare         bool    `short:"n" long:"no-declare" description:"If set, then don't attempt to declare the queue or bind it" default:"false"`
	Prefetch          int     `short:"f" long:"prefetch" description:"The number of outstanding acks a receiver will be limited to, default of 0 means unbounded" default:"0"`
	Priority          int32   `short:"y" long:"priority" description:"The relative priority for receiving messages" default:"0"`
	GlobalPrefetch    bool    `short:"G" long:"global-prefetch" description:"Whether to share the prefetch limit accross all consumers of a channel" default:"false"`
	Key               string  `short:"k" long:"key" description:"The key to use for routing (-d in) or for queue binding (-d out)"`
	Count             int     `short:"c" long:"count" description:"The number of messages to send" default:"10"`
	Interval          int     `short:"i" long:"interval" description:"The delay (in ms) between sending or receiving messages" default:"0"`
	Delete            bool    `short:"D" long:"delete" description:"If set, it will remove the queue specified by the -q argument" default:"false"`
	Info              bool    `short:"I" long:"info" description:"If set, print basic server info (requires management API to be installed on the server)" default:"false"`
	Concurrency       int     `short:"g" long:"concurrency" description:"The number of processes per connection" default:"1"`
	Connections       int     `short:"m" long:"connections" description:"The number of connections to use" default:"1"`
	Size              float64 `short:"z" long:"size" description:"Message size in kB" default:"1"`
	StdDev            int     `short:"t" long:"stddev" description:"Standard deviation of message size" default:"0"`
	Renew             bool    `short:"r" long:"renew" description:"Automatically resubscribe when the server cancels a subscription (used for mirrored queues)" default:"false"`
	Username          string  `short:"u" long:"user" description:"The user to connect as" default:"guest"`
	Password          string  `short:"w" long:"pass" description:"The user's password" default:"guest"`
	Host              string  `short:"H" long:"host" description:"The Rabbit host to connect to" default:"localhost"`
	Port              int     `short:"p" long:"port" description:"The Rabbit port to connect on" default:"5672"`
	Entropy           bool    `short:"e" long:"entropy" description:"Display message level entropy information" default:"false"`
	Version           func()  `short:"V" long:"version" description:"Print rmq version and exit"`
	Verbose           []bool  `short:"v" long:"verbose" description:"Show verbose debug information"`
	AdvertizedVersion string
}

func (o *Options) UsesMgmt() bool {
	return o.Info || len(o.QueueInfo) > 0 || o.Delete
}

func (o *Options) Validate() error {
	if len(o.Direction) > 0 {
		if o.Direction != "in" && o.Direction != "out" {
			return fmt.Errorf("Invalid argument: Illegal direction: %s (must be either 'in' or 'out')", o.Direction)
		}
	}

	if o.Direction == "in" {
		if len(o.Queue) > 0 {
			return fmt.Errorf("Invalid argument: Should not specify a queue on ingress")
		}
		if len(o.Key) < 1 {
			return fmt.Errorf("Invalid argument: Empty routing key")
		}
	}

	if o.Direction == "out" {
		if len(o.Key) > 0 {
			return fmt.Errorf("Invalid argument: Should not specify a routing key on egress")
		}
		if len(o.Queue) < 1 {
			return fmt.Errorf("Invalid argument: Empty queue name")
		}
	}

	if o.Size < 1 {
		return fmt.Errorf("Invalid argument: Illegal message size: %f", o.Size)
	}
	if o.StdDev < 0 {
		return fmt.Errorf("Invalid argument: Negative standard deviation: %d", o.StdDev)
	}

	if len(o.Direction) > 0 && o.UsesMgmt() {
		return fmt.Errorf("Invalid argument: cannot use management and messaging in the same command")
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
