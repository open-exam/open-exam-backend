package shared

type Plugin struct {
	*ServerSide `json:"serverSide"`
}

type ServerSide struct {
	Dockerfile string `json:"dockerfile"`
	Context string `json:"context"`
}

func ParsePlugin(name string) (Plugin, error) {
	return Plugin{}, nil
}