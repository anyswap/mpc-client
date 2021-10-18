// Command mpc-client is main program.
package main

import (
	"fmt"
	"os"

	"github.com/anyswap/mpc-client/cmd/utils"
	"github.com/anyswap/mpc-client/log"
	"github.com/urfave/cli/v2"
)

var (
	clientIdentifier = "MPC-Client"
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(clientIdentifier, gitCommit, gitDate, "the MPC-Client command line interface")
)

func initApp() {
	// Initialize the CLI app and start action
	app.Action = mpcclient
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2020-2021 The MPC-Client Authors"
	app.Commands = []*cli.Command{
		sendEthTxCommand,
		signPlainTextCommand,
		acceptSignCommand,
		getAcceptListCommand,
		getEnodeCommand,
		getGroupCommand,
		utils.LicenseCommand,
		utils.VersionCommand,
	}
	app.Flags = []cli.Flag{
		utils.LogFileFlag,
		utils.LogRotationFlag,
		utils.LogMaxAgeFlag,
		utils.VerbosityFlag,
		utils.JSONFormatFlag,
		utils.ColorFormatFlag,
	}
}

func main() {
	initApp()
	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func mpcclient(ctx *cli.Context) error {
	utils.SetLogger(ctx)
	if ctx.NArg() > 0 {
		return fmt.Errorf("invalid command: %q", ctx.Args().Get(0))
	}

	return cli.ShowAppHelp(ctx)
}
