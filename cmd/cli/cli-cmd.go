package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"

	"github.com/dh1tw/streamdeck"

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

	configFile := flag.String("config", "config file", "")

	flag.Parse()

	ctx := context.Background()
	logger := logging.NewLogger("streamdeck-cli")

	if *configFile == "" {
		return fmt.Errorf("need a config file")
	}

	conf := &viamstreamdeck.Config{}
	deps := resource.Dependencies{}

	err := vmodutils.ReadJSONFromFile(*configFile, conf)
	if err != nil {
		return err
	}

	sd, err := viamstreamdeck.NewStreamDeck(ctx, generic.Named("foo"), deps, streamdeck.Plus, conf, logger)
	if err != nil {
		return err
	}
	defer sd.Close(ctx)

	time.Sleep(time.Second * 10)
	return nil
}
