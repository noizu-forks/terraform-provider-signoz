package schemas

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// The UseStateForUnknown modifiers on server-assigned / server-defaulted
// fields are the fix for perpetual known-after-apply churn on route policies
// and email channels. Codegen regenerates these files, so pin the modifiers:
// if a regen drops them this fails.
func TestRoutePolicySchemaKeepsUseStateForUnknown(t *testing.T) {
	t.Parallel()

	attrs := RoutePolicyResourceSchema(context.Background()).Attributes

	for _, name := range []string{"id", "description", "kind"} {
		a, ok := attrs[name].(schema.StringAttribute)
		if !ok {
			t.Fatalf("attribute %q is not a StringAttribute", name)
		}

		if !hasUseStateForUnknown(a.PlanModifiers) {
			t.Errorf("route_policy %q lost UseStateForUnknown", name)
		}
	}

	tags, ok := attrs["tags"].(schema.ListAttribute)
	if !ok {
		t.Fatal("attribute tags is not a ListAttribute")
	}

	if !hasUseStateForUnknown(tags.PlanModifiers) {
		t.Error("route_policy tags lost UseStateForUnknown")
	}
}

func TestNotificationChannelEmailSpecKeepsUseStateForUnknown(t *testing.T) {
	t.Parallel()

	spec := notificationChannelEmailSpec(t)

	html, ok := spec.Attributes["html"].(schema.StringAttribute)
	if !ok {
		t.Fatal("config.email.spec.html is not a StringAttribute")
	}

	if !hasUseStateForUnknown(html.PlanModifiers) {
		t.Error("email spec html lost UseStateForUnknown")
	}

	headers, ok := spec.Attributes["headers"].(schema.MapAttribute)
	if !ok {
		t.Fatal("config.email.spec.headers is not a MapAttribute")
	}

	if !hasUseStateForUnknown(headers.PlanModifiers) {
		t.Error("email spec headers lost UseStateForUnknown")
	}
}

func notificationChannelEmailSpec(t *testing.T) schema.SingleNestedAttribute {
	t.Helper()

	root := NotificationChannelResourceSchema(context.Background()).Attributes

	config, ok := root["config"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("config is not a SingleNestedAttribute")
	}

	email, ok := config.Attributes["email"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("config.email is not a SingleNestedAttribute")
	}

	spec, ok := email.Attributes["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("config.email.spec is not a SingleNestedAttribute")
	}

	return spec
}

// hasUseStateForUnknown reports whether any modifier is the framework's
// UseStateForUnknown (unexported concrete type, matched by name).
func hasUseStateForUnknown[T any](mods []T) bool {
	for _, m := range mods {
		if strings.Contains(fmt.Sprintf("%T", m), "useStateForUnknownModifier") {
			return true
		}
	}

	return false
}
