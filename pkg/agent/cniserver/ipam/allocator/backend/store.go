package backend

import "net"

type Store interface {
	Lock() error
	Unlock() error
	Close() error
	Reserve(id string, ifname string, ip net.IP, rangeID string) (bool, error)
	LastReservedIP(rangeID string) (net.IP, error)
	Release(ip net.IP) error
	ReleaseByID(id string, ifname string) error
	GetByID(id string, ifname string) []net.IP
}
