package main

import (
	"fmt"
	"os"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	EXIT_CODE_RUNTIME_ERROR = 1
	EXIT_CODE_API_ERROR     = 2
)

var (
	debug, jsonLogs, nonInteractive bool
	dockerHost                      string
	myFlags                         *viper.Viper
)

type TaskFunc func(t *Task, args []string)

var defaultTaskFunc TaskFunc = func(t *Task, args []string) {
	cli := GetDockerClient()
	cli.SetImage(t.image)
	cli.SetCmd(t.cmd)

	if _, err := cli.StartContainer(true, ""); err != nil {
		log.Fatalf("Error executing task: %s", err)
	}
}

type Task struct {
	f, init TaskFunc
	cmd     []string
	image   string
}

func (t *Task) SetCmd(cmd []string) {
	t.cmd = cmd
}

func (t *Task) SetImage(image string) {
	t.image = image
}

func (t *Task) SetFunc(f TaskFunc) {
	t.f = f
}

func (t *Task) SetInitFunc(f TaskFunc) {
	t.init = f
}

type cobraFunc func(cmd *cobra.Command, args []string)

type command struct {
	name    string
	RunTask *Task
	cobra   *cobra.Command
}

func newCommand(n string) *command {
	return &command{
		name:  n,
		cobra: &cobra.Command{Use: n},
	}
}

func (c *command) SetShort(s string) {
	c.cobra.Short = s
}

func (c *command) SetLong(l string) {
	c.cobra.Long = l
}

func (c *command) setPreRun(f cobraFunc) {
	c.cobra.PreRun = f
}

func (c *command) setRun(f cobraFunc) {
	c.cobra.Run = f
}

func (c *command) Task(def interface{}) *Task {
	t := &Task{}

	switch d := def.(type) {
	case string:
		t.SetImage(d)
		t.SetFunc(defaultTaskFunc)
	case TaskFunc:
		t.SetFunc(d)
	default:
		// Slightly unidiomatic to blow up here rather than return an error
		// choosing to so as to keep the API uncluttered and also if you get here it's
		// an implementation error rather than a runtime error.
		fmt.Println("Unknown Task type. Must either be an image (string) or a TaskFunc")
		os.Exit(EXIT_CODE_API_ERROR)
	}
	c.RunTask = t
	return t
}

func (c *command) Flags() *flag.FlagSet {
	return c.cobra.PersistentFlags()
}

func (c *command) BindFlags() {
	c.Flags().VisitAll(func(f *flag.Flag) {
		myFlags.BindPFlag(f.Name, f)
		myFlags.SetDefault(f.Name, f.DefValue)
	})
}

type commands map[string]*command

type cli struct {
	name    string
	cfgFile *string
	cmds    commands
	*command
}

func Cli(n string) *cli {
	c := cli{
		name:    n,
		cmds:    make(commands),
		command: newCommand(n),
	}
	c.cobra.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if debug {
			log.SetLevel(log.DebugLevel)
		}

		if jsonLogs {
			log.SetFormatter(&log.JSONFormatter{})
		}
	}
	myFlags = viper.New()
	return &c
}

func (c *cli) Command(n string) *command {
	cmd := newCommand(n)
	c.cmds[n] = cmd
	cmd.setPreRun(func(c *cobra.Command, args []string) {
		cmd.RunTask.init(cmd.RunTask, args)
	})
	cmd.setRun(func(c *cobra.Command, args []string) {
		cmd.RunTask.f(cmd.RunTask, args)
	})
	c.cobra.AddCommand(cmd.cobra)
	return cmd
}

func (c *cli) FlagValues() *viper.Viper {
	return myFlags
}

func (c *cli) initFlags() {
	var cfg string
	txt := fmt.Sprintf("config file (default is $HOME/.%s.yaml)", c.name)
	c.cobra.PersistentFlags().StringVar(&cfg, "config", "", txt)
	c.cfgFile = &cfg

	var dockerSocket string
	if runtime.GOOS == "windows" {
		dockerSocket = "npipe:////./pipe/docker_engine"
	} else {
		dockerSocket = "unix:///var/run/docker.sock"
	}
	c.Flags().StringVarP(&dockerHost, "docker-host", "H", dockerSocket, "URI of Docker Daemon")
	myFlags.BindPFlag("docker-host", c.Flags().Lookup("docker-host"))
	myFlags.SetDefault("docker-host", dockerSocket)

	c.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
	myFlags.BindPFlag("debug", c.Flags().Lookup("debug"))
	myFlags.SetDefault("debug", true)

	c.Flags().BoolVarP(&jsonLogs, "json", "j", false, "Log in json format")
	myFlags.BindPFlag("json", c.Flags().Lookup("json"))
	myFlags.SetDefault("json", true)

	c.Flags().BoolVarP(&nonInteractive, "non-interactive", "N", false, "Do not create a tty for Docker")
	myFlags.BindPFlag("non-interactive", c.Flags().Lookup("non-interactive"))
	myFlags.SetDefault("non-interactive", false)
}

func (c *cli) initConfig() {
	if *c.cfgFile != "" {
		myFlags.SetConfigFile(*c.cfgFile)
	} else {
		myFlags.SetConfigName(fmt.Sprintf(".%s", c.name))
		myFlags.AddConfigPath("$HOME")
	}
	myFlags.AutomaticEnv()

	// If a config file is found, read it in
	if err := myFlags.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", myFlags.ConfigFileUsed())
	}
}

func (c *cli) initLogs() {
}

func (c *cli) Start() {
	c.initFlags()
	cobra.OnInitialize(c.initConfig)

	if err := c.cobra.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(EXIT_CODE_RUNTIME_ERROR)
	}
}
