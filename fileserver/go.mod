module fileserver

go 1.13

require (
	filestore v0.0.0
	github.com/bouk/httprouter v0.0.0-20160817010721-ee8b3818a7f5
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.3.0 // indirect
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
)

replace filestore => ../filestore
