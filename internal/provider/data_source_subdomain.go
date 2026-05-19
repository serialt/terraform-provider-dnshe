package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
			"id":                  schema.Int64Attribute{Computed: true},
			"subdomain_id":        schema.Int64Attribute{Required: true},
			"subdomain":           schema.StringAttribute{Computed: true},
			"rootdomain":          schema.StringAttribute{Computed: true},
			"full_domain":         schema.StringAttribute{Computed: true},
			"status":              schema.StringAttribute{Computed: true},
			"created_at":          schema.StringAttribute{Computed: true},
			"updated_at":          schema.StringAttribute{Computed: true},
			"expires_at":          schema.StringAttribute{Computed: true},
			"never_expires":       schema.Int64Attribute{Computed: true},
			"cloudflare_zone_id":  schema.StringAttribute{Computed: true},
			"provider_account_id": schema.Int64Attribute{Computed: true},
		},
	}
}
func (d *subdomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subdomainDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.GetSubdomain(int(data.SubdomainID.ValueInt64()))
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
