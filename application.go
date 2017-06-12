package main

import (
	"fmt"
	"os"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	debug, jsonLogs, nonInteractive bool
	dockerHost, cfgFile, appName    string
)

type Application struct {
	Name, ShortDescription, LongDescription string
	commands                                Commands
	cmd                                     *cobra.Command
}

func (a *Application) RunWith(c Commands) {
	appName = a.Name
	c.forEach(func(name string, cmd *Command) error {
		a.add(name, cmd)
		return nil
	})
	if err := a.initCobra(); err != nil {
		os.Exit(1)
	}
}

func (a *Application) add(name string, cmd *Command) {
	cmd.Name = name
	if a.commands == nil {
		a.commands = make(Commands)
	}
	cmd.generate()
	a.commands[name] = cmd
}

func (a *Application) generate() {
	a.cmd = &cobra.Command{
		Use:   a.Name,
		Short: a.ShortDescription,
		Long:  a.LongDescription,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Setup logging
			if debug {
				log.SetLevel(log.DebugLevel)
			}

			if jsonLogs {
				log.SetFormatter(&log.JSONFormatter{})
			}
		},
	}
	a.initRootCmd(a.cmd)
}

func (a *Application) initCobra() error {
	a.generate()
	a.commands.forEach(func(name string, c *Command) error {
		a.cmd.AddCommand(c.cmd)
		return nil
	})
	if err := a.cmd.Execute(); err != nil {
		return err
	}
	return nil
}

func (a *Application) initRootCmd(rootCmd *cobra.Command) {
	cobra.OnInitialize(a.initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.%s.yaml)", appName))

	var dockerSocket string
	if runtime.GOOS == "windows" {
		dockerSocket = "npipe:////./pipe/docker_engine"
	} else {
		dockerSocket = "unix:///var/run/docker.sock"
	}
	rootCmd.PersistentFlags().StringVarP(&dockerHost, "docker-host", "H", dockerSocket, "URI of Docker Daemon")
	viper.BindPFlag("docker-host", rootCmd.PersistentFlags().Lookup("docker-host"))
	viper.SetDefault("docker-host", dockerSocket)

	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.SetDefault("debug", true)

	rootCmd.PersistentFlags().BoolVarP(&jsonLogs, "json", "j", false, "Log in json format")
	viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	viper.SetDefault("json", true)

	rootCmd.PersistentFlags().BoolVarP(&nonInteractive, "non-interactive", "N", false, "Do not use a tty, non-interactive output only")
	viper.BindPFlag("non-interactive", rootCmd.PersistentFlags().Lookup("non-interactive"))
	viper.SetDefault("non-interactive", false)
}

func (a *Application) initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(fmt.Sprintf(".%s", appName)) // name of config file (without extension)
	viper.AddConfigPath("$HOME")                     // adding home directory as first search path
	viper.AutomaticEnv()                             // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
