package ipfslite

import (
	"context"
	"os"
	"time"

	"github.com/libp2p/go-tcp-transport"

	"github.com/ipfs/go-datastore"
	leveldb "github.com/ipfs/go-ds-leveldb"
	ipfsconfig "github.com/ipfs/go-ipfs-config"
	ipns "github.com/ipfs/go-ipns"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
	routing "github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dualdht "github.com/libp2p/go-libp2p-kad-dht/dual"
	record "github.com/libp2p/go-libp2p-record"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/multiformats/go-multiaddr"
)

// DefaultBootstrapPeers returns the default go-ipfs bootstrap peers (for use
// with NewLibp2pHost.
func DefaultBootstrapPeers() []peer.AddrInfo {
	defaults, _ := ipfsconfig.DefaultBootstrapPeers()
	return defaults
}

// LevelDatastore returns a new instance of LevelDB persisting
// to the given path with the default options.
func LevelDatastore(path string) (datastore.Batching, error) {
	return leveldb.NewDatastore(path, &leveldb.Options{})
}

// Libp2pOptionsExtra provides some useful libp2p options
// to create a fully featured libp2p host. It can be used with
// SetupLibp2p.
var Libp2pOptionsExtra = []libp2p.Option{
	libp2p.Transport(tcp.NewTCPTransport),
	libp2p.DisableRelay(),
	libp2p.NATPortMap(),
	libp2p.ConnectionManager(connmgr.NewConnManager(100, 600, time.Minute)),
	libp2p.EnableNATService(),
	libp2p.Security(libp2ptls.ID, libp2ptls.New),
}

// SetupLibp2p returns a routed host and DHT instances that can be used to
// easily create a ipfslite Peer. You may consider to use Peer.Bootstrap()
// after creating the IPFS-Lite Peer to connect to other peers. When the
// datastore parameter is nil, the DHT will use an in-memory datastore, so all
// provider records are lost on program shutdown.
//
// Additional libp2p options can be passed. Note that the Identity,
// ListenAddrs and PrivateNetwork options will be setup automatically.
// Interesting options to pass: NATPortMap() EnableAutoRelay(),
// libp2p.EnableNATService(), DisableRelay(), ConnectionManager(...)... see
// https://godoc.org/github.com/libp2p/go-libp2p#Option for more info.
//
// The secret should be a 32-byte pre-shared-key byte slice.
func SetupLibp2p(
	ctx context.Context,
	hostKey crypto.PrivKey,
	listenAddrs []multiaddr.Multiaddr,
	ds datastore.Batching,
	opts ...libp2p.Option,
) (host.Host, *dualdht.DHT, error) {

	var ddht *dualdht.DHT
	var err error

	finalOpts := []libp2p.Option{
		libp2p.Identity(hostKey),
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			ddht, err = newDHT(ctx, h, ds)
			return ddht, err
		}),
	}
	finalOpts = append(finalOpts, opts...)

	h, err := libp2p.New(
		ctx,
		finalOpts...,
	)
	if err != nil {
		return nil, nil, err
	}

	return h, ddht, nil
}

func newDHT(ctx context.Context, h host.Host, ds datastore.Batching) (*dualdht.DHT, error) {
	dhtOpts := []dualdht.Option{
		dualdht.DHTOption(dht.NamespacedValidator("pk", record.PublicKeyValidator{})),
		dualdht.DHTOption(dht.NamespacedValidator("ipns", ipns.Validator{KeyBook: h.Peerstore()})),
		dualdht.DHTOption(dht.Concurrency(10)),
		dualdht.DHTOption(dht.Mode(dht.ModeAuto)),
	}
	if ds != nil {
		dhtOpts = append(dhtOpts, dualdht.DHTOption(dht.Datastore(ds)))
	}

	return dualdht.New(ctx, h, dhtOpts...)
}

// FileExists check if the file with the given path exits.
func FileExists(filename string) bool {
	fi, err := os.Lstat(filename)
	if fi != nil || (err != nil && !os.IsNotExist(err)) {
		return true
	}
	return false
}
