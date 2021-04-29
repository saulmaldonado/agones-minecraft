package controller_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Node Controller", func() {
	const (
		Timeout  = time.Second * 10
		Duration = time.Second * 10
		Interval = time.Millisecond * 250
	)

	var (
		NodeName string          = "mc-node"
		ctx      context.Context = context.Background()
	)

	Context("When creating Node", func() {
		It("Should create a new DNS record for Nodes", func() {
			By("By createing a new Node")

			node := &corev1.Node{
				TypeMeta: v1.TypeMeta{
					Kind:       "Node",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      NodeName,
					Namespace: corev1.NamespaceDefault,
					Labels: map[string]string{
						"agones-mc/domain": "saulmaldonado.me",
					},
				},
			}

			Expect(testClient.Create(ctx, node)).Should(Succeed())

			By("Checking for agones-mc/externalDNS annotation")
			createdNode := &corev1.Node{}
			nodeKey := types.NamespacedName{Namespace: corev1.NamespaceDefault, Name: NodeName}

			Eventually(func() bool {
				if err := testClient.Get(ctx, nodeKey, createdNode); err != nil {
					return false
				}
				return createdNode.Annotations["agones-mc/externalDNS"] == "mc-node.saulmaldonado.me."
			}, Timeout, Interval).Should(BeTrue())
		})

		It("Should remove DNS records for deleted Nodes", func() {
			By("Removing Node")

			node := &corev1.Node{
				TypeMeta: v1.TypeMeta{
					Kind:       "Node",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      NodeName,
					Namespace: corev1.NamespaceDefault,
					Labels: map[string]string{
						"agones-mc/domain": "saulmaldonado.me",
					},
				},
			}

			Expect(testClient.Delete(ctx, node)).Should(Succeed())

			By("Checking if node is deleted")
			nodeKey := types.NamespacedName{Namespace: corev1.NamespaceDefault, Name: NodeName}
			createdNode := corev1.Node{}

			Eventually(func() error {
				return testClient.Get(ctx, nodeKey, &createdNode)
			}, Timeout, Interval).ShouldNot(Succeed())

			By("Checking mock DNS store for DNS record")
			Eventually(func() bool {
				for _, record := range FakeDns.DnsRecords {
					if record == "mc-node.saulmaldonado.me." {
						return true
					}
				}
				return false
			}, Timeout, Interval).Should(BeFalse())
		})
	})
})
