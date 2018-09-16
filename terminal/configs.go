package terminal

type TermConfigs struct {
	RemoteAdd string
	DockerConfig
}

type DockerConfig struct {
	Repo        string
	Image       string
	ContainerID string
}
