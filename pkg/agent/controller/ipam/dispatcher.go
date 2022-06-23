package ipam

import (
	"strings"
	"sync"

	"k8s.io/klog/v2"
)

var singleDispatcher *Dispatcher
var once sync.Once

type handlerFunc func(*Controller, string) error

// Dispatcher can dispatch key to handle func
type Dispatcher struct {
	handlersPool map[string]handlerFunc
}

// Register put handle func into a map by key
func (dispatcher *Dispatcher) Register(name string, handler handlerFunc) {
	dispatcher.handlersPool[name] = handler
}

// DoHandler invoke the real handler func
func (dispatcher *Dispatcher) DoHandler(name string, controller *Controller) error {
	// handlerFunc := dispatcher.handlersPool[name]
	// handlerFunc(controller)
	klog.Info("dispatch name:", name)
	parts := strings.Split(name, ":")
	handlerFunc := dispatcher.handlersPool[parts[0]]
	handlerFunc(controller, parts[1])
	return nil
}

// GetDispatcher return the single instance dispatcher
func GetDispatcher() *Dispatcher {
	once.Do(func() {
		singleDispatcher = &Dispatcher{
			handlersPool: make(map[string]handlerFunc),
		}
	},
	)
	return singleDispatcher
}

func init() {
	dispatcher := GetDispatcher()
	dispatcher.Register("namespace/add", addNamespaceHandler)
	dispatcher.Register("namespace/del", delNamespaceHandler)
	dispatcher.Register("namespace/update", updateNamespaceHandler)

	dispatcher.Register("subnet/add", addSubnetHandler)
	dispatcher.Register("subnet/update", updateSubnetHandler)
	dispatcher.Register("subnet/del", delSubnetHandler)
	klog.Info("register dispathcer:", dispatcher.handlersPool)
}
