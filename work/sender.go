package work

import (
	"fmt"
	"github.com/0x6e6562/gosnow"
	log "github.com/cihub/seelog"
	"github.com/dustin/randbo"
	"github.com/spaolacci/murmur3"
	"github.com/streadway/amqp"
	"math/rand"
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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < opts.Count; i++ {
		sizeInKb := opts.Size * 1024
		size := int(sizeInKb)
		if opts.StdDev > 0 {
			dev := float64(opts.StdDev)
			s := r.NormFloat64()*dev*100 + sizeInKb
			size = int(s)
		}

		if size == 0 {
			size++
		}

		buf := make([]byte, size)
		_, err = randbo.New().Read(buf)
		if err != nil {
			signal <- err
		}

		sum, err := s.send(group, opts, buf)
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

func (s *client) send(group uint64, o *Options, payload []byte) ([]byte, error) {

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

	if o.Persistent {
		envelope.DeliveryMode = amqp.Persistent
	}

	mandatory := false
	immediate := false

	if err := s.ch.Publish(o.Exchange, o.Key, mandatory, immediate, envelope); err != nil {
		return nil, fmt.Errorf("Could not publish to exchange %s: %s", o.Exchange, err)
	}

	h := murmur3.New32()
	h.Write(payload)
	sum := h.Sum(nil)

	size := float32(len(payload)) / 1024
	log.Infof("[%d] sending %.2f kB (%x)", id, size, sum)

	return sum, nil
}
