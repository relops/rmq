package work

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/dchest/uniuri"
	"github.com/spaolacci/murmur3"
	"github.com/streadway/amqp"
	"hash"
	"time"
)

func StartReceiver(c *client, signal chan error, opts *Options) {

	ch, err := c.openChannel()
	if err != nil {
		signal <- err
		return
	}

	if !opts.NoDeclare {
		q, err := declareQueue(ch, opts.Queue)

		if err != nil {
			signal <- err
			return
		}

		if opts.Queue == "" {
			opts.Queue = q.Name
		}
	}

	if err := ch.Qos(opts.Prefetch, 0, opts.GlobalPrefetch); err != nil {
		signal <- err
		return
	}

	tag := uniuri.NewLen(4)
	deliveries, err := subscribe(ch, opts, tag)
	if err != nil {
		signal <- err
		return
	}

	log.Infof("receiver (%s) subscribed to queue %s (prefetch=%d; global=%v; priority=%d) ", tag, opts.Queue, opts.Prefetch, opts.GlobalPrefetch, opts.Priority)

	cancelSubscription := make(chan bool)
	go handle(deliveries, opts, tag, c.signal, cancelSubscription)

	cancel := make(chan string, 1)
	ch.NotifyCancel(cancel)

	select {
	case tag := <-cancel:
		_ = tag
		cancelSubscription <- true
		if opts.Renew {
			log.Info("automatically renewing subscription")
			c, err := NewClient(opts, c.flake)
			if err != nil {
				signal <- err
				return
			}
			StartReceiver(c, signal, opts)
		}
	}

	signal <- nil
}

func handle(deliveries <-chan amqp.Delivery, opts *Options, tag string, signal chan error, cancelSubscription chan bool) {

	// TODO This could consume a lot of memory if not LRU'ed
	entropy := make(map[string]hash.Hash)

	for {
		select {
		case <-cancelSubscription:
			log.Infof("subscription cancelled by server")
			return
		default:
			for d := range deliveries {

				now := time.Now().UnixNano()

				if opts.Interval > 0 {
					time.Sleep(time.Duration(opts.Interval) * time.Millisecond)
				}

				if err := d.Ack(false); err != nil {
					signal <- err
				}

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

				size := float32(len(d.Body)) / 1024

				var latency float64
				ts, hasLatency := d.Headers[timestampHeader]
				if hasLatency {
					then, ok := ts.(int64)
					if !ok {
						signal <- fmt.Errorf("Invalid nanos timestamp header: %+v", ts)
					}

					latency = float64((now - then)) / (1000 * 1000)
					if latency < 0 {
						log.Warnf("[%s] %s negative latency: %f (ms); sent at %d (nanos), received at %d (nanos)", tag, d.MessageId, latency, then, now)
					}
				}

				if opts.Entropy {
					label := shortLabel(d.CorrelationId)
					if hasLatency {
						log.Infof("[%s] %s receiving %.2f kB (%x) @ %.2f ms [%s, %x]", tag, d.MessageId, size, sum, latency, label, ent)
					} else {
						log.Infof("[%s] %s receiving %.2f kB (%x) [%s, %x]", tag, d.MessageId, size, sum, label, ent)
					}

				} else {
					if hasLatency {
						log.Infof("[%s] %s receiving %.2f kB (%x) @ %.2f ms", tag, d.MessageId, size, sum, latency)
					} else {
						log.Infof("[%s] %s receiving %.2f kB (%x)", tag, d.MessageId, size, sum)
					}

				}
			}
		}
	}

	log.Info("receiver exiting")
	signal <- nil
}

func subscribe(ch *amqp.Channel, o *Options, tag string) (<-chan amqp.Delivery, error) {
	autoAck := false
	exclusive := false
	noLocal := false
	noWait := false
	var args amqp.Table

	if o.Priority > 0 {
		args = amqp.Table{}
		args["x-priority"] = o.Priority
	}

	return ch.Consume(o.Queue, tag, autoAck, exclusive, noLocal, noWait, args)
}

func declareQueue(ch *amqp.Channel, name string) (amqp.Queue, error) {
	durable := false
	autoDelete := true
	exclusive := false
	noWait := false
	var args amqp.Table

	return ch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}
