package provider

import (
	"context"
	"strconv"

	"github.com/serialt/terraform-provider-dnshe/dnshe"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &dnsRecordResource{}

type dnsRecordResource struct{ client *dnshe.Client }

type dnsRecordResourceModel struct {
	ID          types.String `tfsdk:"id"` // 本地一般采用 RecordID 充当字符串 ID
	SubdomainID types.Int64  `tfsdk:"subdomain_id"`
	Type        types.String `tfsdk:"type"`
	Name        types.String `tfsdk:"name"`
	Content     types.String `tfsdk:"content"`
	TTL         types.Int64  `tfsdk:"ttl"`
	Priority    types.Int64  `tfsdk:"priority"`
	Line        types.String `tfsdk:"line"`
}

func NewDNSRecordResource() resource.Resource { return &dnsRecordResource{} }
func (r *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}
func (r *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*dnshe.Client)
	}
}
func (r *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subdomain_id": schema.Int64Attribute{Required: true},
			"type":         schema.StringAttribute{Required: true},
			"name":         schema.StringAttribute{Optional: true, Computed: true},
			"content":      schema.StringAttribute{Required: true},
			"ttl":          schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(600)},
			"priority":     schema.Int64Attribute{Optional: true},
			"line":         schema.StringAttribute{Optional: true, Computed: true},
		},
	}
}

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := dnshe.CreateDNSRecordRequest{
		SubdomainID: int(data.SubdomainID.ValueInt64()),
		Type:        data.Type.ValueString(),
		Name:        data.Name.ValueString(),
		Content:     data.Content.ValueString(),
	}
	if !data.TTL.IsUnknown() && !data.TTL.IsNull() {
		apiReq.TTL = int(data.TTL.ValueInt64())
	}
	if !data.Priority.IsUnknown() && !data.Priority.IsNull() {
		prio := int(data.Priority.ValueInt64())
		apiReq.Priority = &prio
	}
	if !data.Line.IsUnknown() && !data.Line.IsNull() {
		apiReq.Line = data.Line.ValueString()
	}

	res, err := r.client.CreateDNSRecord(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("创建解析记录失败", err.Error())
		return
	}

	data.ID = types.StringValue(res.RecordID)
	if data.Line.IsUnknown() {
		data.Line = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// API 仅提供通过 subdomain_id 列出所有记录，因此我们通过遍历匹配目标 RecordID
	res, err := r.client.ListDNSRecords(int(data.SubdomainID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("拉取解析记录失败", err.Error())
		return
	}

	// fmt.Println(res)
	found := false
	for _, rec := range res.Records {
		if rec.RecordID == data.ID.ValueString() {
			if rec.Name != "" {
				data.Name = types.StringValue(rec.Name)
			} else {
				data.Name = types.StringNull()
			}
			data.Type = types.StringValue(rec.Type)
			data.Content = types.StringValue(rec.Content)
			data.TTL = types.Int64Value(int64(rec.TTL))
			if rec.Priority != nil {
				data.Priority = types.Int64Value(int64(*rec.Priority))
			} else {
				data.Priority = types.Int64Null()
			}
			if rec.Line != "" {
				data.Line = types.StringValue(rec.Line)
			} else {
				data.Line = types.StringNull()
			}
			found = true
			break
		}
	}

	if !found {
		// 说明远端已经不存在该条记录，从本地 State 彻底移除它
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idInt, _ := strconv.Atoi(plan.ID.ValueString())
	apiReq := dnshe.UpdateDNSRecordRequest{
		ID:       idInt,
		RecordID: plan.ID.ValueString(),
		Type:     plan.Type.ValueString(),
		Name:     plan.Name.ValueString(),
		Content:  plan.Content.ValueString(),
	}
	if !plan.TTL.IsUnknown() && !plan.TTL.IsNull() {
		apiReq.TTL = int(plan.TTL.ValueInt64())
	}
	if !plan.Priority.IsUnknown() && !plan.Priority.IsNull() {
		prio := int(plan.Priority.ValueInt64())
		apiReq.Priority = &prio
	}
	if !plan.Line.IsUnknown() && !plan.Line.IsNull() {
		apiReq.Line = plan.Line.ValueString()
	}

	_, err := r.client.UpdateDNSRecord(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("更新解析记录失败", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idInt, _ := strconv.Atoi(data.ID.ValueString())
	_, err := r.client.DeleteDNSRecord(idInt, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("删除解析记录失败", err.Error())
		return
	}
}
