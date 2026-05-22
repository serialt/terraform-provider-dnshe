package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serialt/terraform-provider-dnshe/dnshe"
)

var _ provider.Provider = &DNSHEProvider{}
var _ provider.ProviderWithFunctions = &DNSHEProvider{}

type DNSHEProvider struct {
	version string
}
type DNSHEProviderModel struct {
	BaseURL   types.String `tfsdk:"base_url"`
	ApiKey    types.String `tfsdk:"api_key"`
	ApiSecret types.String `tfsdk:"api_secret"`
}

func (p *DNSHEProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dnshe"
	resp.Version = p.version
}

func (p *DNSHEProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "DNSHE Terraform Provider，用于全自动化管理自定义二级域名及 DNS 解析记录。",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "Base API URL for the DNSHE service. If empty, the provider default will be used.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "DNSHE API key used for authenticating requests. Can also be set via the DNSHE_API_KEY environment variable.",
				Optional:    true,
			},
			"api_secret": schema.StringAttribute{
				Description: "DNSHE API secret used for authenticating requests. Can also be set via the DNSHE_API_SECRET environment variable. This value is sensitive.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *DNSHEProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data DNSHEProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiKey := os.Getenv("DNSHE_API_KEY")
	apiSecret := os.Getenv("DNSHE_API_SECRET")
	baseURL := data.BaseURL.ValueString()

	if !data.ApiKey.IsNull() {
		apiKey = data.ApiKey.ValueString()
	}
	if !data.ApiSecret.IsNull() {
		apiSecret = data.ApiSecret.ValueString()
	}

	if apiKey == "" || apiSecret == "" {
		resp.Diagnostics.AddError("凭证缺失", "必须配置 api_key 和 api_secret (或设置对应的环境变量)")
		return
	}

	client := dnshe.NewClient(baseURL, apiKey, apiSecret)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *DNSHEProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSRecordResource,
		NewSubdomainResource,
	}
}

func (p *DNSHEProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDNSRecordsDataSource,
		NewQuotaDataSource,
		NewSubdomainDataSource,
		NewSubdomainsDataSource,
	}
}

func (p *DNSHEProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DNSHEProvider{
			version: version,
		}
	}
}
