package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSubdomains(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSubdomainConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("data.dnshe_subdomain.test_domain", "subdomain", "krab"),
					// resource.TestCheckResourceAttrSet("data.dnshe_subdomain.test_domain", "id"),

					// 🛠️ 使用这个方法，完全不依赖任何第三方 State 包，100% 避开类型冲突
					resource.TestCheckResourceAttrWith("data.dnshe_subdomains.test_domain", "id", func(value string) error {
						t.Log("\n======================= 🔍 DNSHE 数据打样 🔍 =======================")
						t.Logf("  数据源实际解析到的 ID 值为 : %s", value)
						fmt.Printf("  数据源实际解析到的 ID 值为 : %s", value)
						return nil
					}),

					// resource.TestCheckResourceAttrWith("data.dnshe_subdomains.test_domain", "full_domain", func(value string) error {
					// 	t.Logf("  数据源实际获取的完整域名为 : %s", value)
					// 	fmt.Printf("  数据源实际获取的完整域名为 : %s", value)
					// 	t.Log("====================================================================")
					// 	return nil
					// }),
				),
			},
		},
	})
}

const testAccDataSourceSubdomainsConfig = `
data "dnshe_subdomains" "test_domain" {
//   subdomain  = "krab"
}
`
