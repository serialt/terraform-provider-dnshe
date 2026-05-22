package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceDNSQuota(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDNSQuotaConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("data.dnshe_subdomain.test_domain", "subdomain", "krab"),
					// resource.TestCheckResourceAttrSet("data.dnshe_subdomain.test_domain", "id"),
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

const testAccDataSourceDNSQuotaConfig = `
data "dnshe_domain_quota" "test_quota" {
}
`
