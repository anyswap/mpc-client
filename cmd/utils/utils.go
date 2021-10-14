// Package utils provides common sub commands and command flags.
package utils

import (
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/anyswap/mpc-client/log"
	"github.com/anyswap/mpc-client/params"
	"github.com/urfave/cli/v2"
)

var (
	clientIdentifier string
	gitCommit        string
	gitDate          string
)

// catch signal and cleanup related
var (
	CleanupChan  = make(chan struct{})
	TopWaitGroup = new(sync.WaitGroup)
)

// NewApp creates an app with sane defaults.
func NewApp(identifier, gitcommit, gitdate, usage string) *cli.App {
	notifySignals()
	clientIdentifier = identifier
	gitCommit = gitcommit
	gitDate = gitdate
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Version = params.VersionWithCommit(gitCommit, gitDate)
	app.Usage = usage
	return app
}

func notifySignals() {
	signal.Reset() // to cancal imported mod (eg. okex) to catch signal and call os.Exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)
	go func() {
		sig := <-signalChan
		log.Info("receive interrupt signal", "signal", sig)
		close(CleanupChan)
		<-time.After(1 * time.Second)
		os.Exit(1)
	}()
	go func() {
		<-CleanupChan
		sig := <-signalChan
		log.Info("receive duplicate interrupt signal and exit", "signal", sig)
		os.Exit(1)
	}()
}

// IsCleanuping is cleanuping
func IsCleanuping() bool {
	select {
	case <-CleanupChan:
		return true
	default:
		return false
	}
}

// WaitAndCleanup wait and cleanup
func WaitAndCleanup(doCleanup func()) {
	<-CleanupChan
	doCleanup()
}
