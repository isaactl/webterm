package main

type Container struct {
    Image string
    ID    string
    Name  string
}

func (c *Container) CreateContainer() error {
    return nil
}

func (c *Container) PullImage(image string) error {
    return nil
}

func (c *Container) Start() error {
    return nil
}

func (c *Container) Stop() error {
    return nil
}