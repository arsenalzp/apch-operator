package controllers

import (
	"apache-operator/api/v1alpha1"

	discovery "k8s.io/api/discovery/v1"
)

// The function-helper to generate Apache HTTPD pattern for load balance
// from EndpointSlice object - proto://ip:port
func genBackEndsList(proto string, es discovery.EndpointSlice) []v1alpha1.EndPoint {
	var epl = make([]v1alpha1.EndPoint, 0)
	for _, e := range es.Endpoints {
		for _, ip := range e.Addresses {
			if ip == "" {
				continue
			}
			for _, p := range es.Ports {
				if p.Port == nil {
					continue
				}
				ep := v1alpha1.EndPoint{
					Port:      *p.Port,
					IPAddress: ip,
					Proto:     proto,
					Status:    *e.Conditions.Ready,
				}
				epl = append(epl, ep)
			}
		}
	}

	return epl
}
