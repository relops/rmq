package main

import (
	"fmt"
	"github.com/0x6e6562/gosnow"
	log "github.com/cihub/seelog"
	"github.com/jessevdk/go-flags"
	"github.com/relops/rmq/work"
	"os"
	"strings"
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
	VERSION string = "0.1.0"
)

func init() {

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
		if !strings.Contains(err.Error(), "Usage:") && !strings.Contains(err.Error(), "direction") {
			fmt.Fprintf(os.Stderr, "Initialization error: %s\n", err)
		}
		os.Exit(1)
	}

	if err := opts.Validate(); err != nil {
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	flake, err := gosnow.NewSnowFlake(201)
	if err != nil {
		log.Errorf("Could not initialize snowflake: %s", err)
		os.Exit(1)
	}

	signal := make(chan error)

	if opts.IsSender() {
		go work.StartSender(signal, flake, &opts)
	} else {
		go work.StartReceiver(signal, flake, &opts)
	}

	err = <-signal

	if err != nil {
		log.Error(err)
	}
}

func printVersionAndExit() {
	fmt.Fprintf(os.Stderr, "%s %s\n", "rmq", VERSION)
	os.Exit(0)
}
