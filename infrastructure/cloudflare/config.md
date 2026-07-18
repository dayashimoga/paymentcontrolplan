# Cloudflare Configuration for PCP

## DNS Records
# A     api.pcp.example.com  -> Load Balancer IP  (Proxied)
# CNAME www.pcp.example.com  -> api.pcp.example.com (Proxied)

## WAF Rules (Cloudflare Ruleset)

### Rate Limiting
# Rule: API Rate Limit
# Expression: (http.request.uri.path matches "^/api/")
# Action: Block after 1000 requests per minute per IP

### Bot Protection
# Rule: Block Known Bad Bots
# Expression: (cf.bot_management.score lt 30)
# Action: Managed Challenge

### Geo Blocking (if needed)
# Rule: Block Sanctioned Countries
# Expression: (ip.geoip.country in {"KP" "IR" "SY"})
# Action: Block

## Page Rules
# api.pcp.example.com/health  -> Cache Level: Bypass
# api.pcp.example.com/ready   -> Cache Level: Bypass
# api.pcp.example.com/api/*   -> Cache Level: Bypass, SSL: Full (Strict)

## SSL/TLS
# Mode: Full (Strict)
# Minimum TLS: 1.2
# TLS 1.3: Enabled
# HSTS: Enabled (max-age=31536000, includeSubDomains)

## Security Headers (Transform Rules)
# X-Content-Type-Options: nosniff
# X-Frame-Options: DENY
# Referrer-Policy: strict-origin-when-cross-origin
# Permissions-Policy: camera=(), microphone=(), geolocation=()

## Terraform Cloudflare Provider
# resource "cloudflare_zone" "pcp" {
#   zone = "pcp.example.com"
#   plan = "pro"
# }
#
# resource "cloudflare_record" "api" {
#   zone_id = cloudflare_zone.pcp.id
#   name    = "api"
#   value   = module.eks.cluster_endpoint
#   type    = "CNAME"
#   proxied = true
# }
#
# resource "cloudflare_rate_limit" "api" {
#   zone_id   = cloudflare_zone.pcp.id
#   threshold = 1000
#   period    = 60
#   match {
#     request { url_pattern = "api.pcp.example.com/api/*" }
#   }
#   action { mode = "ban", timeout = 300 }
# }
