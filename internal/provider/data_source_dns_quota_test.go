package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					resource.TestCheckResourceAttrWith("data.dnshe_dns_quota.test_quota", "total", func(value string) error {
						fmt.Printf("\n======================= 🔍 DNSHE quota 🔍 =======================\n")
						fmt.Printf("  total 值为 : %s", value)

						return nil
					}),
				),
			},
		},
	})
}

const testAccDataSourceDNSQuotaConfig = `
data "dnshe_dns_quota" "test_quota" {
}
`
