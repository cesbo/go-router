# Middleware for go-router

## CloudFlare

The CloudFlare middleware checks for a CF-Connecting-IP header
If request from CloudFlare network then request.RemoteAddr will be set to CF-Connecting-IP

Usage:

```go
handler := router.NewRouter()
handler.Use(mw.CloudFlareMW())
```
