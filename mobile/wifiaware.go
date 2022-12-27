package datahop

type WifiAwareClient interface {
	Connect(peerId string)
	Disconnect()
	Host() string
}

type WifiAwareServer interface {
	Start() //(string, string)
	Stop()
}

type WifiAwareServerNotifier interface {
	StartOnSuccess()
	StartOnFailure(code int)
	StopOnSuccess()
	StopOnFailure(code int)
	NetworkReady(ip string, port int)
	//NetworkInfo(network string, password string)
	//ClientsConnected(num int)
}

type WifiAwareClientNotifier interface {
	//OnConnectionSuccess(started, completed int64, rssi, speed, freq int)
	OnConnectionFailure(code int, started, failed int64)
	OnConnectionSuccess(ip string, port int, peerId string)
	OnDisconnect()
}
