package controller_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
)

var _ = Describe("GameServer Controller", func() {

	const (
		Timeout  = time.Second * 10
		Duration = time.Second * 10
		Interval = time.Millisecond * 250
	)

	var (
		GameServerContainer string          = "mc-server"
		GameServerName      string          = "mc-server"
		GameServerNodeName  string          = "mc-node"
		GameServerPort      int32           = 7000
		ctx                 context.Context = context.Background()
	)

	Context("When creating GameServer", func() {
		It("Should create a new DNS record for GameServers", func() {
			By("createing a new GameServer")

			gs := &agonesv1.GameServer{
				Status: agonesv1.GameServerStatus{State: agonesv1.GameServerStateScheduled,
					NodeName: GameServerNodeName,
					Ports: []agonesv1.GameServerStatusPort{
						{Name: "mc", Port: GameServerPort},
					},
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "GameServer",
					APIVersion: agonesv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"agones-mc/domain": "saulmaldonado.me",
					},
					Name:      GameServerName,
					Namespace: metav1.NamespaceDefault,
				},
				Spec: agonesv1.GameServerSpec{
					Container: GameServerContainer,
					Ports: []agonesv1.GameServerPort{
						{
							Name:          "mc",
							PortPolicy:    "Dynamic",
							Container:     &GameServerContainer,
							ContainerPort: 25565,
							Protocol:      "TCP",
						},
					},
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  GameServerContainer,
									Image: "itzg/minecraft-server",
								},
							},
						},
					},
				},
			}

			Expect(testClient.Create(ctx, gs)).Should(Succeed())

			By("Checking for agones-mc/externalDNS annotation")

			gameServerKey := types.NamespacedName{Namespace: metav1.NamespaceDefault, Name: GameServerName}
			createdGameServer := &agonesv1.GameServer{}

			Eventually(func() bool {
				if err := testClient.Get(ctx, gameServerKey, createdGameServer); err != nil {
					return false
				}
				return createdGameServer.Annotations["agones-mc/externalDNS"] == "mc-server.saulmaldonado.me."
			}, Timeout, Interval).Should(BeTrue())

			By("Checking mock DNS store for records with correct GameServer name and domain name")

			Eventually(func() bool {
				for _, record := range FakeDns.DnsRecords {
					if record == "_minecraft._tcp.mc-server.saulmaldonado.me. 0 0 7000 mc-node.saulmaldonado.me." {
						return true
					}
				}
				return false
			}).Should(BeTrue())
		})
	})

	Context("When deleting a GameServer", func() {
		It("Should remove DNS record", func() {

			By("Deleting GameServer")
			gs := agonesv1.GameServer{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"agones-mc/domain": "saulmaldonado.me",
					},
					Name:      GameServerName,
					Namespace: metav1.NamespaceDefault,
				},
			}

			Expect(testClient.Delete(ctx, &gs)).Should(Succeed())

			By("Checking if GameServer is deleted")

			gameServerKey := types.NamespacedName{Namespace: metav1.NamespaceDefault, Name: GameServerName}
			createdGameServer := agonesv1.GameServer{}

			Eventually(func() error {
				return testClient.Get(ctx, gameServerKey, &createdGameServer)
			}, Timeout, Interval).ShouldNot(Succeed())

			By("Checking mock DNS store for DNS records")
			Eventually(func() bool {
				for _, record := range FakeDns.DnsRecords {
					if record == "_minecraft._tcp.mc-server.saulmaldonado.me. 0 0 7000 mc-node.saulmaldonado.me." {
						return true
					}
				}
				return false
			}).Should(BeFalse())

		})
	})

})
