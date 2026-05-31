package resource

import (
	"context"
	"fmt"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &routePolicyResource{}
	_ resource.ResourceWithConfigure   = &routePolicyResource{}
	_ resource.ResourceWithImportState = &routePolicyResource{}
)

// NewRoutePolicyResource is a helper function to simplify the provider implementation.
func NewRoutePolicyResource() resource.Resource {
	return &routePolicyResource{}
}

// routePolicyResource is the resource implementation.
type routePolicyResource struct {
	client *client.Client
}

// routePolicyResourceModel maps the resource schema data.
type routePolicyResourceModel struct {
	ID               types.String `tfsdk:"id"`
	RouteName        types.String `tfsdk:"route_name"`
	Description      types.String `tfsdk:"description"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Matchers         types.List   `tfsdk:"matchers"`
	ChannelIDs       types.List   `tfsdk:"channel_ids"`
	ContinueMatching types.Bool   `tfsdk:"continue_matching"`
	GroupWait        types.String `tfsdk:"group_wait"`
	GroupInterval    types.String `tfsdk:"group_interval"`
	RepeatInterval   types.String `tfsdk:"repeat_interval"`
	CreatedAt        types.String `tfsdk:"created_at"`
	CreatedBy        types.String `tfsdk:"created_by"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	UpdatedBy        types.String `tfsdk:"updated_by"`
}

// Configure adds the provider configured client to the resource.
func (r *routePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *client.Client, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			operationConfigure, "signoz_route_policy",
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *routePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "signoz_route_policy"
}

// Schema defines the schema for the resource.
func (r *routePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages route policy resources in SigNoz.",
		Attributes: map[string]schema.Attribute{
			attr.RouteName: schema.StringAttribute{
				Required:    true,
				Description: "Name of the route policy.",
			},
			attr.Description: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the route policy.",
				Default:     stringdefault.StaticString(""),
			},
			attr.Enabled: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the route policy is enabled. By default, it is true.",
				Default:     booldefault.StaticBool(true),
			},
			attr.Matchers: schema.ListNestedAttribute{
				Required:    true,
				Description: "List of matchers for the route policy.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.MatchName: schema.StringAttribute{
							Required:    true,
							Description: "Name of the matcher field (e.g., severity, service).",
						},
						attr.MatchValue: schema.StringAttribute{
							Required:    true,
							Description: "Value to match against.",
						},
						attr.MatchType: schema.StringAttribute{
							Required:    true,
							Description: "Type of match to perform (e.g., equals, regex).",
						},
					},
				},
			},
			attr.ChannelIDs: schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of notification channel IDs to route alerts to.",
			},
			attr.ContinueMatching: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to continue matching subsequent route policies after this one matches. By default, it is false.",
				Default:     booldefault.StaticBool(false),
			},
			attr.GroupWait: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How long to wait before sending a notification about new alerts that are added to a group. By default, it is 30s.",
				Default:     stringdefault.StaticString("30s"),
			},
			attr.GroupInterval: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How long to wait before sending a notification about new alerts added to an already firing group. By default, it is 5m.",
				Default:     stringdefault.StaticString("5m"),
			},
			attr.RepeatInterval: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How long to wait before re-sending a notification. By default, it is 4h.",
				Default:     stringdefault.StaticString("4h"),
			},
			// computed.
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated unique ID for the route policy.",
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the route policy.",
			},
			attr.CreatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Creator of the route policy.",
			},
			attr.UpdatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the route policy.",
			},
			attr.UpdatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Last updater of the route policy.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *routePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan routePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body.
	payload := &model.RoutePolicy{
		Name:           plan.RouteName.ValueString(),
		Description:    plan.Description.ValueString(),
		Enabled:        plan.Enabled.ValueBool(),
		Continue:       plan.ContinueMatching.ValueBool(),
		GroupWait:      plan.GroupWait.ValueString(),
		GroupInterval:  plan.GroupInterval.ValueString(),
		RepeatInterval: plan.RepeatInterval.ValueString(),
	}

	payload.SetMatchers(plan.Matchers)
	payload.SetChannelIDs(plan.ChannelIDs)

	tflog.Debug(ctx, "Creating route policy", map[string]any{"routePolicy": payload})

	// Create new route policy.
	routePolicy, err := r.client.CreateRoutePolicy(ctx, payload)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, "signoz_route_policy")
		return
	}

	tflog.Debug(ctx, "Created route policy", map[string]any{"routePolicy": routePolicy})

	// Map response to schema and populate Computed attributes.
	plan.ID = types.StringValue(routePolicy.ID)
	plan.RouteName = types.StringValue(routePolicy.Name)
	plan.Description = types.StringValue(routePolicy.Description)
	plan.Enabled = types.BoolValue(routePolicy.Enabled)
	plan.ContinueMatching = types.BoolValue(routePolicy.Continue)
	plan.GroupWait = types.StringValue(routePolicy.GroupWait)
	plan.GroupInterval = types.StringValue(routePolicy.GroupInterval)
	plan.RepeatInterval = types.StringValue(routePolicy.RepeatInterval)
	plan.CreatedAt = types.StringValue(routePolicy.CreatedAt)
	plan.CreatedBy = types.StringValue(routePolicy.CreatedBy)
	plan.UpdatedAt = types.StringValue(routePolicy.UpdatedAt)
	plan.UpdatedBy = types.StringValue(routePolicy.UpdatedBy)

	var diagMatchers, diagChannelIDs diag.Diagnostics
	plan.Matchers, diagMatchers = routePolicy.MatchersToTerraform()
	resp.Diagnostics.Append(diagMatchers...)

	plan.ChannelIDs, diagChannelIDs = routePolicy.ChannelIDsToTerraform()
	resp.Diagnostics.Append(diagChannelIDs...)

	// Set state to populated data.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *routePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var state routePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading route policy", map[string]any{"id": state.ID.ValueString()})

	// Get refreshed route policy from SigNoz.
	routePolicy, err := r.client.GetRoutePolicy(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, "signoz_route_policy")
		return
	}

	// Overwrite items with refreshed state.
	state.ID = types.StringValue(routePolicy.ID)
	state.RouteName = types.StringValue(routePolicy.Name)
	state.Description = types.StringValue(routePolicy.Description)
	state.Enabled = types.BoolValue(routePolicy.Enabled)
	state.ContinueMatching = types.BoolValue(routePolicy.Continue)
	state.GroupWait = types.StringValue(routePolicy.GroupWait)
	state.GroupInterval = types.StringValue(routePolicy.GroupInterval)
	state.RepeatInterval = types.StringValue(routePolicy.RepeatInterval)
	state.CreatedAt = types.StringValue(routePolicy.CreatedAt)
	state.CreatedBy = types.StringValue(routePolicy.CreatedBy)
	state.UpdatedAt = types.StringValue(routePolicy.UpdatedAt)
	state.UpdatedBy = types.StringValue(routePolicy.UpdatedBy)

	var diagMatchers, diagChannelIDs diag.Diagnostics
	state.Matchers, diagMatchers = routePolicy.MatchersToTerraform()
	resp.Diagnostics.Append(diagMatchers...)

	state.ChannelIDs, diagChannelIDs = routePolicy.ChannelIDsToTerraform()
	resp.Diagnostics.Append(diagChannelIDs...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *routePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan, state routePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan.
	payload := &model.RoutePolicy{
		ID:             state.ID.ValueString(),
		Name:           plan.RouteName.ValueString(),
		Description:    plan.Description.ValueString(),
		Enabled:        plan.Enabled.ValueBool(),
		Continue:       plan.ContinueMatching.ValueBool(),
		GroupWait:      plan.GroupWait.ValueString(),
		GroupInterval:  plan.GroupInterval.ValueString(),
		RepeatInterval: plan.RepeatInterval.ValueString(),
		CreatedAt:      state.CreatedAt.ValueString(),
		CreatedBy:      state.CreatedBy.ValueString(),
		UpdatedAt:      state.UpdatedAt.ValueString(),
		UpdatedBy:      state.UpdatedBy.ValueString(),
	}

	payload.SetMatchers(plan.Matchers)
	payload.SetChannelIDs(plan.ChannelIDs)

	tflog.Debug(ctx, "Updating route policy", map[string]any{"routePolicy": payload})

	// Update existing route policy.
	err := r.client.UpdateRoutePolicy(ctx, state.ID.ValueString(), payload)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, "signoz_route_policy")
		return
	}

	// Fetch updated route policy.
	routePolicy, err := r.client.GetRoutePolicy(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, "signoz_route_policy")
		return
	}

	// Overwrite items with refreshed state.
	plan.ID = types.StringValue(routePolicy.ID)
	plan.RouteName = types.StringValue(routePolicy.Name)
	plan.Description = types.StringValue(routePolicy.Description)
	plan.Enabled = types.BoolValue(routePolicy.Enabled)
	plan.ContinueMatching = types.BoolValue(routePolicy.Continue)
	plan.GroupWait = types.StringValue(routePolicy.GroupWait)
	plan.GroupInterval = types.StringValue(routePolicy.GroupInterval)
	plan.RepeatInterval = types.StringValue(routePolicy.RepeatInterval)
	plan.CreatedAt = types.StringValue(routePolicy.CreatedAt)
	plan.CreatedBy = types.StringValue(routePolicy.CreatedBy)
	plan.UpdatedAt = types.StringValue(routePolicy.UpdatedAt)
	plan.UpdatedBy = types.StringValue(routePolicy.UpdatedBy)

	var diagMatchers, diagChannelIDs diag.Diagnostics
	plan.Matchers, diagMatchers = routePolicy.MatchersToTerraform()
	resp.Diagnostics.Append(diagMatchers...)

	plan.ChannelIDs, diagChannelIDs = routePolicy.ChannelIDsToTerraform()
	resp.Diagnostics.Append(diagChannelIDs...)

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *routePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state routePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing route policy.
	err := r.client.DeleteRoutePolicy(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationDelete, "signoz_route_policy")
		return
	}
}

// ImportState imports Terraform state into the resource.
func (r *routePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
