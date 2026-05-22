package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceSubdomains(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSubdomainsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.dnshe_subdomains.test_domain", "id", func(value string) error {
						fmt.Printf("\n======================= 🔍 DNSHE 数据打样 🔍 =======================\n")
						fmt.Printf("  数据源实际解析到的 ID 值为 : %s", value)
						return nil
					}),
					func(s *terraform.State) error {
						for _, ms := range s.RootModule().Resources {
							fmt.Printf("唯一 ID : %s\n", ms.Primary.ID)
							fmt.Printf("------------------------- 属性列表 (JSON 格式化) -------------------------\n")
							jsonBytes, _ := json.MarshalIndent(ms, "", "  ")
							fmt.Println(string(jsonBytes))
							fmt.Printf("====================================================================\n\n")
						}
						return nil
					},
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
