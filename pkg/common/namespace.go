package common

import "fmt"

type NamespaceMapping int

const namespaceFormat = "%s-%s"

const (
	NmProjectDomain NamespaceMapping = iota
	NmDomain        NamespaceMapping = iota
	NmProject       NamespaceMapping = iota
)

// GetNamespaceName returns kubernetes namespace name
func GetNamespaceName(mapping NamespaceMapping, project, domain string) string {
	switch mapping {
	case NmDomain:
		return domain
	case NmProject:
		return project
	case NmProjectDomain:
		fallthrough
	default:
		return fmt.Sprintf(namespaceFormat, project, domain)
	}
}
