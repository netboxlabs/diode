package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/netboxlabs/diode/diode-server/auth"
	"github.com/netboxlabs/diode/diode-server/auth/cli"
)

func main() {
	var cfg auth.Config
	envconfig.MustProcess("", &cfg)

	authClientManager := auth.NewHydraClientManager(
		cfg.OAuth2.AdminServerURL,
		slog.New(slog.NewTextHandler(os.Stderr, nil)),
	)
	commands := cli.MakeDefaultCommands(authClientManager)
	cli.RunWithCommands(commands)
}
