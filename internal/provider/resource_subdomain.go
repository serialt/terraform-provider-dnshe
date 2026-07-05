package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

	data, err = r.fetchAndUpdateState(ctx, data, int64(res.SubdomainID))
	if err != nil {
		tflog.Error(ctx, "Failed to fetch created record", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subdomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data subdomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.fetchAndUpdateState(ctx, data, data.ID.ValueInt64())
	if err != nil {
		tflog.Warn(ctx, "Record not found, removing from state", map[string]interface{}{"id": data.ID.ValueInt64()})
		resp.State.RemoveResource(ctx)
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

	tflog.Info(ctx, "Please modify subdomain in web")
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

// fetchAndUpdateState 从 API 获取最新的记录数据并更新模型状态
func (r *subdomainResource) fetchAndUpdateState(ctx context.Context, model subdomainResourceModel, subdomainID int64) (subdomainResourceModel, error) {
	domainResp, err := r.client.GetSubdomain((int(subdomainID)))
	if err != nil {
		return model, err
	}

	subdomain := domainResp.Subdomain
	model.FullDomain = types.StringValue(subdomain.FullDomain)
	model.Status = types.StringValue(subdomain.Status)
	model.ExpiresAt = types.StringValue(subdomain.ExpiresAt)
	model.ID = types.Int64Value(int64(subdomain.ID))

	return model, err
}
