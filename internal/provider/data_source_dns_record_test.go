package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceDNSRecords(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// {
			// 	Config: testAccDataSourceDNSRecordsConfig,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		func(s *terraform.State) error {
			// 			for _, ms := range s.RootModule().Resources {
			// 				fmt.Printf("唯一 ID : %s\n", ms.Primary.ID)
			// 				fmt.Printf("------------------------- 属性列表 (JSON 格式化) -------------------------\n")
			// 				jsonBytes, _ := json.MarshalIndent(ms, "", "  ")
			// 				fmt.Println(string(jsonBytes))
			// 				fmt.Printf("====================================================================\n\n")
			// 			}
			// 			return nil
			// 		},
			// 	),
			// },
			{
				Config: testAccDataSourceDNSRecordsConfigV1,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						for _, ms := range s.RootModule().Resources {
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

const testAccDataSourceDNSRecordsConfig = `
data "dnshe_dns_records" "test_record" {
  subdomain_id  = "5004916620"
}
`

const testAccDataSourceDNSRecordsConfigV1 = `
data "dnshe_dns_records" "test_v1_record" {
  subdomain  = "krab.bbroot.com"
}
`
