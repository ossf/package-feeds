module github.com/ossf/package-feeds/cmd/scheduled-feed

go 1.15

require (
	github.com/ossf/package-feeds/feeds v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.7.0
	gocloud.dev v0.21.0
)

replace github.com/ossf/package-feeds/feeds => ../../feeds
