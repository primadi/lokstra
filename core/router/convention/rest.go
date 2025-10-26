package convention

// RESTConvention implements REST routing convention
type RESTConvention struct{}

func (c *RESTConvention) Name() ConventionType {
	return REST
}

func (c *RESTConvention) ResolveMethod(methodName string, resource string, resourcePlural string) (httpMethod string, pathTemplate string, found bool) {
	if resourcePlural == "" {
		resourcePlural = resource + "s"
	}

	// Map service methods to REST routes
	switch methodName {
	case "List", "GetAll", "FindAll":
		return "GET", "/" + resourcePlural, true

	case "Get", "Find", "GetByID", "FindByID":
		return "GET", "/" + resourcePlural + "/{id}", true

	case "Create", "Add", "Insert":
		return "POST", "/" + resourcePlural, true

	case "Update", "Modify", "Edit":
		return "PUT", "/" + resourcePlural + "/{id}", true

	case "Delete", "Remove":
		return "DELETE", "/" + resourcePlural + "/{id}", true

	case "Patch":
		return "PATCH", "/" + resourcePlural + "/{id}", true

	default:
		// Unknown method
		return "", "", false
	}
}

// Ensure RESTConvention implements Convention
var _ Convention = (*RESTConvention)(nil)
