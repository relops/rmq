package work

import (
	"fmt"
	"github.com/0x6e6562/gosnow"
	log "github.com/cihub/seelog"
	"github.com/spaolacci/murmur3"
	"github.com/streadway/amqp"
	"hash"
	"time"
)

func StartReceiver(signal chan error, flake *gosnow.SnowFlake, opts *Options) {
	c, err := newClient(opts)

	if err != nil {
		signal <- err
	}

	_, err = declareQueue(c.ch, opts.Queue)

	if err != nil {
		signal <- err
	}

	deliveries, err := subscribe(c.ch, opts.Queue)
	if err != nil {
		signal <- err
	}

	log.Infof("receiver subscribed to queue: %s", opts.Queue)

	go handle(deliveries, opts, c.signal)

}

func handle(deliveries <-chan amqp.Delivery, opts *Options, signal chan error) {

	// TODO This could consume a lot of memory if not LRU'ed
	entropy := make(map[string]hash.Hash)

	for d := range deliveries {

		now := time.Now().UnixNano()

		h := murmur3.New32()
		h.Write(d.Body)
		sum := h.Sum(nil)

		e, ok := entropy[d.CorrelationId]
		if !ok {
			e = murmur3.New32()
		}

		e.Write(sum)
		ent := e.Sum(nil)
		entropy[d.CorrelationId] = e

		var latency float64
		ts, hasLatency := d.Headers[timestampHeader]
		if hasLatency {
			then, ok := ts.(int64)
			if !ok {
				signal <- fmt.Errorf("Invalid nanos timestamp header: %+v", ts)
			}

			latency = float64((now - then)) / (1000 * 1000)
			if latency < 0 {
				log.Warnf("[%s] negative latency: %f (ms); sent at %d (nanos), received at %d (nanos)", d.MessageId, latency, then, now)
			}
		}

		if opts.Entropy {
			label := shortLabel(d.CorrelationId)
			if hasLatency {
				log.Infof("[%s] receiving %d bytes (%x)\t%.2f ms [%s, %x]", d.MessageId, len(d.Body), sum, latency, label, ent)
			} else {
				log.Infof("[%s] receiving %d bytes (%x)\t[%s, %x]", d.MessageId, len(d.Body), sum, label, ent)
			}

		} else {
			if hasLatency {
				log.Infof("[%s] receiving %d bytes\t(%x) in %.2f ms", d.MessageId, len(d.Body), sum, latency)
			} else {
				log.Infof("[%s] receiving %d bytes\t(%x)", d.MessageId, len(d.Body), sum)
			}

		}

	}

	log.Info("receiver exiting")
	signal <- nil
}

func subscribe(ch *amqp.Channel, queue string) (<-chan amqp.Delivery, error) {
	autoAck := false
	exclusive := false
	noLocal := false
	noWait := false
	var args amqp.Table

	return ch.Consume(queue, "", autoAck, exclusive, noLocal, noWait, args)
}

func declareQueue(ch *amqp.Channel, name string) (amqp.Queue, error) {
	durable := false
	autoDelete := true
	exclusive := false
	noWait := false
	var args amqp.Table

	return ch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}
