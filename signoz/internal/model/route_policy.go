package model

import (
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoutePolicy model.
type RoutePolicy struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description,omitempty"`
	Enabled        bool           `json:"enabled"`
	Matchers       []RouteMatcher `json:"matchers"`
	ChannelIDs     []string       `json:"channel_ids"`
	Continue       bool           `json:"continue"`
	GroupWait      string         `json:"group_wait,omitempty"`
	GroupInterval  string         `json:"group_interval,omitempty"`
	RepeatInterval string         `json:"repeat_interval,omitempty"`
	CreatedAt      string         `json:"created_at,omitempty"`
	CreatedBy      string         `json:"created_by,omitempty"`
	UpdatedAt      string         `json:"updated_at,omitempty"`
	UpdatedBy      string         `json:"updated_by,omitempty"`
}

// RouteMatcher model.
type RouteMatcher struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	MatchType string `json:"match_type"`
}

// MatcherAttrTypes returns the attribute types for a single matcher object.
func MatcherAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.MatchName:  types.StringType,
		attr.MatchValue: types.StringType,
		attr.MatchType:  types.StringType,
	}
}

// MatchersToTerraform converts matchers to a Terraform list value.
func (r RoutePolicy) MatchersToTerraform() (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: MatcherAttrTypes()}

	if len(r.Matchers) == 0 {
		return types.ListValueMust(elemType, []tfattr.Value{}), nil
	}

	elements := make([]tfattr.Value, 0, len(r.Matchers))
	for _, m := range r.Matchers {
		obj, diags := types.ObjectValue(MatcherAttrTypes(), map[string]tfattr.Value{
			attr.MatchName:  types.StringValue(m.Name),
			attr.MatchValue: types.StringValue(m.Value),
			attr.MatchType:  types.StringValue(m.MatchType),
		})
		if diags.HasError() {
			return types.ListNull(elemType), diags
		}
		elements = append(elements, obj)
	}

	return types.ListValue(elemType, elements)
}

// ChannelIDsToTerraform converts channel IDs to a Terraform list value.
func (r RoutePolicy) ChannelIDsToTerraform() (types.List, diag.Diagnostics) {
	channelIDs := utils.Map(r.ChannelIDs, func(value string) tfattr.Value {
		return types.StringValue(value)
	})

	return types.ListValue(types.StringType, channelIDs)
}

// SetMatchers sets matchers from a Terraform list value.
func (r *RoutePolicy) SetMatchers(tfMatchers types.List) {
	matchers := make([]RouteMatcher, 0, len(tfMatchers.Elements()))
	for _, elem := range tfMatchers.Elements() {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		matcher := RouteMatcher{}
		if v, ok := attrs[attr.MatchName]; ok {
			if s, ok2 := v.(types.String); ok2 && !utils.IsNullOrUnknown(s) {
				matcher.Name = s.ValueString()
			}
		}
		if v, ok := attrs[attr.MatchValue]; ok {
			if s, ok2 := v.(types.String); ok2 && !utils.IsNullOrUnknown(s) {
				matcher.Value = s.ValueString()
			}
		}
		if v, ok := attrs[attr.MatchType]; ok {
			if s, ok2 := v.(types.String); ok2 && !utils.IsNullOrUnknown(s) {
				matcher.MatchType = s.ValueString()
			}
		}
		matchers = append(matchers, matcher)
	}
	r.Matchers = matchers
}

// SetChannelIDs sets channel IDs from a Terraform list value.
func (r *RoutePolicy) SetChannelIDs(tfChannelIDs types.List) {
	channelIDs := utils.Map(tfChannelIDs.Elements(), func(value tfattr.Value) string {
		return strings.Trim(value.String(), "\"")
	})
	r.ChannelIDs = channelIDs
}
