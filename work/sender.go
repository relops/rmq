package work

import (
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/dustin/randbo"
	"github.com/spaolacci/murmur3"
	"github.com/streadway/amqp"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func StartSender(s *client, signal chan error, opts *Options, wg *sync.WaitGroup) {

	ch, err := s.openChannel()
	if err != nil {
		signal <- err
		return
	}

	group, err := s.flake.Next()
	if err != nil {
		signal <- err
		return
	}

	h := murmur3.New32()

	if len(opts.Args.MessageBody) > 0 {

		m := make(map[string]string)
		for _, kv := range opts.Args.MessageBody {
			s := strings.SplitN(kv, "=", 2)
			if len(s) == 1 {
				m[s[0]] = ""
			} else {
				m[s[0]] = s[1]
			}
		}
		encoded, err := json.Marshal(m)
		if err != nil {
			signal <- err
			return
		}

		sum, err := s.send(ch, group, opts, encoded)
		if err != nil {
			signal <- err
			return
		}

		h.Write(sum)

	} else {

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
				return
			}

			sum, err := s.send(ch, group, opts, buf)
			if err != nil {
				signal <- err
				return
			}

			h.Write(sum)

			time.Sleep(time.Duration(opts.Interval) * time.Millisecond)
		}
	}

	if opts.Entropy {
		log.Infof("[%d] sender entropy (%x)", group, h.Sum(nil))
	}

	wg.Done()
	signal <- nil
}

func (s *client) send(ch *amqp.Channel, group uint64, o *Options, payload []byte) ([]byte, error) {

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

	if err := ch.Publish(o.Exchange, o.Key, mandatory, immediate, envelope); err != nil {
		return nil, fmt.Errorf("Could not publish to exchange %s: %s", o.Exchange, err)
	}

	h := murmur3.New32()
	h.Write(payload)
	sum := h.Sum(nil)

	size := float32(len(payload)) / 1024

	if len(o.Verbose) > 0 {
		log.Infof("[%d] sending %.2f kB (%x) to %s", id, size, sum, o.Key)
	} else {
		log.Infof("[%d] sending %.2f kB (%x)", id, size, sum)
	}

	return sum, nil
}
