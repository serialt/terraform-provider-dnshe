data "dnshe_dns_records" "test_records" {
  subdomain_id = "5004916620"
}


data "dnshe_dns_records" "test_v1_record" {
  subdomain = "krab.bbroot.com"
}
