# terraform {
#   # required_providers {
#   #   dnshe = {
#   #     source = "serialt/dnshe"
#   #   }
#   # }
# }


# provider "dnshe" {

# }

data "dnshe_subdomains" "test_domain" {
  search = "bbroot"
}

output "domain" {
  value = data.dnshe_subdomains.test_domain

}

data "dnshe_subdomain" "test_domain_single" {
  subdomain_id = "5004916620"
}

output "domain_single" {
  value = data.dnshe_subdomain.test_domain_single

}

data "dnshe_dns_quota" "test_quota" {
}
output "test_quota" {
  value = data.dnshe_dns_quota.test_quota

}

data "dnshe_dns_records" "test_records" {
  subdomain_id = "5004916620"
}

output "test_records" {
  value = data.dnshe_dns_records.test_records
}
