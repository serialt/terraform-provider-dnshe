package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serialt/terraform-provider-dnshe/dnshe"
)

var _ resource.Resource = &subdomainResource{}

type subdomainResource struct{ client *dnshe.Client }

type subdomainResourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Subdomain  types.String `tfsdk:"subdomain"`
	RootDomain types.String `tfsdk:"rootdomain"`
	FullDomain types.String `tfsdk:"full_domain"`
	Status     types.String `tfsdk:"status"`
	ExpiresAt  types.String `tfsdk:"expires_at"`
	AutoRenew  types.Bool   `tfsdk:"auto_renew"` // 供架构联动，非必须
}

func NewSubdomainResource() resource.Resource { return &subdomainResource{} }
func (r *subdomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subdomain"
}
func (r *subdomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*dnshe.Client)
	}
}
func (r *subdomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:      true,
				Description:   "Computed numeric ID of the subdomain resource assigned by the provider.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"subdomain": schema.StringAttribute{
				Required:    true,
				Description: "Subdomain label (the left-most portion of the domain, e.g. 'app' for app.example.com).",
			},
			"rootdomain": schema.StringAttribute{
				Required:    true,
				Description: "Root domain under which the subdomain is registered (e.g. example.com).",
			},
			"full_domain": schema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified domain name of the registered subdomain (e.g. app.example.com).",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current provisioning status of the subdomain (e.g. active, pending, deleted).",
			},
			"expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "Expiration timestamp of the subdomain registration in ISO 8601 format, if applicable.",
			},
			"auto_renew": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the subdomain is configured to auto-renew at expiration. This is used for UI/automation and may be ignored by the provider.",
			},
		},
	}
}

func (r *subdomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data subdomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.RegisterSubdomain(data.Subdomain.ValueString(), data.RootDomain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("注册子域名失败", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(res.SubdomainID))
	data.FullDomain = types.StringValue(res.FullDomain)

	// 通过 Read 同步额外字段状态 (如 Status 和 ExpiresAt)
	r.readIntoModel(ctx, int(data.ID.ValueInt64()), &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subdomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data subdomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readIntoModel(ctx, int(data.ID.ValueInt64()), &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subdomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state subdomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 根据 API 特性：如果改了名字，只能走先删后建。
	// 这里我们演示对存量域名触发官方提供的 RenewSubdomain(id) 的续期链路
	_, err := r.client.RenewSubdomain(int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("续期失败", err.Error())
		return
	}

	plan.ID = state.ID
	r.readIntoModel(ctx, int(plan.ID.ValueInt64()), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subdomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data subdomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteSubdomain(int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("删除子域名失败", err.Error())
		return
	}
}

func (r *subdomainResource) readIntoModel(_ context.Context, id int, model *subdomainResourceModel, diags *diag.Diagnostics) {
	res, err := r.client.GetSubdomain(id)
	if err != nil {
		diags.AddError("查询详情失败", fmt.Sprintf("ID %d 无法加载: %s", id, err.Error()))
		return
	}
	model.Status = types.StringValue(res.Subdomain.Status)
	model.ExpiresAt = types.StringValue(res.Subdomain.ExpiresAt)
	model.FullDomain = types.StringValue(res.Subdomain.FullDomain)
}
