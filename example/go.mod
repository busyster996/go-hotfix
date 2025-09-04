module github.com/busyster996/go-hotfix/example

go 1.24

replace github.com/busyster996/go-hotfix => ../

require (
	github.com/busyster996/go-hotfix v0.0.0-00010101000000-000000000000
	github.com/pires/go-proxyproto v0.8.1
	github.com/reiver/go-oi v1.0.0
	github.com/reiver/go-telnet v0.0.0-20250617105250-7da9ad70a2b2
	github.com/soheilhy/cmux v0.1.5
)

require (
	github.com/agiledragon/gomonkey/v2 v2.13.0 // indirect
	github.com/traefik/yaegi v0.16.1 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)
