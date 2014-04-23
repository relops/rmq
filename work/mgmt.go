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

func DeleteQueue(rmqc *rabbithole.Client, queue string) {
	res, err := rmqc.DeleteQueue("/", queue)
	if err != nil {
		log.Errorf("Could not initialize management interface: %s", err)
		os.Exit(1)
	}

	switch res.StatusCode {
	case 204:
		fmt.Printf("Deleted %s\n", queue)
	case 404:
		fmt.Printf("Queue %s not found\n", queue)
	default:
		fmt.Printf("Could not complete operation on queue %s, status %d\n", queue, res.StatusCode)
	}

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
