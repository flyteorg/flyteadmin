package common

import "fmt"

type NamespaceMapping int

const namespaceFormat = "%s-%s"
const namespaceVariable = "namespace"

const (
	ProjectDomain NamespaceMapping = iota
	Domain        NamespaceMapping = iota
)

// GetNamespaceName returns kubernetes namespace name
func GetNamespaceName(mapping NamespaceMapping, project, domain string) string {
	switch mapping {
	case Domain:
		return domain
	case ProjectDomain:
		return fmt.Sprintf(namespaceFormat, project, domain)
	default:
		return fmt.Sprintf(namespaceFormat, project, domain)
	}
}
