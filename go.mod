module gojob

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/boltdb/bolt v1.3.1
	github.com/gin-gonic/gin v1.7.0
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-sql-driver/mysql v1.4.1
	github.com/go-xorm/core v0.6.2
	github.com/go-xorm/xorm v0.7.3
	github.com/hashicorp/go-hclog v0.9.1
	github.com/hashicorp/raft v1.1.1
	github.com/hashicorp/raft-boltdb v0.0.0-20191021154308-4207f1bf0617
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/rakyll/statik v0.1.6
	github.com/robfig/cron v1.2.0
	github.com/satori/go.uuid v1.2.0
	github.com/sony/sonyflake v1.0.0
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	go.uber.org/atomic v1.4.0
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43

replace github.com/hashicorp/raft v1.1.1 => github.com/wj596/raft v1.1.2-0.20200102091329-34daf92ee810
