terraform {
  required_providers {
    dnshe = {
      source = "serialt/dnshe"
    }
  }
}
provider "dnshe" {

}

resource "dnshe_dns_record" "record" {
  subdomain_id = "xxxxxxx"
  type         = "A"
  name         = "vvvcvvv.kkkkk.bbroot.com"
  content      = "1.1.1.1"

}

resource "dnshe_dns_record" "record2" {
  subdomain_id = "xxxxxxx"
  type         = "A"
  name         = "8888.kkkkk.bbroot.com"
  content      = "8.8.8.8"

}


resource "dnshe_dns_record" "record3" {
  subdomain_id = "xxxxxxx"
  type         = "A"
  name         = "555.kkkkk.bbroot.com"
  content      = "1.2.3.4"

}
