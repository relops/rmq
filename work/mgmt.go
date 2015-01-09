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

func DeleteQueue(rmqc *rabbithole.Client, vhost string, queue string) {
	res, err := rmqc.DeleteQueue(vhost, queue)
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

func CreateMirror(rmqc *rabbithole.Client, vhost string, name string, match string, rf, priority int, nodes ...string) {

	def := rabbithole.PolicyDefinition{}
	def["ha-mode"] = "all"

	if rf > 0 {
		def["ha-mode"] = "exactly"
		def["ha-params"] = rf
	} else if len(nodes) > 0 {
		def["ha-mode"] = "nodes"
		def["ha-params"] = nodes
	}

	p := rabbithole.Policy{}
	p.ApplyTo = "queues"
	p.Pattern = match
	p.Definition = def
	p.Vhost = vhost
	p.Priority = priority
	res, err := rmqc.PutPolicy(vhost, name, p)

	if err != nil {
		log.Errorf("Could not initialize management interface: %s", err)
		os.Exit(1)
	}

	switch res.StatusCode {
	case 204:
		fmt.Printf("Created policy %s (match=%s, priority=%d)\n", name, match, priority)
	case 400:
		fmt.Printf("Error in policy definition %+v, status %d\n", p, res.StatusCode)
	default:
		fmt.Printf("Could not complete operation on policy %+v, status %d\n", p, res.StatusCode)
	}

}

func DeleteMirror(rmqc *rabbithole.Client, vhost string, name string) {
	res, err := rmqc.DeletePolicy(vhost, name)
	if err != nil {
		log.Errorf("Could not initialize management interface: %s", err)
		os.Exit(1)
	}

	switch res.StatusCode {
	case 204:
		fmt.Printf("Deleted HA policy %s\n", name)
	case 404:
		fmt.Printf("HA policy %s not found\n", name)
	default:
		fmt.Printf("Could not complete operation on policy %s, status %d\n", name, res.StatusCode)
	}
}

func Mirroring(rmqc *rabbithole.Client) {

	ps, err := rmqc.ListPolicies()
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

	table.Append([]string{"Policy Name", "Queues", "Replication", "Priority"})
	for _, p := range ps {

		mode, ok := p.Definition["ha-mode"]
		if ok {
			var label string

			switch mode {
			case "all":
				label = "on every node"
			case "exactly":
				replicas, ok := p.Definition["ha-params"].(float64)
				if !ok {
					fmt.Printf("Unknown HA param: %s", p.Definition["ha-params"])
					os.Exit(1)
				}
				label = fmt.Sprintf("on %.0f node(s)", replicas)
			case "nodes":
				label = fmt.Sprintf("on %v", p.Definition["ha-params"])
			default:
				fmt.Printf("Unknown HA mode: %s", mode)
				os.Exit(1)
			}

			info := []string{p.Name, p.Pattern, label, fmt.Sprintf("%d", p.Priority)}
			table.Append(info)
		}
	}

	table.Render()

}
