package main

import (
	"fmt"
	"github.com/0x6e6562/gosnow"
	log "github.com/cihub/seelog"
	"github.com/jessevdk/go-flags"
	"github.com/michaelklishin/rabbit-hole"
	"github.com/relops/rmq/work"
	"os"
	"strings"
	"sync"
)

var logConfig = `
<seelog type="sync">
	<outputs formatid="main">
		<console/>
	</outputs>
	<formats>
		<format id="main" format="%Date(2006-02-01 03:04:05.000) - %Msg%n"/>
	</formats>
</seelog>`

var (
	opts    work.Options
	parser         = flags.NewParser(&opts, flags.Default)
	VERSION string = "0.2.1"
)

func init() {
	opts.AdvertizedVersion = VERSION
	opts.Version = printVersionAndExit

	// We might want to make this overridable
	logger, err := log.LoggerFromConfigAsString(logConfig)

	if err != nil {
		fmt.Printf("Could not load seelog configuration: %s\n", err)
		return
	}

	log.ReplaceLogger(logger)
}

func main() {

	if _, err := parser.Parse(); err != nil {
		os.Exit(0)
	}

	if err := opts.Validate(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}

	if opts.UsesMgmt() {

		// TODO make more configurable
		rmqc, _ := rabbithole.NewClient("http://127.0.0.1:15672", "guest", "guest")

		if opts.Info {
			work.Info(rmqc)
		}

		if len(opts.QueueInfo) > 0 {
			work.Queues(rmqc)
		}
		os.Exit(0)
	}

	flake, err := gosnow.NewSnowFlake(201)
	if err != nil {
		log.Errorf("Could not initialize snowflake: %s", err)
		os.Exit(1)
	}

	signal := make(chan error)

	var wg sync.WaitGroup

	for i := 0; i < (&opts).Connections; i++ {

		c, err := work.NewClient(&opts, flake)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		for i := 0; i < (&opts).Concurrency; i++ {

			if opts.IsSender() {

				wg.Add(1)
				go work.StartSender(c, signal, &opts, &wg)

			} else {
				go work.StartReceiver(c, signal, &opts)
			}

		}
	}

	err = <-signal

	if err != nil {
		if shouldLogError(err) {
			log.Error(err)
		}
		os.Exit(1)
	}

	wg.Wait()
}

func shouldLogError(err error) bool {

	if strings.Contains(err.Error(), "PRECONDITION_FAILED") {
		log.Error(err)
		log.Info("Potential attempt to redeclare an existing queue - to avoid this, use the -n option")
		return false
	}

	return true

}

func printVersionAndExit() {
	fmt.Fprintf(os.Stderr, "%s %s\n", "rmq", VERSION)
	os.Exit(0)
}
