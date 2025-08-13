module photo-processing-server

go 1.21

require (
	gocv.io/x/gocv v0.34.0
	github.com/lib/pq v1.10.9
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.1
	github.com/gin-gonic/gin v1.9.1
	github.com/gin-contrib/cors v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/sirupsen/logrus v1.9.3
	github.com/golang-migrate/migrate/v4 v4.16.2
	github.com/google/uuid v1.4.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)