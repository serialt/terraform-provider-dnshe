package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serialt/terraform-provider-dnshe/dnshe"
)

type subdomainsDataSource struct{ client *dnshe.Client }
type subdomainsDSModel struct {
	ID         types.String       `tfsdk:"id"`
	RootDomain types.String       `tfsdk:"rootdomain"`
	Search     types.String       `tfsdk:"search"`
	Subdomains []subdomainDSModel `tfsdk:"subdomains"`
}

func NewSubdomainsDataSource() datasource.DataSource { return &subdomainsDataSource{} }
func (d *subdomainsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subdomains"
}
func (d *subdomainsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*dnshe.Client)
	}
}
func (d *subdomainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Computed identifier for this data source (constant 'subdomain_list').",
			},
			"rootdomain": schema.StringAttribute{
				Optional:    true,
				Description: "Filter results to subdomains under this root domain (e.g. example.com).",
			},
			"search": schema.StringAttribute{
				Optional:    true,
				Description: "Search string to filter subdomains by name.",
			},
			"subdomains": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Provider numeric ID for the subdomain.",
						},
						"subdomain_id": schema.Int64Attribute{
							Computed:    true,
							Description: "Numeric ID of the subdomain record.",
						},
						"subdomain": schema.StringAttribute{
							Computed:    true,
							Description: "Subdomain label (left-most portion of the domain).",
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
				},
			},
		},
	}
}
func (d *subdomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subdomainsDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := dnshe.ListSubdomainsParams{RootDomain: data.RootDomain.ValueString()}
	if data.Search.ValueString() != "" {
		p.Search = data.Search.ValueString()
	}
	res, err := d.client.ListSubdomains(p)
	if err != nil {
		resp.Diagnostics.AddError("API错误", err.Error())
		return
	}

	data.ID = types.StringValue("subdomain_list")
	data.Subdomains = []subdomainDSModel{}
	for _, sub := range res.Subdomains {
		data.Subdomains = append(data.Subdomains, subdomainDSModel{
			ID:                types.Int64Value(int64(sub.ID)),
			SubdomainID:       types.Int64Value(int64(sub.ID)),
			Subdomain:         types.StringValue(sub.Subdomain),
			RootDomain:        types.StringValue(sub.RootDomain),
			FullDomain:        types.StringValue(sub.FullDomain),
			Status:            types.StringValue(sub.Status),
			CreatedAt:         types.StringValue(sub.CreatedAt),
			UpdatedAt:         types.StringValue(sub.UpdatedAt),
			ExpiresAt:         types.StringValue(sub.ExpiresAt),
			NeverExpires:      types.Int64Value(int64(sub.NeverExpires)),
			CloudflareZoneID:  types.StringValue(sub.CloudflareZoneID),
			ProviderAccountID: types.Int64Value(int64(sub.ProviderAccountID)),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
