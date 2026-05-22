package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testAccResourceDNSrecordCreateConfig = `
resource "dnshe_dns_record" "test_record" {
  subdomain_id = "5004916620"
  type         = "A"
  name         = "k8.krab.bbroot.com"
  content      = "1.1.1.1"
}
`

const testAccResourceDNSrecordUpdateConfig = `
resource "dnshe_dns_record" "test_record" {
  subdomain_id = "5004916620"
  type         = "A"
  name         = "k8.krab.bbroot.com"
  content      = "9.9.9.9"
}
`

func TestAccResourceDNSRecord(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccResourceDNSrecordCreateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnshe_dns_record.test_record", "subdomain_id", "5004916620"),
					resource.TestCheckResourceAttr("dnshe_dns_record.test_record", "type", "A"),
					resource.TestCheckResourceAttr("dnshe_dns_record.test_record", "name", "k8.krab.bbroot.com"),
					resource.TestCheckResourceAttr("dnshe_dns_record.test_record", "content", "1.1.1.1"),
				),
			},
			// Update
			{
				Config: testAccResourceDNSrecordUpdateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnshe_dns_record.test_record", "content", "9.9.9.9"),
				),
			},
			// Debugging output (prints resource JSON to test logs)
			{
				Config: testAccResourceDNSrecordUpdateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
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
