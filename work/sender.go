package work

import (
	"fmt"
	"github.com/0x6e6562/gosnow"
	log "github.com/cihub/seelog"
	"github.com/dustin/randbo"
	"github.com/spaolacci/murmur3"
	"github.com/streadway/amqp"
	"time"
)

func StartSender(signal chan error, flake *gosnow.SnowFlake, opts *Options) {
	s, err := newClient(opts)

	if err != nil {
		signal <- err
	}

	s.flake = flake

	group, err := s.flake.Next()
	if err != nil {
		signal <- err
	}

	h := murmur3.New32()

	for i := 0; i < opts.Count; i++ {

		buf := make([]byte, opts.Size)
		_, err = randbo.New().Read(buf)
		if err != nil {
			signal <- err
		}

		sum, err := s.send(group, opts.Exchange, opts.Key, buf)
		if err != nil {
			signal <- err
		}

		h.Write(sum)

		time.Sleep(time.Duration(opts.Interval) * time.Millisecond)
	}

	if opts.Entropy {
		log.Infof("[%d] sender entropy (%x)", group, h.Sum(nil))
	}

	signal <- nil
}

func (s *client) send(group uint64, x, key string, payload []byte) ([]byte, error) {

	id, err := s.flake.Next()
	if err != nil {
		return nil, err
	}

	groupString := fmt.Sprintf("%d", group)

	envelope := amqp.Publishing{
		MessageId:     fmt.Sprintf("%d", id),
		CorrelationId: groupString,
		Body:          payload,
		DeliveryMode:  amqp.Transient,
		Headers:       amqp.Table{timestampHeader: time.Now().UnixNano()},
	}

	mandatory := false
	immediate := false

	if err := s.ch.Publish(x, key, mandatory, immediate, envelope); err != nil {
		return nil, fmt.Errorf("Could not publish to exchange %s: %s", x, err)
	}

	h := murmur3.New32()
	h.Write(payload)
	sum := h.Sum(nil)

	log.Infof("[%d] sending %d bytes (%x)", id, len(payload), sum)

	return sum, nil
}
