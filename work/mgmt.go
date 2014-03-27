package work

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/michaelklishin/rabbit-hole"
	"os"
)

func Info(rmqc *rabbithole.Client) {
	o, err := rmqc.Overview()
	if err != nil {
		log.Errorf("Could not initialize management interface: %s", err)
		os.Exit(1)
	}
	fmt.Printf("RabbitMQ Server %s\n", o.RabbitMQVersion)
}
