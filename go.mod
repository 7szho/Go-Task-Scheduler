module crony

go 1.22.2

require (
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/spf13/viper v1.7.1
	go.uber.org/zap v1.27.0
	gorm.io/driver/mysql v1.6.0
	gorm.io/gorm v1.30.0
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20240122114842-bbd7aa9bf6fb // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a // indirect
	google.golang.org/grpc v1.26.0 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/ini.v1 v1.51.0 // indirect
	gopkg.in/yaml.v2 v2.2.4 // indirect
	sigs.k8s.io/yaml v1.5.0 // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/coreos/etcd v3.3.27+incompatible
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.10
	google.golang.org/genproto/googleapis/api => google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a
	google.golang.org/genproto/googleapis/rpc => google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a
)
