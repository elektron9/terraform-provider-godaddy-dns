---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "godaddy-dns_record Resource - terraform-provider-godaddy-dns"
subcategory: ""
description: |-
  GoDaddy DNS record
---

# godaddy-dns_record (Resource)

GoDaddy DNS record

## Example Usage

```terraform
resource "godaddy-dns_record" "cname" {
  domain = "mydomain.com"

  type = "CNAME"
  name = "redirect"
  data = "target.otherdomain.com"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `data` (String) contents: target for CNAME, ip address for A etc
- `domain` (String) managed domain (top-level)
- `name` (String) name (part of fqdn), may include `.` for sub-domains
- `type` (String) type: A, CNAME etc

### Read-Only

- `ttl` (Number) TTL, > 600 < 86400, def 3600