package annotation

import (
	"testing"
)

func TestParseUserServiceFile(t *testing.T) {
	filePath := "../../project_templates/02_app_framework/03_enterprise_router_service/modules/user/application/user_service.go"

	annotations, err := parseFileAnnotations(filePath)
	if err != nil {
		t.Fatalf("parseFileAnnotations() error = %v", err)
	}

	t.Logf("Found %d annotations", len(annotations))
	for i, ann := range annotations {
		t.Logf("[%d] @%s on line %d, target=%s", i, ann.Name, ann.Line, ann.TargetName)
		t.Logf("    Args: %#v", ann.Args)
		t.Logf("    PositionalArgs: %#v", ann.PositionalArgs)

		if ann.Name == "Route" {
			// Test readArgs
			if args, err := ann.readArgs("route", "middlewares"); err == nil {
				t.Logf("    readArgs result: %#v", args)
			} else {
				t.Logf("    readArgs error: %v", err)
			}
		}
	}
}
