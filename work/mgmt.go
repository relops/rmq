package work

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/michaelklishin/rabbit-hole"
	"github.com/olekukonko/tablewriter"
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

func Queues(rmqc *rabbithole.Client) {

	qs, err := rmqc.ListQueues()
	if err != nil {
		log.Errorf("Could not initialize management interface: %s", err)
		os.Exit(1)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetRowSeparator("")
	table.SetColumnSeparator("")
	table.SetCenterSeparator("")
	table.SetAlignment(tablewriter.ALIGN_CENTRE)

	table.Append([]string{"Queue", "Ready"})
	for _, q := range qs {
		info := []string{q.Name, fmt.Sprintf("%d", q.Messages)}
		table.Append(info)
	}
	table.Render()

}
