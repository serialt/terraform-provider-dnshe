terraform {
  required_providers {
    dnshe = {
      source = "serialt/dnshe"
    }
  }

}

# Provider with default values
provider "dnshe" {
  base_url   = "https://api005.dnshe.com/index.php?m=domain_hub"
  api_key    = "xxxxx"
  api_secret = "xxxxxxxxxx"

  # DNSHE_API_KEY = ""
  # DNSHE_API_SECRET = ""
}
