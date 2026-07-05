data "dnshe_subdomain" "test_domain_single" {
  subdomain_id = "5004455322"
}

data "dnshe_subdomain" "test_domain_single_by_domain" {
  subdomain = "krab.bbroot.com"
}
