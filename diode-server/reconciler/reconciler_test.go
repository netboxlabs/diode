package reconciler_test

import (
	"net"
	"os"
	"strconv"
)

func getFreePort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return strconv.Itoa(0), err
	}

	addr := listener.Addr().(*net.TCPAddr)

	if err = listener.Close(); err != nil {
		return strconv.Itoa(0), err
	}
	return strconv.Itoa(addr.Port), nil
}

func setupEnv(redisAddr string) {
	host, port, _ := net.SplitHostPort(redisAddr)
	grpcPort, _ := getFreePort()
	_ = os.Setenv("GRPC_PORT", grpcPort)
	_ = os.Setenv("REDIS_HOST", host)
	_ = os.Setenv("REDIS_PORT", port)
	_ = os.Setenv("REDIS_USERNAME", "")
	_ = os.Setenv("REDIS_PASSWORD", "")
	_ = os.Setenv("REDIS_DB", "0")
	_ = os.Setenv("REDIS_STREAM_DB", "1")
	_ = os.Setenv("MIGRATION_ENABLED", "false")
	_ = os.Setenv("NETBOX_API_URL", "http://example.com")
	_ = os.Setenv("AUTO_APPLY_CHANGESETS", "true")
	_ = os.Setenv("RECONCILER_RATE_LIMITER_RPS", "20")
	_ = os.Setenv("RECONCILER_RATE_LIMITER_BURST", "1")
	_ = os.Setenv("POSTGRES_HOST", "localhost")
	_ = os.Setenv("POSTGRES_PORT", "5432")
	_ = os.Setenv("POSTGRES_DB_NAME", "diode")
	_ = os.Setenv("POSTGRES_USER", "diode")
	_ = os.Setenv("POSTGRES_PASSWORD", "diode")
	_ = os.Setenv("NETBOX_DIODE_PLUGIN_API_BASE_URL", "http://127.0.0.1:8080/api/plugins/diode")
	_ = os.Setenv("DIODE_AUTH_TOKEN_URL", "http://diode-auth:8080/diode/auth/token")
	_ = os.Setenv("DIODE_TO_NETBOX_CLIENT_ID", "not-real-client-id")
	_ = os.Setenv("DIODE_TO_NETBOX_CLIENT_SECRET", "not-real-client-secret")
}

func teardownEnv() {
	_ = os.Unsetenv("GRPC_PORT")
	_ = os.Unsetenv("REDIS_HOST")
	_ = os.Unsetenv("REDIS_PORT")
	_ = os.Unsetenv("REDIS_USERNAME")
	_ = os.Unsetenv("REDIS_PASSWORD")
	_ = os.Unsetenv("REDIS_DB")
	_ = os.Unsetenv("REDIS_STREAM_DB")
	_ = os.Unsetenv("MIGRATION_ENABLED")
	_ = os.Unsetenv("NETBOX_API_URL")
	_ = os.Unsetenv("AUTO_APPLY_CHANGESETS")
	_ = os.Unsetenv("RECONCILER_RATE_LIMITER_RPS")
	_ = os.Unsetenv("RECONCILER_RATE_LIMITER_BURST")
	_ = os.Unsetenv("POSTGRES_HOST")
	_ = os.Unsetenv("POSTGRES_PORT")
	_ = os.Unsetenv("POSTGRES_DB_NAME")
	_ = os.Unsetenv("POSTGRES_USER")
	_ = os.Unsetenv("POSTGRES_PASSWORD")
	_ = os.Unsetenv("NETBOX_DIODE_PLUGIN_API_BASE_URL")
	_ = os.Unsetenv("DIODE_AUTH_TOKEN_URL")
	_ = os.Unsetenv("DIODE_TO_NETBOX_CLIENT_ID")
	_ = os.Unsetenv("DIODE_TO_NETBOX_CLIENT_SECRET")
}
