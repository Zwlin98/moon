package cluster

type ClusterConfig interface {
	GetNodes() map[string]string
	NodeInfo(string) string
}

type DefaultConfig map[string]string

func (c DefaultConfig) GetNodes() map[string]string {
	return c
}

func (c DefaultConfig) NodeInfo(node string) string {
	return c[node]
}
