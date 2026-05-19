package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dnshe": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.

	if os.Getenv("DNSHE_API_KEY") == "" {
		t.Fatal("环境缺失: 必须设置环境变量 DNSHE_API_KEY 才能运行集成测试")
	}
	if os.Getenv("DNSHE_API_SECRET") == "" {
		t.Fatal("环境缺失: 必须设置环境变量 DNSHE_API_SECRET 才能运行集成测试")
	}

}
