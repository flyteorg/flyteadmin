package common

import "fmt"

const namespaceFormat = "%s-%s"
const namespaceVariable = "namespace"
const domainVariable = "domain"

// Return kubernetes namespace name
func GetNamespaceName(mapping, project, domain string) string {
	if mapping == domainVariable {
		return domain
	}
	return fmt.Sprintf(namespaceFormat, project, domain)
}
