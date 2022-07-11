package resourcewatcher

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type ExportStreamWriter streamWriter

func NewStreamWriter(stream ResponseStream) *ExportStreamWriter {
	return (*ExportStreamWriter)(newStreamWriter(stream))
}

func (sw *ExportStreamWriter) Write(we *WatchEvent) error {
	return (*streamWriter)(sw).Write(we)
}

func CreateWatcher(p *ExportResourceEventProxy) (watch.Interface, error) {
	return createWatcher((*resourceEventProxy)(p))
}

type ExportResourceEventProxy resourceEventProxy

func NewResourceEventProxy(sw StreamWriter, c cache.Getter, r resourceKind, o runtime.Object, ilrv string) *ExportResourceEventProxy {
	return (*ExportResourceEventProxy)(newresourceEventProxy(sw, c, r, o, ilrv))
}

func (p *ExportResourceEventProxy) ExportGetLastResourceVersion() string {
	return p.lastResourceVersion
}

func (p *ExportResourceEventProxy) WatchHandlerFunc(watcher watch.Interface) func(stopCh <-chan struct{}) error {
	return (*resourceEventProxy)(p).WatchHandlerFunc(watcher)
}
