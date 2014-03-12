package work

import (
	"github.com/0x6e6562/gosnow"
	log "github.com/cihub/seelog"
	"github.com/streadway/amqp"
)

type client struct {
	con    *amqp.Connection
	ch     *amqp.Channel
	signal chan error
	flake  *gosnow.SnowFlake
}

func newClient(opts *Options) (*client, error) {
	var err error
	s := client{}

	s.con, err = amqp.Dial(opts.uri())
	if err != nil {
		return nil, err
	}

	s.ch, err = s.con.Channel()
	if err != nil {
		return nil, err
	}

	direction := "receiver"
	if opts.IsSender() {
		direction = "sender"
	}
	log.Infof("%s connected to %s", direction, opts.Host)

	return &s, err
}

func shortLabel(c string) string {
	l := len(c)
	return c[l-7 : l-1]
}
