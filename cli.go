package cali

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

// TaskFunc is a function executed by a Task when the command the Task belongs to is run
type TaskFunc func(t *Task, args []string)

// defaultTaskFunc is the TaskFunc which is executed unless a custom TaskFunc is
// attached to the Task
var defaultTaskFunc TaskFunc = func(t *Task, args []string) {
	if err := t.InitDocker(); err != nil {
		log.Fatalf("Error initialising Docker: %s", err)
	}
	if _, err := t.StartContainer(true, ""); err != nil {
		log.Fatalf("Error executing task: %s", err)
	}
}

// Task is the action performed when it's parent command is run
type Task struct {
	f, init TaskFunc
	*DockerClient
}

// SetFunc sets the TaskFunc which is run when the parent command is run
// if this is left, the defaultTaskFunc will be executed instead
func (t *Task) SetFunc(f TaskFunc) {
	t.f = f
}

// SetInitFunc sets the TaskFunc which is executed before the main TaskFunc. It's
// pupose is to do any setup of the DockerClient which depends on command line args
// for example
func (t *Task) SetInitFunc(f TaskFunc) {
	t.init = f
}

// cobraFunc represents the function signiture which cobra uses for it's Run, PreRun, PostRun etc.
type cobraFunc func(cmd *cobra.Command, args []string)

// command is the actual command run by the cli and essentially just wraps cobra.Command and
// has an associated Task
type command struct {
	name    string
	RunTask *Task
	cobra   *cobra.Command
}

// newCommand returns an freshly initialised command
func newCommand(n string) *command {
	return &command{
		name:  n,
		cobra: &cobra.Command{Use: n},
	}
}

// SetShort sets the short description of the command
func (c *command) SetShort(s string) {
	c.cobra.Short = s
}

// SetLong sets the long description of the command
func (c *command) SetLong(l string) {
	c.cobra.Long = l
}

// setPreRun sets the cobra.Command.PreRun function
func (c *command) setPreRun(f cobraFunc) {
	c.cobra.PreRun = f
}

// setRun sets the cobra.Command.Run function
func (c *command) setRun(f cobraFunc) {
	c.cobra.Run = f
}

// Task is something executed by a command
func (c *command) Task(def interface{}) *Task {
	t := &Task{DockerClient: NewDockerClient()}

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

// Flags returns the FlagSet for the command and is used to set new flags for the command
func (c *command) Flags() *flag.FlagSet {
	return c.cobra.PersistentFlags()
}

// BindFlags needs to be called after all flags for a command have been defined
func (c *command) BindFlags() {
	c.Flags().VisitAll(func(f *flag.Flag) {
		myFlags.BindPFlag(f.Name, f)
		myFlags.SetDefault(f.Name, f.DefValue)
	})
}

// commands is a set of commands
type commands map[string]*command

// cli is the application itself
type cli struct {
	name    string
	cfgFile *string
	cmds    commands
	*command
}

// Cli returns a brand new cli
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

// Command returns a brand new command attached to it's parent cli
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

// FlagValues returns the wrapped viper object allowing the API consumer to use methods
// like GetString to get values from config
func (c *cli) FlagValues() *viper.Viper {
	return myFlags
}

// initFlags does the intial setup of the root command's persistent flags
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

// initConfig does the initial setup of viper
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

// Start the fans please!
func (c *cli) Start() {
	c.initFlags()
	cobra.OnInitialize(c.initConfig)

	if err := c.cobra.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(EXIT_CODE_RUNTIME_ERROR)
	}
}
