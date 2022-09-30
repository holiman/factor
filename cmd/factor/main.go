// Copyright 2022 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/factor/lib"
	"github.com/naoina/toml"
	"github.com/urfave/cli/v2"
)

func NewApp(usage string) *cli.App {
	app := cli.NewApp()
	app.Usage = usage
	app.Copyright = "Copyright 2022 The go-ethereum Authors"
	return app
}

var (
	app           = NewApp("Exection Layer Relay")
	verbosityFlag = &cli.IntFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail",
		Value: 3,
	}
	configFileFlag = &cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
		Value: "conf.toml",
	}
)

func init() {
	app.Name = "Factor"
	app.Flags = []cli.Flag{
		configFileFlag,
		verbosityFlag,
	}
	app.Action = relay

}
func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func relay(c *cli.Context) error {
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(c.Int(verbosityFlag.Name)), log.StreamHandler(os.Stdout, log.TerminalFormat(true))))
	conffile := c.String(configFileFlag.Name)
	var config lib.Config
	if data, err := os.ReadFile(conffile); err != nil {
		log.Error("Error reading config file", "file", conffile, "err", err)
		return err
	} else if err := toml.Unmarshal(data, &config); err != nil {
		log.Error("Error reading config file", "file", config, "err", err)
		return err
	}
	log.Info("Spinning up muxer...")
	mux, err := lib.NewRelayPI(config.ElClients)
	if err != nil {
		return err
	}
	log.Info("Spinning up relayer...")
	fetcher, err := lib.NewFetcher(config.ClClient, mux)
	fetcher.Start()
	abortChan := make(chan os.Signal, 1)
	signal.Notify(abortChan, os.Interrupt)
	sig := <-abortChan
	log.Info("Exiting...", "signal", sig)
	fetcher.Stop()
	return nil
}
