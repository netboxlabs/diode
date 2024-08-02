package reconciler_test

import (
	"net"
	"os"
	"strconv"
)

const bufSize = 1024 * 1024

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
	_ = os.Setenv("REDIS_PASSWORD", "")
	_ = os.Setenv("REDIS_DB", "0")
	_ = os.Setenv("REDIS_STREAM_DB", "1")
	_ = os.Setenv("NETBOX_API_URL", "http://example.com")
	_ = os.Setenv("DIODE_TO_NETBOX_API_KEY", "diode_to_netbox_api_key")
	_ = os.Setenv("NETBOX_TO_DIODE_API_KEY", "netbox_to_diode_api_key")
	_ = os.Setenv("DIODE_API_KEY", "diode_api_key")
}

func teardownEnv() {
	_ = os.Unsetenv("GRPC_PORT")
	_ = os.Unsetenv("REDIS_HOST")
	_ = os.Unsetenv("REDIS_PORT")
	_ = os.Unsetenv("REDIS_PASSWORD")
	_ = os.Unsetenv("REDIS_DB")
	_ = os.Unsetenv("REDIS_STREAM_DB")
	_ = os.Unsetenv("NETBOX_API_URL")
	_ = os.Unsetenv("DIODE_TO_NETBOX_API_KEY")
	_ = os.Unsetenv("NETBOX_TO_DIODE_API_KEY")
	_ = os.Unsetenv("DIODE_API_KEY")
}
