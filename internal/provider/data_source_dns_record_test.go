package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceDNSRecords(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSubdomainConfig,
				Check:  resource.ComposeAggregateTestCheckFunc(
				// resource.TestCheckResourceAttr("data.dnshe_subdomain.test_domain", "subdomain", "krab"),
				// resource.TestCheckResourceAttrSet("data.dnshe_subdomain.test_domain", "id"),

				),
			},
		},
	})
}

const testAccDataSourceDNSRecordsConfig = `
data "dnshe_dns_records" "test_domain" {
  subdomain_id  = "xxxxx"
}
`
