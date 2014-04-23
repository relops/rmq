package work

import (
	"github.com/0x6e6562/gosnow"
	log "github.com/cihub/seelog"
	"github.com/streadway/amqp"
)

const (
	timestampHeader = "nanos"
)

type client struct {
	con    *amqp.Connection
	signal chan error
	flake  *gosnow.SnowFlake
}

func NewClient(opts *Options, flake *gosnow.SnowFlake) (*client, error) {
	var err error
	s := client{flake: flake}

	config := amqp.Config{Properties: amqp.Table{"product": "rmq", "version": opts.AdvertizedVersion}}

	s.con, err = amqp.DialConfig(opts.uri(), config)
	if err != nil {
		return nil, err
	}

	blockings := s.con.NotifyBlocked(make(chan amqp.Blocking))
	go func() {
		for b := range blockings {
			if b.Active {
				log.Warnf("Connection blocked: %q", b.Reason)
			} else {
				log.Warn("Connection unblocked")
			}
		}
	}()

	direction := "receiver"
	if opts.IsSender() {
		direction = "sender"
	}
	log.Infof("%s connected to %s", direction, opts.Host)

	return &s, err
}

func (c *client) openChannel() (*amqp.Channel, error) {
	return c.con.Channel()
}

func shortLabel(c string) string {
	l := len(c)
	return c[l-7 : l-1]
}
