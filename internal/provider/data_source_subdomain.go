package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serialt/terraform-provider-dnshe/dnshe"
)

// ==========================================
// 1. 单个子域名数据源 (dnshe_subdomain)
// ==========================================
type subdomainDataSource struct {
	client *dnshe.Client
}
type subdomainDSModel struct {
	ID                types.Int64  `tfsdk:"id"`
	SubdomainID       types.Int64  `tfsdk:"subdomain_id"`
	Subdomain         types.String `tfsdk:"subdomain"`
	RootDomain        types.String `tfsdk:"rootdomain"`
	FullDomain        types.String `tfsdk:"full_domain"`
	Status            types.String `tfsdk:"status"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	ExpiresAt         types.String `tfsdk:"expires_at"`
	NeverExpires      types.Int64  `tfsdk:"never_expires"`
	CloudflareZoneID  types.String `tfsdk:"cloudflare_zone_id"`
	ProviderAccountID types.Int64  `tfsdk:"provider_account_id"`
}

func NewSubdomainDataSource() datasource.DataSource { return &subdomainDataSource{} }
func (d *subdomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subdomain"
}
func (d *subdomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*dnshe.Client)
	}
}
func (d *subdomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Computed provider ID for this subdomain resource.",
			},
			"subdomain_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Numeric ID of the subdomain to look up.",
				Validators: []validator.Int64{
					// 当 subdomain_id 存在时，确保 subdomain 没有被设置
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("subdomain")),
				},
			},
			"subdomain": schema.StringAttribute{
				Optional:    true,
				Description: "Subdomain label (left-most portion of the domain).",
				Validators: []validator.String{
					// 当 subdomain 存在时，确保 subdomain_id 没有被设置
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("subdomain_id")),
				},
			},
			"rootdomain": schema.StringAttribute{
				Computed:    true,
				Description: "Root domain under which the subdomain is registered.",
			},
			"full_domain": schema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified domain name (e.g. sub.example.com).",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current provisioning status of the subdomain (e.g. active, pending).",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the subdomain was created (ISO 8601).",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the last update to the subdomain (ISO 8601).",
			},
			"expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "Expiration timestamp of the subdomain registration (ISO 8601), if applicable.",
			},
			"never_expires": schema.Int64Attribute{
				Computed:    true,
				Description: "Indicator whether the subdomain never expires (1) or not (0).",
			},
			"cloudflare_zone_id": schema.StringAttribute{
				Computed:    true,
				Description: "Associated Cloudflare zone ID if the domain is proxied via Cloudflare.",
			},
			"provider_account_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Numeric ID of the provider account that owns the subdomain.",
			},
		},
	}
}
func (d *subdomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subdomainDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var subdomainId int
	if !data.SubdomainID.IsNull() {
		subdomainId = int(data.SubdomainID.ValueInt64())
	}
	// 如果子域名存在，则优先查询子域名
	if !data.Subdomain.IsNull() {
		subdomainResp, err := d.client.ListSubdomains(dnshe.ListSubdomainsParams{
			Search: data.Subdomain.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("API err", err.Error())
			return
		}
		if subdomainResp.Count != 1 {
			resp.Diagnostics.AddError("API err", "Multiple records found. ")
			return
		}
		subdomainId = subdomainResp.Subdomains[0].ID
	}

	res, err := d.client.GetSubdomain(subdomainId)
	if err != nil {
		resp.Diagnostics.AddError("API错误", err.Error())
		return
	}
	data.ID = types.Int64Value(int64(res.Subdomain.ID))
	data.Subdomain = types.StringValue(res.Subdomain.Subdomain)
	data.RootDomain = types.StringValue(res.Subdomain.RootDomain)
	data.FullDomain = types.StringValue(res.Subdomain.FullDomain)
	data.Status = types.StringValue(res.Subdomain.Status)
	data.CreatedAt = types.StringValue(res.Subdomain.CreatedAt)
	data.UpdatedAt = types.StringValue(res.Subdomain.UpdatedAt)
	data.ExpiresAt = types.StringValue(res.Subdomain.ExpiresAt)
	data.NeverExpires = types.Int64Value(int64(res.Subdomain.NeverExpires))
	data.CloudflareZoneID = types.StringValue(res.Subdomain.CloudflareZoneID)
	data.ProviderAccountID = types.Int64Value(int64(res.Subdomain.ProviderAccountID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
