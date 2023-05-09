package controllers

import (
	"apache-operator/api/v1alpha1"
	"reflect"
	"testing"

	discovery "k8s.io/api/discovery/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	condition         bool = true
	expectedEndPoints      = []v1alpha1.EndPoint{
		{
			IPAddress: "1.1.1.1",
			Proto:     "https",
			Port:      8888,
			Status:    true,
		},
		{
			IPAddress: "2.2.2.2",
			Proto:     "https",
			Port:      8888,
			Status:    true,
		},
	}
	backEndServiceName         = "apacheWebTest"
	expectedResourceVers       = "testResourceVersion"
	proto                      = "https"
	port                 int32 = 8888
	endPointsSlice             = discovery.EndpointSliceList{
		TypeMeta: v1.TypeMeta{
			Kind:       "EndpointSliceList",
			APIVersion: "discovery.k8s.io/v1",
		},
		ListMeta: v1.ListMeta{},
		Items: []discovery.EndpointSlice{
			{
				TypeMeta: v1.TypeMeta{
					Kind:       "EndpointSlice",
					APIVersion: "discovery.k8s.io/v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:            "testEndpointSlice",
					Labels:          map[string]string{"kubernetes.io/service-name": backEndServiceName},
					ResourceVersion: expectedResourceVers,
				},
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{"1.1.1.1", "2.2.2.2"},
						Conditions: discovery.EndpointConditions{
							Ready: &condition,
						},
					},
				},
				Ports: []discovery.EndpointPort{
					{
						Port: &port,
					},
					{
						Port: &port,
					},
				},
			},
		},
	}
)

func TestgenBackEndsList(t *testing.T) {
	outEndPoints, outResourceVers := genBackEndsList(backEndServiceName, proto, endPointsSlice)
	if reflect.DeepEqual(outEndPoints, expectedEndPoints) || outResourceVers != expectedResourceVers {
		t.Errorf("Output %v not equal to expected %v", outEndPoints, outResourceVers)
	}
}
