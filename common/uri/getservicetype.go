package uri

import "strings"

func GetServiceTypeFromUri(serviceUri string) string {
	// Example: lokstra://package.Interface/instance
	// We need to extract "package.Interface" as the service type
	if !strings.HasPrefix(serviceUri, "lokstra://") {
		return ""
	}
	parts := strings.SplitN(serviceUri[len("lokstra://"):], "/", 2)
	if len(parts) < 1 {
		return ""
	}
	return parts[0]
}
