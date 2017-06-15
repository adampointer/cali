# cali
--
    import "github.com/adampointer/cali"


## Usage

```go
const (
	EXIT_CODE_RUNTIME_ERROR = 1
	EXIT_CODE_API_ERROR     = 2
)
```

#### func  Cli

```go
func Cli(n string) *cli
```
Cli returns a brand new cli

#### type CreateResponse

```go
type CreateResponse struct {
	Id             string         `json:"id"`
	Status         string         `json:"status"`
	ProgressDetail ProgressDetail `json:"progressDetail"`
	Progress       string         `json:"progress,omitempty"`
}
```

CreateResponse is the response from Docker API when pulling an image

#### type DockerClient

```go
type DockerClient struct {
	Cli      *client.Client
	HostConf *container.HostConfig
	NetConf  *network.NetworkingConfig
	Conf     *container.Config
}
```

DockerClient is a slimmed down implementation of the docker cli

#### func  GetDockerClient

```go
func GetDockerClient() *DockerClient
```
GetDockerClient creates or returns the DockerClient singleton

#### func  NewDockerClient

```go
func NewDockerClient(cli *client.Client) *DockerClient
```
NewDockerClient returns a new DockerClient initialised with the API object

#### func (*DockerClient) AddBind

```go
func (c *DockerClient) AddBind(bnd string)
```
AddBind adds a bind mount to the HostConfig

#### func (*DockerClient) AddBinds

```go
func (c *DockerClient) AddBinds(bnds []string)
```
AddBinds adds multiple bind mounts to the HostConfig

#### func (*DockerClient) AddEnv

```go
func (c *DockerClient) AddEnv(env string)
```
AddEnvs adds an environment variable to the HostConfig

#### func (*DockerClient) AddEnvs

```go
func (c *DockerClient) AddEnvs(envs []string)
```
AddEnvs adds multiple envs to the HostConfig

#### func (*DockerClient) BindFromGit

```go
func (c *DockerClient) BindFromGit(cfg *GitCheckoutConfig, noGit func()) error
```
BindFromGit creates a data container with a git clone inside and mounts its
volumes inside your app container If there is no valid Git repo set in config,
the noGit callback function will be executed instead

#### func (*DockerClient) ContainerExists

```go
func (c *DockerClient) ContainerExists(name string) bool
```
ContainerExists determines if the container with this name exist

#### func (*DockerClient) DeleteContainer

```go
func (c *DockerClient) DeleteContainer(id string) error
```
DeleteContainer - Delete a container

#### func (*DockerClient) Git

```go
func (c *DockerClient) Git() *Git
```
Git returns a new instance

#### func (*DockerClient) ImageExists

```go
func (c *DockerClient) ImageExists(image string) bool
```
ImageExists determines if an image exist locally

#### func (*DockerClient) Privileged

```go
func (c *DockerClient) Privileged(p bool)
```
Privileged sets whether the container should run as privileged

#### func (*DockerClient) PullImage

```go
func (c *DockerClient) PullImage(image string) error
```
PullImage - Pull an image locally

#### func (*DockerClient) SetBinds

```go
func (c *DockerClient) SetBinds(bnds []string)
```
SetBinds sets the bind mounts in the HostConfig

#### func (*DockerClient) SetCmd

```go
func (c *DockerClient) SetCmd(cmd []string)
```
SetCmd sets the command to run in the container

#### func (*DockerClient) SetConf

```go
func (c *DockerClient) SetConf(co *container.Config)
```
SetConf sets the container.Config struct for the new container

#### func (*DockerClient) SetDefaults

```go
func (c *DockerClient) SetDefaults()
```
SetDefaults sets container, host and net configs to defaults. Called when
instantiating a new client or can be called manually at any time to reset API
configs back to empty defaults

#### func (*DockerClient) SetEnvs

```go
func (c *DockerClient) SetEnvs(envs []string)
```
SetEnvs sets the environment variables in the HostConfig

#### func (*DockerClient) SetHostConf

```go
func (c *DockerClient) SetHostConf(h *container.HostConfig)
```
SetHostConf sets the container.HostConfig struct for the new container

#### func (*DockerClient) SetImage

```go
func (c *DockerClient) SetImage(img string)
```
SetImage sets the image in HostConfig

#### func (*DockerClient) SetNetConf

```go
func (c *DockerClient) SetNetConf(n *network.NetworkingConfig)
```
SetNetConf sets the network.NetworkingConfig struct for the new container

#### func (*DockerClient) SetWorkDir

```go
func (c *DockerClient) SetWorkDir(wd string)
```
SetWorkDir sets the working directory of the container

#### func (*DockerClient) StartContainer

```go
func (c *DockerClient) StartContainer(rm bool, name string) (string, error)
```
StartContainer will create and start a container with logs and optional cleanup

#### type Event

```go
type Event struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}
```

Event holds the json structure for Docker API events

#### type Git

```go
type Git struct {
}
```

Git is used to interact with containerised git

#### func (*Git) Checkout

```go
func (g *Git) Checkout(cfg *GitCheckoutConfig) (string, error)
```
GitCheckout will create and start a container, checkout repo and leave container
stopped so volume can be imported

#### func (*Git) Pull

```go
func (g *Git) Pull(name, relPath string) (string, error)
```

#### type GitCheckoutConfig

```go
type GitCheckoutConfig struct {
	Repo, Branch, RelPath string
}
```

GitCheckoutConfig is input for Git.Checkout

#### type ProgressDetail

```go
type ProgressDetail struct {
	Current int `json:"current,omitempty"`
	Total   int `json:"total,omitempty"`
}
```

ProgressDetail records the progress achieved downloading an image

#### type Task

```go
type Task struct {
	*DockerClient
}
```

Task is the action performed when it's parent command is run

#### func (*Task) SetFunc

```go
func (t *Task) SetFunc(f TaskFunc)
```
SetFunc sets the TaskFunc which is run when the parent command is run if this is
left, the defaultTaskFunc will be executed instead

#### func (*Task) SetInitFunc

```go
func (t *Task) SetInitFunc(f TaskFunc)
```
SetInitFunc sets the TaskFunc which is executed before the main TaskFunc. It's
pupose is to do any setup of the DockerClient which depends on command line args
for example

#### type TaskFunc

```go
type TaskFunc func(t *Task, args []string)
```

TaskFunc is a function executed by a Task when the command the Task belongs to
is run
