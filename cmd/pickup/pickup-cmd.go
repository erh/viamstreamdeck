package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/services/generic"

	"github.com/erh/viamstreamdeck"
	"github.com/erh/vmodutils"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {

	configFile := flag.String("config", "", "config file")
	host := flag.String("host", "", "host to connect to")
	secs := flag.Int("seconds", 60, "seconds to run for")

	flag.Parse()

	ctx := context.Background()
	logger := logging.NewLogger("pickup")

	logger.Infof("using config file [%s] and host [%s]", *configFile, *host)

	if *configFile == "" {
		return fmt.Errorf("need a config file")
	}

	conf := &viamstreamdeck.PickupConfig{}

	err := vmodutils.ReadJSONFromFile(*configFile, conf)
	if err != nil {
		return err
	}

	_, _, err = conf.Validate("")
	if err != nil {
		return err
	}

	client, err := vmodutils.ConnectToHostFromCLIToken(ctx, *host, logger)
	if err != nil {
		return err
	}
	defer client.Close(ctx)

	deps, err := vmodutils.MachineToDependencies(client)
	if err != nil {
		return err
	}

	sd, err := viamstreamdeck.NewPickup(ctx, generic.Named("foo"), deps, conf, logger)
	if err != nil {
		return err
	}
	defer sd.Close(ctx)

	time.Sleep(time.Second * time.Duration(*secs))
	return nil
}
