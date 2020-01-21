module github.com/lyft/flyteadmin

go 1.13

require (
	cloud.google.com/go v0.51.0 // indirect
	github.com/NYTimes/gizmo v0.4.3
	github.com/Selvatico/go-mocket v1.0.7
	github.com/aws/aws-sdk-go v1.28.2
	github.com/benbjohnson/clock v1.0.0
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b // indirect
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/fatih/color v1.9.0 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/securecookie v1.1.1
	github.com/graymeta/stow v0.0.0-20190522170649-903027f87de7
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/lib/pq v1.3.0
	github.com/lyft/flyteidl v0.16.5
	github.com/lyft/flytepropeller v0.1.30
	github.com/lyft/flytestdlib v0.2.31
	github.com/magiconair/properties v1.8.1
	github.com/mitchellh/mapstructure v1.1.2
	github.com/ncw/swift v1.0.49-0.20191117165619-017f012e58fa // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/client_model v0.1.0 // indirect
	github.com/prometheus/common v0.8.0 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.1 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/crypto v0.0.0-20200109152110-61a87790db17 // indirect
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200113162924-86b910548bc1 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/api v0.15.1-0.20200114000728-abe50cf84d67 // indirect
	google.golang.org/genproto v0.0.0-20200113173426-e1de0a7b01eb // indirect
	google.golang.org/grpc v1.26.0
	gopkg.in/gormigrate.v1 v1.6.0
	gopkg.in/ini.v1 v1.51.1 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
	k8s.io/api v0.0.0-20191031200350-b49a72c274e0 // indirect
	k8s.io/apimachinery v0.0.0-20191031200210-047e3ea32d7f
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/utils v0.0.0-20200109141947-94aeca20bf09 // indirect
	sigs.k8s.io/controller-runtime v0.3.1-0.20191029211253-40070e2a1958
)

replace (
	github.com/vektra/mockery => github.com/enghabu/mockery v0.0.0-20191009061720-9d0c8670c2f0
	k8s.io/api => github.com/lyft/api v0.0.0-20191031200350-b49a72c274e0
	k8s.io/apimachinery => github.com/lyft/apimachinery v0.0.0-20191031200210-047e3ea32d7f
)
