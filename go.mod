module github.com/topfreegames/khan

go 1.15

require (
	cloud.google.com/go v0.79.0 // indirect
	github.com/Pallinder/go-randomdata v0.0.0-20160927131605-01563c9f5c2d
	github.com/asaskevich/govalidator v0.0.0-20180315120708-ccb8e960c48f // indirect
	github.com/bluele/factory-go v0.0.0-20160811033936-8a28e9752dbc
	github.com/globalsign/mgo v0.0.0-20180615134936-113d3961e731
	github.com/go-gorp/gorp v2.2.0+incompatible
	github.com/golang/mock v1.5.0
	github.com/gosuri/uilive v0.0.0-20160202011846-efb88ccd0599 // indirect
	github.com/gosuri/uiprogress v0.0.0-20160202012259-a9f819bfc744
	github.com/jarcoal/httpmock v1.0.4
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jrallison/go-workers v0.0.0-20180112190529-dbf81d0b75bb
	github.com/jteeuwen/go-bindata v3.0.7+incompatible // indirect
	github.com/klauspost/compress v0.0.0-20161025140425-8df558b6cb6f // indirect
	github.com/klauspost/cpuid v0.0.0-20160302075316-09cded8978dc // indirect
	github.com/klauspost/crc32 v0.0.0-20161016154125-cb6bfca970f6 // indirect
	github.com/kylelemons/go-gypsy v0.0.0-20160905020020-08cad365cd28 // indirect
	github.com/labstack/echo v2.2.0+incompatible
	github.com/lib/pq v1.0.0
	github.com/mailru/easyjson v0.7.7
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.11.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/rcrowley/go-metrics v0.0.0-20180125231941-8732c616f529
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.4.0
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/topfreegames/extensions/v9 v9.0.0
	github.com/topfreegames/goose v0.0.0-20160616205307-c7f6dd34057c
	github.com/uber-go/atomic v1.0.0 // indirect
	github.com/uber-go/zap v0.0.0-20160809182253-d11d2851fcab
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/valyala/fasthttp v0.0.0-20160818100357-834fb48f1040 // indirect
	github.com/valyala/fasttemplate v0.0.0-20170224212429-dcecefd839c4
	github.com/wildlife-studios/crypto v1.1.0
	github.com/ziutek/mymysql v1.5.5-0.20160909221029-df6241f6355c // indirect
	go.uber.org/atomic v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/tools v0.1.4 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/olivere/elastic.v5 v5.0.66
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/jrallison/go-workers v1.0.0 => github.com/topfreegames/go-workers v1.0.0

replace github.com/codahale/hdrhistogram => github.com/HdrHistogram/hdrhistogram-go v0.0.0-20200919145931-8dac23c8dac1
