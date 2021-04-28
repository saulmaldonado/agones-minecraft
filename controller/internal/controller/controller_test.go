package controller

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

	Context("When creating GameServer", func() {
		It("Should add externalDNS annotation to GameServer", func() {
			By("By createing a new GameServer")

			const (
				Timeout  = time.Second * 10
				Duration = time.Second * 10
				Interval = time.Millisecond * 250
			)

			var (
				GameServerContainer string = "mc-server"
				GameServerName      string = "mc-server"
			)

			ctx := context.Background()

			gs := &agonesv1.GameServer{
				Status: agonesv1.GameServerStatus{State: agonesv1.GameServerStateScheduled},
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

			gameServerKey := types.NamespacedName{Namespace: metav1.NamespaceDefault, Name: GameServerName}
			createdGameServer := &agonesv1.GameServer{}

			Eventually(func() bool {
				if err := testClient.Get(ctx, gameServerKey, createdGameServer); err != nil {
					return false
				}
				return createdGameServer.Annotations["agones-mc/externalDNS"] == "mc-server.saulmaldonado.me."
			}, Timeout, Interval).Should(BeTrue())

		})
	})
})
