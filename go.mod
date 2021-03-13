module atomizer.io/test-console

go 1.16

require (
	atomizer.io/amqp v1.0.5
	atomizer.io/engine v1.0.3
	atomizer.io/montecarlopi v0.0.1
	devnw.com/alog v1.0.6
	devnw.com/validator v1.0.4 // indirect
	github.com/google/uuid v1.2.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/streadway/amqp v1.0.0 // indirect
)

replace atomizer.io/amqp => ../amqp
