package controllers

import (
	"apache-operator/api/v1alpha1"

	discovery "k8s.io/api/discovery/v1"
)

// The function-helper is intendet to generate
// Apache HTTPD config pattern by using EndpointSlice resource
// Output of this function is is array of v1alpha1.EndPoint
func genBackEndsList(backEndServiceName, proto string, endPointSliceList discovery.EndpointSliceList) ([]v1alpha1.EndPoint, string) {
	var endPointSlice discovery.EndpointSlice
	var endPointsList = make([]v1alpha1.EndPoint, 0)
	var endPointsListVers string

	// Retrieve EndPointSlice which is related to ApacheWeb resource: label "kubernetes.io/service-name" is equal to
	// apacheWeb.Spec.LoadBalancer.BackEndService property
	for _, i := range endPointSliceList.Items {
		if i.Labels["kubernetes.io/service-name"] == backEndServiceName {
			endPointSlice = i
			endPointsListVers = i.GetResourceVersion()
			break
		}
	}

	for _, e := range endPointSlice.Endpoints {
		for _, ip := range e.Addresses {
			if ip == "" {
				continue
			}

			for _, p := range endPointSlice.Ports {
				if p.Port == nil {
					continue
				}
				endPoint := v1alpha1.EndPoint{
					Port:      *p.Port,
					IPAddress: ip,
					Proto:     proto,
					Status:    *e.Conditions.Ready,
				}
				endPointsList = append(endPointsList, endPoint)
			}
		}
	}

	return endPointsList, endPointsListVers
}
