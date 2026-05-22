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
			"id":         schema.StringAttribute{Computed: true},
			"rootdomain": schema.StringAttribute{Optional: true},
			"search":     schema.StringAttribute{Optional: true},
			"subdomains": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                  schema.Int64Attribute{Computed: true},
						"subdomain_id":        schema.Int64Attribute{Computed: true},
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
