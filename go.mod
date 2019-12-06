module github.com/lyft/flyteadmin

go 1.13

require (
	cloud.google.com/go v0.47.0 // indirect
	github.com/Azure/azure-sdk-for-go v10.2.1-beta+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.9.2 // indirect
	github.com/NYTimes/gizmo v0.4.3
	github.com/Selvatico/go-mocket v1.0.7
	github.com/aws/aws-sdk-go v1.25.30
	github.com/benbjohnson/clock v1.0.0
	github.com/benlaurie/objecthash v0.0.0-20180202135721-d1e3d6079fc1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b // indirect
	github.com/cheekybits/is v0.0.0-20150225183255-68e9c0620927 // indirect
	github.com/coocood/freecache v1.1.0 // indirect
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/dnaeon/go-vcr v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.4.8-0.20191012010759-4bf2d1fec783 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/securecookie v1.1.1
	github.com/graymeta/stow v0.0.0-20190522170649-903027f87de7
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/grpc/grpc-go v1.24.0
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jinzhu/gorm v1.9.11
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lib/pq v1.2.0
	github.com/lyft/flyteidl v0.16.1
	github.com/lyft/flyteplugins v0.1.9 // indirect
	github.com/lyft/flytepropeller v0.1.12
	github.com/lyft/flytestdlib v0.2.28
	github.com/magiconair/properties v1.8.1
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/ncw/swift v1.0.49-0.20190728102658-a24ef33bc9b7 // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/satori/uuid v1.2.0 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/crypto v0.0.0-20191107222254-f4817d981bb6 // indirect
	golang.org/x/net v0.0.0-20191108063844-7e6e90b9ea88 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/api v0.13.1-0.20191107195140-22b92ae6b4f3 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6 // indirect
	google.golang.org/grpc v1.25.0
	gopkg.in/gormigrate.v1 v1.6.0
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	gopkg.in/yaml.v2 v2.2.5 // indirect
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190820062731-7e43eff7c80a+incompatible
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0-20191030222137-2b95a09bc58d // indirect
	sigs.k8s.io/controller-runtime v0.2.2
)

replace github.com/grpc/grpc-go v1.24.0 => google.golang.org/grpc v1.24.0
