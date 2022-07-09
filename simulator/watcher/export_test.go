package watcher

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func NewStreamWriter(stream ResponseStream) *streamWriter {
	return newStreamWriter(stream)
}

func CreateWatcher(p *resourceEventProxy) (watch.Interface, error) {
	return createWatcher(p)
}

func NewResourceEventProxy(sw StreamWriter, c cache.Getter, r resourceKind, o runtime.Object, ilrv string) *resourceEventProxy {
	return newresourceEventProxy(sw, c, r, o, ilrv)
}

func (p *resourceEventProxy) ExportGetLastResourceVersion() string {
	return p.lastResourceVersion
}
