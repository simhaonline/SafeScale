/*
 * Copyright 2018-2020, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"sort"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/CS-SI/SafeScale/cli/safescale/commands"
	"github.com/CS-SI/SafeScale/lib/client"
	"github.com/CS-SI/SafeScale/lib/server/utils"
	"github.com/CS-SI/SafeScale/lib/utils/debug"
	"github.com/CS-SI/SafeScale/lib/utils/temporal"

	// Autoload embedded provider drivers
	_ "github.com/CS-SI/SafeScale/lib/server"
)

var profileCloseFunc = func() {}

func cleanup(onAbort bool) {
	if onAbort {
		fmt.Println("\nBe careful stopping safescale will not stop the execution on safescaled, but will try to go back to the previous state as much as possible!")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you really want to stop the command ? [y]es [n]o: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("failed to read the input : ", err.Error())
			text = "y"
		}
		if strings.TrimRight(text, "\n") == "y" {
			err = client.New().JobManager.Stop(utils.GetUUID(), temporal.GetExecutionTimeout())
			if err != nil {
				fmt.Printf("failed to stop the process %v\n", err)
			}
		}
	}
	profileCloseFunc()
	os.Exit(0)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-c
			cleanup(true)
		}
	}()

	app := cli.NewApp()
	app.Writer = os.Stderr
	app.Name = "safescale"
	app.Usage = "safescale COMMAND"
	app.Version = Version + ", build " + Revision + " compiled with "+runtime.Version()+" (" + BuildDate + ")"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "CS-SI",
			Email: "safescale@c-s.fr",
		},
	}

	app.EnableBashCompletion = true

	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, V",
		Usage: "Print program version",
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Increase verbosity",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "Show debug information",
		},
		cli.StringFlag{
			Name:  "profile",
			Usage: "Profiles binary; can contain 'cpu', 'ram', 'web' and a combination of them (ie 'cpu,ram')",
			// TODO: extends profile to accept <what>:params, for example cpu:$HOME/safescale.cpu.pprof, or web:192.168.2.1:1666
		},
		// cli.IntFlag{
		// 	Name:  "port, p",
		// 	Usage: "Bind to specified port `PORT`",
		// 	Value: 50051,
		// },
	}

	app.Before = func(c *cli.Context) error {
		// Define trace settings of the application (what to trace if trace is wanted)
		// NOTE: is it the good behavior ? Shouldn't we fail ?
		// If trace settings cannot be registered, report it but do not fail
		// err := debug.RegisterTraceSettings(appTrace)
		// if err != nil {
		// 	logrus.Errorf(err.Error())
		// }

		// Sets profiling
		if c.IsSet("profile") {
			what := c.String("profile")
			profileCloseFunc = debug.Profile(what)
		}

		if strings.Contains(path.Base(os.Args[0]), "-cover") {
			logrus.SetLevel(logrus.TraceLevel)
			utils.Verbose = true
		} else {
			logrus.SetLevel(logrus.WarnLevel)
		}

		// Defines trace level wanted by user
		if utils.Verbose = c.Bool("verbose"); utils.Verbose {
			logrus.SetLevel(logrus.InfoLevel)
			utils.Verbose = true
		}
		if utils.Debug = c.Bool("debug"); utils.Debug {
			if utils.Verbose {
				logrus.SetLevel(logrus.TraceLevel)
			} else {
				logrus.SetLevel(logrus.DebugLevel)
			}
		}

		return nil
	}

	app.After = func(c *cli.Context) error {
		cleanup(false)
		return nil
	}

	app.Commands = append(app.Commands, commands.NetworkCmd)
	sort.Sort(cli.CommandsByName(commands.NetworkCmd.Subcommands))

	app.Commands = append(app.Commands, commands.TenantCmd)
	sort.Sort(cli.CommandsByName(commands.TenantCmd.Subcommands))

	app.Commands = append(app.Commands, commands.HostCmd)
	sort.Sort(cli.CommandsByName(commands.HostCmd.Subcommands))

	app.Commands = append(app.Commands, commands.VolumeCmd)
	sort.Sort(cli.CommandsByName(commands.VolumeCmd.Subcommands))

	app.Commands = append(app.Commands, commands.SSHCmd)
	sort.Sort(cli.CommandsByName(commands.SSHCmd.Subcommands))

	app.Commands = append(app.Commands, commands.BucketCmd)
	sort.Sort(cli.CommandsByName(commands.BucketCmd.Subcommands))

	app.Commands = append(app.Commands, commands.ShareCmd)
	sort.Sort(cli.CommandsByName(commands.ShareCmd.Subcommands))

	app.Commands = append(app.Commands, commands.ImageCmd)
	sort.Sort(cli.CommandsByName(commands.ImageCmd.Subcommands))

	app.Commands = append(app.Commands, commands.TemplateCmd)
	sort.Sort(cli.CommandsByName(commands.TemplateCmd.Subcommands))

	app.Commands = append(app.Commands, commands.ClusterCommand)
	sort.Sort(cli.CommandsByName(commands.ClusterCommand.Subcommands))

	sort.Sort(cli.CommandsByName(app.Commands))

	// err := app.Run(os.Args)
	// if err != nil {
	// 	fmt.Println("Error Running App: " + err.Error())
	// }
	_ = app.Run(os.Args)

	cleanup(false)
}
