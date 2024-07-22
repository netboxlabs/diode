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
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return strconv.Itoa(addr.Port), nil
}

func setupEnv(redisAddr string) {
	host, port, _ := net.SplitHostPort(redisAddr)
	grpc_port, _ := getFreePort()
	os.Setenv("GRPC_PORT", grpc_port)
	os.Setenv("REDIS_HOST", host)
	os.Setenv("REDIS_PORT", port)
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("REDIS_DB", "0")
	os.Setenv("REDIS_STREAM_DB", "1")
	os.Setenv("NETBOX_API_URL", "http://example.com")
	os.Setenv("DIODE_TO_NETBOX_API_KEY", "diode_to_netbox_api_key")
	os.Setenv("NETBOX_TO_DIODE_API_KEY", "netbox_to_diode_api_key")
	os.Setenv("DIODE_API_KEY", "diode_api_key")
}

func teardownEnv() {
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("REDIS_DB")
	os.Unsetenv("REDIS_STREAM_DB")
	os.Unsetenv("NETBOX_API_URL")
	os.Unsetenv("DIODE_TO_NETBOX_API_KEY")
	os.Unsetenv("NETBOX_TO_DIODE_API_KEY")
	os.Unsetenv("DIODE_API_KEY")
}
