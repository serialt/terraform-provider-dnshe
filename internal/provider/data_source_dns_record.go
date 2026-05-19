package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serialt/terraform-provider-dnshe/dnshe"
)

// ==========================================
// 3. DNS 记录列表数据源 (dnshe_dns_records)
// ==========================================
type dnsRecordsDataSource struct{ client *dnshe.Client }
type dnsRecordModel struct {
	ID       types.Int64  `tfsdk:"id"`
	RecordID types.String `tfsdk:"record_id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Content  types.String `tfsdk:"content"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Priority types.Int64  `tfsdk:"priority"`
	Line     types.String `tfsdk:"line"`
	Proxied  types.Bool   `tfsdk:"proxied"`
	Status   types.String `tfsdk:"status"`
}
type dnsRecordsDSModel struct {
	ID          types.String     `tfsdk:"id"`
	SubdomainID types.Int64      `tfsdk:"subdomain_id"`
	Records     []dnsRecordModel `tfsdk:"records"`
}

func NewDNSRecordsDataSource() datasource.DataSource { return &dnsRecordsDataSource{} }
func (d *dnsRecordsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_records"
}
func (d *dnsRecordsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*dnshe.Client)
	}
}
func (d *dnsRecordsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true},
			"subdomain_id": schema.Int64Attribute{Required: true},
			"records": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":        schema.Int64Attribute{Computed: true},
						"record_id": schema.StringAttribute{Computed: true},
						"name":      schema.StringAttribute{Computed: true},
						"type":      schema.StringAttribute{Computed: true},
						"content":   schema.StringAttribute{Computed: true},
						"ttl":       schema.Int64Attribute{Computed: true},
						"priority":  schema.Int64Attribute{Computed: true},
						"line":      schema.StringAttribute{Computed: true},
						"proxied":   schema.BoolAttribute{Computed: true},
						"status":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}
func (d *dnsRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dnsRecordsDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.ListDNSRecords(int(data.SubdomainID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("API错误", err.Error())
		return
	}

	data.ID = types.StringValue("dns_record")
	data.Records = []dnsRecordModel{}
	for _, r := range res.Records {
		prio := int64(0)
		if r.Priority != nil {
			prio = int64(*r.Priority)
		}
		data.Records = append(data.Records, dnsRecordModel{
			ID:       types.Int64Value(int64(r.ID)),
			RecordID: types.StringValue(r.RecordID),
			Name:     types.StringValue(r.Name),
			Type:     types.StringValue(r.Type),
			Content:  types.StringValue(r.Content),
			TTL:      types.Int64Value(int64(r.TTL)),
			Priority: types.Int64Value(prio),
			Line:     types.StringValue(r.Line),
			Proxied:  types.BoolValue(r.Proxied),
			Status:   types.StringValue(r.Status),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
