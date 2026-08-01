package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &pluginDashboardResource{}
	_ resource.ResourceWithConfigure   = &pluginDashboardResource{}
	_ resource.ResourceWithImportState = &pluginDashboardResource{}
)

// NewPluginDashboardResource manages SigNoz V2 dashboards composed of
// WebComponent plugin panels (Lacrosse plugin host).
func NewPluginDashboardResource() resource.Resource {
	return &pluginDashboardResource{}
}

type pluginDashboardResource struct {
	client *client.Client
}

type pluginDashboardModel struct {
	ID          types.String      `tfsdk:"id"`
	Title       types.String      `tfsdk:"title"`
	Description types.String      `tfsdk:"description"`
	Tags        types.List        `tfsdk:"tags"`
	Pages       []pluginPageModel `tfsdk:"page"`
	// Rendered JSON payload (computed) for debugging / drift inspection.
	Document types.String `tfsdk:"document"`
}

type pluginPageModel struct {
	Name   types.String        `tfsdk:"name"`
	Panels []pluginPanelModel  `tfsdk:"plugin_panel"`
}

type pluginPanelModel struct {
	ID         types.String `tfsdk:"id"`
	Title      types.String `tfsdk:"title"`
	PluginID   types.String `tfsdk:"plugin_id"`
	Element    types.String `tfsdk:"element"`
	Attributes types.Map    `tfsdk:"attributes"`
	Properties types.String `tfsdk:"properties"` // JSON object string
	ConfigKeys types.List   `tfsdk:"config_keys"`
	X          types.Int64  `tfsdk:"x"`
	Y          types.Int64  `tfsdk:"y"`
	W          types.Int64  `tfsdk:"w"`
	H          types.Int64  `tfsdk:"h"`
}

func (r *pluginDashboardResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = SigNozPluginDashboard
}

func (r *pluginDashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		addErr(&resp.Diagnostics, fmt.Errorf("expected *client.Client, got %T", req.ProviderData), operationConfigure, SigNozPluginDashboard)
		return
	}
	r.client = c
}

func (r *pluginDashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	emptyList, _ := types.ListValue(types.StringType, []attr.Value{})

	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a SigNoz **V2** dashboard page layout populated with **WebComponent plugin panels** " +
			"(`signoz/WebComponentPanel`). Use this to manage Lacrosse plugin host dashboards as code " +
			"(e.g. `plugin_id = \"hello\"`, `element = \"lax-hello-wc\"`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Dashboard UUID from SigNoz.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Dashboard title shown in SigNoz UI.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Dashboard description.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(emptyList),
				MarkdownDescription: "Tags (e.g. lacrosse, plugins).",
			},
			"document": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Rendered V2 dashboard JSON document (for inspection).",
			},
		},
		Blocks: map[string]schema.Block{
			"page": schema.ListNestedBlock{
				MarkdownDescription: "Logical dashboard page / section. At least one page is required.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Page name (e.g. Overview, DevTime).",
						},
					},
					Blocks: map[string]schema.Block{
						"plugin_panel": schema.ListNestedBlock{
							MarkdownDescription: "WebComponent plugin panel on this page.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "Stable panel id (UUID). Auto-generated if omitted.",
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"title": schema.StringAttribute{
										Required: true,
									},
									"plugin_id": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Registered plugin id (Phase 0: `hello`).",
									},
									"element": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Custom element tag (must contain a hyphen), e.g. `lax-hello-wc`.",
									},
									"attributes": schema.MapAttribute{
										ElementType:         types.StringType,
										Optional:            true,
										MarkdownDescription: "HTML attributes on the custom element.",
									},
									"properties": schema.StringAttribute{
										Optional:            true,
										MarkdownDescription: "JSON object of JS properties for the element.",
									},
									"config_keys": schema.ListAttribute{
										ElementType:         types.StringType,
										Optional:            true,
										MarkdownDescription: "Config keys this panel instance needs resolved (Phase 2).",
									},
									"x": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(0),
									},
									"y": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(0),
									},
									"w": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(6),
									},
									"h": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(4),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *pluginDashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pluginDashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, doc, diags := planToV2(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateDashboardV2(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating plugin dashboard", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.Document = types.StringValue(doc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pluginDashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pluginDashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.client.GetDashboardV2(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err, operationRead, SigNozPluginDashboard)
		return
	}
	if got.Title != "" {
		state.Title = types.StringValue(got.Title)
	}
	if got.Description != "" {
		state.Description = types.StringValue(got.Description)
	}
	// Keep planned pages/panels as source of truth; remote round-trip of nested
	// layout is lossy across SigNoz versions. Refresh document snapshot only.
	if got.Data != nil {
		if b, err := json.MarshalIndent(got.Data, "", "  "); err == nil {
			state.Document = types.StringValue(string(b))
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *pluginDashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pluginDashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state pluginDashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, doc, diags := planToV2(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		id = plan.ID.ValueString()
	}
	if err := r.client.UpdateDashboardV2(ctx, id, payload); err != nil {
		resp.Diagnostics.AddError("Error updating plugin dashboard", err.Error())
		return
	}
	plan.ID = types.StringValue(id)
	plan.Document = types.StringValue(doc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pluginDashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pluginDashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteDashboardV2(ctx, state.ID.ValueString()); err != nil {
		// Fall back to v1 delete path if v2 not available.
		if err2 := r.client.DeleteDashboard(ctx, state.ID.ValueString()); err2 != nil {
			addErr(&resp.Diagnostics, err, operationDelete, SigNozPluginDashboard)
			return
		}
	}
}

func (r *pluginDashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func planToV2(ctx context.Context, plan pluginDashboardModel) (model.DashboardV2, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	pd := model.PluginDashboard{
		Title:       plan.Title.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		diags.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		pd.Tags = tags
	}

	if len(plan.Pages) == 0 {
		diags.AddError("Invalid plugin dashboard", "At least one page { ... } block is required.")
		return model.DashboardV2{}, "", diags
	}

	for _, page := range plan.Pages {
		pp := model.PluginPage{Name: page.Name.ValueString()}
		if len(page.Panels) == 0 {
			diags.AddError("Invalid page", fmt.Sprintf("page %q needs at least one plugin_panel", pp.Name))
			continue
		}
		for _, panel := range page.Panels {
			attrs := map[string]string{}
			if !panel.Attributes.IsNull() && !panel.Attributes.IsUnknown() {
				diags.Append(panel.Attributes.ElementsAs(ctx, &attrs, false)...)
			}
			props := map[string]interface{}{}
			if !panel.Properties.IsNull() && !panel.Properties.IsUnknown() && panel.Properties.ValueString() != "" {
				if err := json.Unmarshal([]byte(panel.Properties.ValueString()), &props); err != nil {
					diags.AddError("Invalid properties JSON", err.Error())
				}
			}
			var cfgKeys []string
			if !panel.ConfigKeys.IsNull() && !panel.ConfigKeys.IsUnknown() {
				diags.Append(panel.ConfigKeys.ElementsAs(ctx, &cfgKeys, false)...)
			}
			pp.Panels = append(pp.Panels, model.PluginPanel{
				ID:         panel.ID.ValueString(),
				Title:      panel.Title.ValueString(),
				PluginID:   panel.PluginID.ValueString(),
				Element:    panel.Element.ValueString(),
				Attributes: attrs,
				Properties: props,
				ConfigKeys: cfgKeys,
				X:          int(panel.X.ValueInt64()),
				Y:          int(panel.Y.ValueInt64()),
				W:          int(panel.W.ValueInt64()),
				H:          int(panel.H.ValueInt64()),
			})
		}
		pd.Pages = append(pd.Pages, pp)
	}
	if diags.HasError() {
		return model.DashboardV2{}, "", diags
	}

	v2 := pd.ToDashboardV2()
	doc, err := v2.ToJSON()
	if err != nil {
		diags.AddError("Failed to marshal dashboard document", err.Error())
		return model.DashboardV2{}, "", diags
	}
	tflog.Debug(ctx, "Built plugin dashboard document", map[string]any{"title": v2.Title})
	return v2, doc, diags
}
