package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/serialt/terraform-provider-dnshe/dnshe"
	"github.com/spf13/cast"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &dnsRecordResource{}

type dnsRecordResource struct {
	client *dnshe.Client
}

type dnsRecordResourceModel struct {
	ID          types.String `tfsdk:"id"`        // 组合 ID: dns_record#subdomain_id#record_id
	RecordID    types.String `tfsdk:"record_id"` // API 返回的记录 ID
	SubdomainID types.Int64  `tfsdk:"subdomain_id"`
	Type        types.String `tfsdk:"type"`
	Name        types.String `tfsdk:"name"`
	Content     types.String `tfsdk:"content"`
	TTL         types.Int64  `tfsdk:"ttl"`
	Priority    types.Int64  `tfsdk:"priority"`
	Line        types.String `tfsdk:"line"`
}

func NewDNSRecordResource() resource.Resource {
	return &dnsRecordResource{}
}

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
				Description:   "Computed composite ID in the form dns_record#<subdomain_id>#<record_id>",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subdomain_id": schema.Int64Attribute{
				Required:    true,
				Description: "ID of the parent subdomain this DNS record belongs to.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "DNS record type (e.g., A, AAAA, CNAME, MX, TXT).",
			},
			"record_id": schema.StringAttribute{
				Computed:    true,
				Description: "Provider-assigned unique identifier for the DNS record.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Record name (relative to the zone). Use an empty string for the zone root.",
			},
			"content": schema.StringAttribute{
				Required:    true,
				Description: "Record value/content (IP address, target hostname, text string, etc.).",
			},
			"ttl": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(600),
				Description: "Time to live in seconds. Defaults to 600 if not specified.",
			},
			"priority": schema.Int64Attribute{
				Optional:    true,
				Description: "Priority for MX or SRV records. Omit for record types that do not use priority.",
			},
			"line": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Routing line or WAN line identifier used by the provider (optional).",
			},
		},
	}
}

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 构建 API 请求
	apiReq := dnshe.CreateDNSRecordRequest{
		SubdomainID: int(data.SubdomainID.ValueInt64()),
		Type:        data.Type.ValueString(),
		Name:        data.Name.ValueString(),
		Content:     data.Content.ValueString(),
	}

	// 处理可选字段
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

	// 调用 API 创建记录
	res, err := r.client.CreateDNSRecord(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("创建解析记录失败", fmt.Sprintf("SubdomainID: %d, Error: %v", apiReq.SubdomainID, err))
		return
	}

	// 更新状态
	data, err = r.fetchAndUpdateState(ctx, data, int(data.SubdomainID.ValueInt64()), res.ID)
	if err != nil {
		tflog.Error(ctx, "Failed to fetch created record", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subdomainID, recordID := parseCompositeID(data.ID.ValueString())
	if subdomainID == 0 || recordID == 0 {
		resp.Diagnostics.AddError("Invalid resource ID", "Failed to parse composite ID")
		return
	}

	// 从 API 重新读取最新数据
	data, err := r.fetchAndUpdateState(ctx, data, subdomainID, recordID)
	if err != nil {
		tflog.Warn(ctx, "Record not found, removing from state", map[string]interface{}{"id": data.ID.ValueString()})
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

	subdomainID, recordID := parseCompositeID(plan.ID.ValueString())
	if subdomainID == 0 || recordID == 0 {
		resp.Diagnostics.AddError("Invalid resource ID", "Failed to parse composite ID")
		return
	}

	// 构建 API 请求
	apiReq := dnshe.UpdateDNSRecordRequest{
		ID:      recordID,
		Type:    plan.Type.ValueString(),
		Name:    plan.Name.ValueString(),
		Content: plan.Content.ValueString(),
	}

	// 处理可选字段
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

	// 调用 API 更新记录
	_, err := r.client.UpdateDNSRecord(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("更新解析记录失败", fmt.Sprintf("RecordID: %s, Error: %v", plan.RecordID.ValueString(), err))
		return
	}

	// 从 API 重新读取最新数据
	plan, err = r.fetchAndUpdateState(ctx, plan, subdomainID, recordID)
	if err != nil {
		resp.Diagnostics.AddError("更新后读取记录失败", err.Error())
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

	_, recordID := parseCompositeID(data.ID.ValueString())
	if recordID == 0 {
		resp.Diagnostics.AddError("Invalid resource ID", "Failed to parse record ID")
		return
	}

	// 调用 API 删除记录
	_, err := r.client.DeleteDNSRecord(recordID, "")
	if err != nil {
		resp.Diagnostics.AddError("删除解析记录失败", fmt.Sprintf("RecordID: %d, Error: %v", recordID, err))
		return
	}
}

func (r *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildCompositeID 生成组合 ID: dns_record#subdomain_id#record_id
func buildCompositeID(subdomainID int, recordID int) string {
	return fmt.Sprintf("dns_record#%d#%d", subdomainID, recordID)
}

// parseCompositeID 解析组合 ID，返回 (subdomainID, recordID)
func parseCompositeID(compositeID string) (int, int) {
	parts := strings.Split(compositeID, "#")
	if len(parts) != 3 {
		return 0, 0
	}
	subdomainID := cast.ToInt(parts[1])
	recordID := cast.ToInt(parts[2])
	return subdomainID, recordID
}

// fetchAndUpdateState 从 API 获取最新的记录数据并更新模型状态
func (r *dnsRecordResource) fetchAndUpdateState(ctx context.Context, model dnsRecordResourceModel, subdomainID int, recordID int) (dnsRecordResourceModel, error) {

	subdomain, err := r.client.GetSubdomain(subdomainID)
	if err != nil {
		return model, fmt.Errorf("failed to list DNS records: %w", err)
	}

	// 查找目标记录
	var targetRecord *dnshe.DNSRecord
	for _, record := range subdomain.DNSRecords {
		if record.ID == recordID {
			targetRecord = &record
			break
		}
	}

	if targetRecord == nil {
		return model, fmt.Errorf("record with ID %d not found in subdomain %d", recordID, subdomainID)
	}
	// value := strings.TrimSuffix(targetRecord.Content, fmt.Sprintf(".%v", subdomain.Subdomain.FullDomain))

	// 更新模型字段
	model.ID = types.StringValue(buildCompositeID(subdomainID, recordID))
	model.RecordID = types.StringValue(targetRecord.RecordID)
	model.SubdomainID = types.Int64Value(int64(subdomainID))
	model.Type = types.StringValue(targetRecord.Type)
	model.Content = types.StringValue(targetRecord.Content)
	model.TTL = types.Int64Value(int64(targetRecord.TTL))

	// 处理可选字段
	if targetRecord.Name != "" {
		value := strings.TrimSuffix(targetRecord.Name, fmt.Sprintf(".%v", subdomain.Subdomain.FullDomain))
		model.Name = types.StringValue(value)
	} else {
		model.Name = types.StringNull()
	}

	if targetRecord.Priority != nil {
		model.Priority = types.Int64Value(int64(*targetRecord.Priority))
	} else {
		model.Priority = types.Int64Null()
	}

	if targetRecord.Line != "" {
		model.Line = types.StringValue(targetRecord.Line)
	} else {
		model.Line = types.StringNull()
	}

	return model, nil
}
