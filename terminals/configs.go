package terminals

type TermConfigs struct {
	RemoteAdd string
	UserName  string
	Password  string
	DockerConfig
}

type DockerConfig struct {
	Repo        string
	Image       string
	ContainerID string
}

type WindowSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
	X    uint16
	Y    uint16
}
