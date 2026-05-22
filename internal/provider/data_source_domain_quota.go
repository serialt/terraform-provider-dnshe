package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serialt/terraform-provider-dnshe/dnshe"
)

// ==========================================
// 4. 账号配额数据源 (dnshe_quota)
// ==========================================
type quotaDataSource struct{ client *dnshe.Client }

type quotaDSModel struct {
	ID          types.String `tfsdk:"id"`
	Used        types.Int64  `tfsdk:"used"`
	Base        types.Int64  `tfsdk:"base"`
	InviteBonus types.Int64  `tfsdk:"invite_bonus"`
	Total       types.Int64  `tfsdk:"total"`
	Available   types.Int64  `tfsdk:"available"`
}

func NewQuotaDataSource() datasource.DataSource { return &quotaDataSource{} }
func (d *quotaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_quota"
}
func (d *quotaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*dnshe.Client)
	}
}
func (d *quotaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Computed identifier for this data source (constant 'domain_quota').",
			},
			"used": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of subdomains already used/registered on the account.",
			},
			"base": schema.Int64Attribute{
				Computed:    true,
				Description: "Base quota allocated to the account before bonuses.",
			},
			"invite_bonus": schema.Int64Attribute{
				Computed:    true,
				Description: "Additional quota granted via invitations or promotions.",
			},
			"total": schema.Int64Attribute{
				Computed:    true,
				Description: "Total quota available to the account (base plus bonuses).",
			},
			"available": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of remaining available subdomain slots the account can register.",
			},
		},
	}
}
func (d *quotaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data quotaDSModel
	res, err := d.client.GetQuota()
	if err != nil {
		resp.Diagnostics.AddError("API错误", err.Error())
		return
	}
	data.ID = types.StringValue("domain_quota")
	data.Used = types.Int64Value(int64(res.Quota.Used))
	data.Base = types.Int64Value(int64(res.Quota.Base))
	data.InviteBonus = types.Int64Value(int64(res.Quota.InviteBonus))
	data.Total = types.Int64Value(int64(res.Quota.Total))
	data.Available = types.Int64Value(int64(res.Quota.Available))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
