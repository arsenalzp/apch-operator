package controllers

import (
	"apache-operator/api/v1alpha1"
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Apacheweb controller", func() {
	const (
		ApachewebName      = "test-apacheweb"
		ApachewebNamespace = "default"
		ApachewebSize      = 1
		timeout            = time.Second * 10
		duration           = time.Second * 10
		interval           = time.Millisecond * 250
	)

	var serverPort int32 = 8888
	Context("When creating Apacheweb resource", func() {
		It("should be created", func() {
			By("by operator")
			ctx := context.Background()
			apacheweb := &v1alpha1.Apacheweb{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ApachewebName,
					Namespace: ApachewebNamespace,
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "Apacheweb",
					APIVersion: "apacheweb.arsenal.dev/v1alpha1",
				},
				Spec: v1alpha1.ApachewebSpec{
					Size:       1,
					ServerName: ApachewebName,
					Type:       "lb",
					LoadBalancer: &v1alpha1.LoadBalancer{
						Proto:          "https",
						ServerPort:     &serverPort,
						BackEndService: ApachewebName,
					},
				},
			}
			Expect(k8sClient.Create(ctx, apacheweb)).Should(Succeed())

			apachewebLookupKey := types.NamespacedName{Name: ApachewebName, Namespace: ApachewebNamespace}
			createdApacheweb := &v1alpha1.Apacheweb{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, apachewebLookupKey, createdApacheweb)

				if err != nil {
					return err == nil
				}

				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdApacheweb.Spec.Size).Should(Equal(int32(1)))

			By("checking a service created by controller")
			foundService := &corev1.Service{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, apachewebLookupKey, foundService)

				if err != nil {
					return err == nil
				}

				return true
			}, timeout, interval).Should(BeTrue())
			Expect(foundService.Spec.Selector).Should(Equal(map[string]string{"servername": ApachewebName}))
		})
	})
})
