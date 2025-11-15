module sfsDb

go 1.25.3

require (
	github.com/stretchr/testify v1.7.2
	github.com/syndtr/goleveldb v0.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/syndtr/goleveldb => ../goleveldb //使用本地包，因为需要修改源代码
