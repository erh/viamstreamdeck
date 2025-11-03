package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"go.viam.com/rdk/cli"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
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

	configFile := flag.String("config", "config file", "")
	host := flag.String("host", "", "")
	sleep := flag.Int("sleep", 10, "")

	flag.Parse()

	ctx := context.Background()
	logger := logging.NewLogger("streamdeck-cli")

	if *configFile == "" {
		return fmt.Errorf("need a config file")
	}

	conf := &viamstreamdeck.Config{}
	err := vmodutils.ReadJSONFromFile(*configFile, conf)
	if err != nil {
		return err
	}

	deps := resource.Dependencies{}

	if *host == "" {
		_, things, err := conf.Validate("")
		if err != nil {
			return err
		}

		for _, t := range things {
			if t == "NO" {
				continue
			}
			n := generic.Named(t)
			deps[n] = &TestThing{
				name:   n,
				logger: logger.Sublogger(t),
			}
		}
	} else {
		client, err := cli.ConnectToMachine(ctx, *host, logger)
		if err != nil {
			return err
		}
		defer client.Close(ctx)

		deps, err = vmodutils.MachineToDependencies(client)
		if err != nil {
			return err
		}
	}

	sd, err := viamstreamdeck.NewStreamDeck(ctx, generic.Named("foo"), deps, nil, conf, logger)
	if err != nil {
		return err
	}
	defer sd.Close(ctx)

	time.Sleep(time.Second * time.Duration(*sleep))
	return nil
}

type TestThing struct {
	resource.AlwaysRebuild
	resource.TriviallyCloseable

	name   resource.Name
	logger logging.Logger
}

func (tt *TestThing) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	tt.logger.Infof("TestThing::DoCommand args: %v", cmd)
	return nil, nil
}

func (tt *TestThing) Name() resource.Name {
	return tt.name
}
