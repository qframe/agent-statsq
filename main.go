package main

import (
	"log"
	"os"
	"sync"
	"github.com/zpatrick/go-config"
	"github.com/codegangsta/cli"

	"github.com/qnib/qframe-types"
	"github.com/qframe/cache-inventory"
	"github.com/qframe/cache-statsq"
	"github.com/qframe/collector-tcp"
	"github.com/qframe/collector-docker-events"
	"github.com/qframe/handler-influxdb"
	"github.com/qframe/collector-internal"

)

const (
	dockerHost = "unix:///var/run/docker.sock"
	dockerAPI = "v1.29"
)


func check_err(pname string, err error) {
	if err != nil {
		log.Printf("[EE] Failed to create %s plugin: %s", pname, err.Error())
		os.Exit(1)
	}
}

func Run(ctx *cli.Context) {
	// Create conf
	log.Printf("[II] Start Version: %s", ctx.App.Version)

	cfg := config.NewConfig([]config.Provider{})
	if _, err := os.Stat(ctx.String("config")); err == nil {
		log.Printf("[II] Use config file: %s", ctx.String("config"))
		cfg.Providers = append(cfg.Providers, config.NewYAMLFile(ctx.String("config")))
	} else {
		log.Printf("[II] No config file found")
	}
	cfg.Providers = append(cfg.Providers, config.NewCLI(ctx, false))
	qChan := qtypes.NewQChan()
	qChan.Broadcast()
	//////// Handlers
	// Start InfluxDB
	phi, err := qhandler_influxdb.New(qChan, cfg, "influxdb")
	check_err(phi.Name, err)
	go phi.Run()
	//////// Cache
	// StatsQ
	pfs, err := qcache_statsq.New(qChan, cfg, "statsq")
	check_err(pfs.Name, err)
	go pfs.Run()
	// Inventory
	pci, err := qcache_inventory.New(qChan, cfg, "inventory")
	check_err(pci.Name, err)
	go pci.Run()
	//////// Collectors
	// Internal metrics
	pcint, err := qcollector_internal.New(qChan, cfg, "internal")
	check_err(pcint.Name, err)
	go pcint.Run()
	// start docker-events
	pcde, err := qcollector_docker_events.New(qChan, cfg, "docker-events")
	check_err(pcde.Name, err)
	go pcde.Run()
	// TCP
	pct, err := qcollector_tcp.New(qChan, cfg, "tcp")
	check_err(pct.Name, err)
	go pct.Run()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func main() {
	app := cli.NewApp()
	app.Name = "StatsQ agent (StatsD plus tags) to push metrics towards InfluxDB"
	app.Usage = "statsq [options]"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "qframe.yml",
			Usage: "Config file, will overwrite flag default if present.",
		},
	}
	app.Action = Run
	app.Run(os.Args)
}
