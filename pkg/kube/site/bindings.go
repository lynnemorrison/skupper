package site

import (
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"

	skupperv1alpha1 "github.com/skupperproject/skupper/pkg/apis/skupper/v1alpha1"
	"github.com/skupperproject/skupper/pkg/kube"
	"github.com/skupperproject/skupper/pkg/qdr"
	"github.com/skupperproject/skupper/pkg/site"
)

type BindingContext interface {
	Select(connector *skupperv1alpha1.Connector) *TargetSelection
	Expose(ports *ExposedPortSet)
	Unexpose(host string)
}

type BindingAdaptor struct {
	context   BindingContext
	mapping   *qdr.PortMapping
	exposed   ExposedPorts
	selectors map[string]*TargetSelection
}

func (a *BindingAdaptor) init(context BindingContext, config *qdr.RouterConfig) {
	a.context = context
	if a.mapping == nil {
		a.mapping = qdr.RecoverPortMapping(config)
	}
	a.exposed = ExposedPorts{}
	a.selectors = map[string]*TargetSelection{}
}

func (a *BindingAdaptor) cleanup() {
	for _, s := range a.selectors {
		s.Close()
	}
}

func (a *BindingAdaptor) ConnectorUpdated(connector *skupperv1alpha1.Connector, specChanged bool) bool {
	if !specChanged {
		if connector.Spec.Selector != "" {
			if selector, ok := a.selectors[connector.Name]; ok {
				selector.connector = connector // need to update to latest resource regardless of spec change
			} else {
				log.Printf("Warning: spec for connector %s/%s has not changed but pods not tracked", connector.Namespace, connector.Name)
				a.selectors[connector.Name] = a.context.Select(connector)
			}
		}
		return false
	}
	if selector, ok := a.selectors[connector.Name]; ok {
		selectorChanged := selector.connector.Spec != connector.Spec
		selector.connector = connector
		if !selectorChanged {
			// don't need to change the pod watcher, but may need to reconfigure for other change to spec
			return true
		} else {
			// selector has changed so need to close current pod watcher
			selector.Close()
			if connector.Spec.Selector == "" {
				// no longer using a selector, so just delete the old watcher
				delete(a.selectors, connector.Name)
				return true
			}
			// else create a new watcher below
		}
	} else if connector.Spec.Selector == "" {
		return true
	}
	a.selectors[connector.Name] = a.context.Select(connector)
	// can't yet update configuration; need to wait for the new
	// watcher to return any matching pods and update config at
	// that point
	return false
}

func (a *BindingAdaptor) ConnectorDeleted(connector *skupperv1alpha1.Connector) {
	if current, ok := a.selectors[connector.Name]; ok {
		current.Close()
		delete(a.selectors, connector.Name)
	}
}

func (a *BindingAdaptor) ListenerUpdated(listener *skupperv1alpha1.Listener) {
	allocatedRouterPort, err := a.mapping.GetPortForKey(listener.Name)
	if err != nil {
		log.Printf("Unable to get port for listener %s/%s: %s", listener.Namespace, listener.Name, err)
	} else {
		port := Port{
			Name:       listener.Name,
			Port:       listener.Spec.Port,
			TargetPort: allocatedRouterPort,
			Protocol:   listener.Protocol(),
		}
		if exposed := a.exposed.Expose(listener.Spec.Host, port); exposed != nil {
			a.context.Expose(exposed)
		}
	}
}

func (a *BindingAdaptor) ListenerDeleted(listener *skupperv1alpha1.Listener) {
	a.context.Unexpose(listener.Spec.Host)
	a.mapping.ReleasePortForKey(listener.Name)
}

func (a *BindingAdaptor) updateBridgeConfigForConnector(siteId string, connector *skupperv1alpha1.Connector, config *qdr.BridgeConfig) {
	if connector.Spec.Host != "" {
		site.UpdateBridgeConfigForConnectorWithHost(siteId, connector, connector.Spec.Host, config)
	} else if connector.Spec.Selector != "" {
		if selector, ok := a.selectors[connector.Name]; ok {
			for podUID, host := range selector.List() {
				site.UpdateBridgeConfigForConnectorWithHostProcess(siteId, connector, host, podUID, config)
			}
		} else {
			log.Printf("Error: not yet tracking pods for connector %s/%s with selector set", connector.Namespace, connector.Name)
		}
	} else {
		log.Printf("Error: connector %s/%s has neither host nor selector set", connector.Namespace, connector.Name)
	}
}

func (a *BindingAdaptor) updateBridgeConfigForListener(siteId string, listener *skupperv1alpha1.Listener, config *qdr.BridgeConfig) {
	if port, err := a.mapping.GetPortForKey(listener.Name); err == nil {
		site.UpdateBridgeConfigForListenerWithHostAndPort(siteId, listener, "0.0.0.0", port, config)
	} else {
		log.Printf("Could not allocate port for %s/%s: %s", listener.Namespace, listener.Name, err)
	}
}

type TargetSelection struct {
	watcher         *kube.PodWatcher
	stopCh          chan struct{}
	site            *Site
	connector       *skupperv1alpha1.Connector
	name            string
	namespace       string
	includeNotReady bool
}

func (w *TargetSelection) Close() {
	close(w.stopCh)
}

func (w *TargetSelection) List() map[string]string {
	pods := w.watcher.List()
	targets := make(map[string]string, len(pods))

	for _, pod := range pods {
		if kube.IsPodReady(pod) || w.includeNotReady {
			if kube.IsPodRunning(pod) && pod.DeletionTimestamp == nil {
				log.Printf("Pod %s selected for connector %s in %s", pod.ObjectMeta.Name, w.name, w.namespace)
				targets[string(pod.UID)] = pod.Status.PodIP
			} else {
				log.Printf("Pod %s not running for connector %s in %s", pod.ObjectMeta.Name, w.name, w.namespace)
			}
		} else {
			log.Printf("Pod %s not ready for connector %s in %s", pod.ObjectMeta.Name, w.name, w.namespace)
		}
	}
	return targets

}

func (w *TargetSelection) handle(key string, pod *corev1.Pod) error {
	err := w.site.updateRouterConfigForGroups(w.site.bindings)
	if err != nil {
		return w.site.updateConnectorStatus(w.connector, err)
	}
	if len(w.List()) == 0 {
		log.Printf("No pods available for %s/%s", w.connector.Namespace, w.connector.Name)
		return w.site.updateConnectorStatus(w.connector, fmt.Errorf("No targets for selector"))
	}
	log.Printf("Pods are available for %s/%s", w.connector.Namespace, w.connector.Name)
	return w.site.updateConnectorStatus(w.connector, nil)
}