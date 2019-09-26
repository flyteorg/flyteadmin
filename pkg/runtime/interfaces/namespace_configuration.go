package interfaces

type NamespaceMappingConfig struct {
	Mapping string `json:mapping`
}

type NamespaceMappingConfiguration interface {
	GetNamespaceMappingConfig() string
}
