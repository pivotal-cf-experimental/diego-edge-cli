package route_helpers_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/pivotal-cf-experimental/lattice-cli/cli/route_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoutingInfoHelpers", func() {
	var (
		route1 route_helpers.AppRoute
		route2 route_helpers.AppRoute
		route3 route_helpers.AppRoute

		routes route_helpers.AppRoutes
	)

	BeforeEach(func() {
		route1 = route_helpers.AppRoute{
			Hostnames: []string{"foo1.example.com", "bar1.examaple.com"},
			Port:      11111,
		}
		route2 = route_helpers.AppRoute{
			Hostnames: []string{"foo2.example.com", "bar2.examaple.com"},
			Port:      22222,
		}
		route3 = route_helpers.AppRoute{
			Hostnames: []string{"foo3.example.com", "bar3.examaple.com"},
			Port:      33333,
		}

		routes = route_helpers.AppRoutes{route1, route2, route3}
	})

	Describe("AppRoutes", func() {
		Describe("RoutingInfo", func() {
			var routingInfo receptor.RoutingInfo

			JustBeforeEach(func() {
				routingInfo = routes.RoutingInfo()
			})

			It("wraps the serialized routes with the correct key", func() {
				expectedBytes, err := json.Marshal(routes)
				Expect(err).ToNot(HaveOccurred())

				payload, err := routingInfo[route_helpers.AppRouter].MarshalJSON()
				Expect(err).ToNot(HaveOccurred())

				Expect(payload).To(MatchJSON(expectedBytes))
			})

			Context("when AppRoutes is empty", func() {
				BeforeEach(func() {
					routes = route_helpers.AppRoutes{}
				})

				It("marshals an empty list", func() {
					payload, err := routingInfo[route_helpers.AppRouter].MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(payload).To(MatchJSON(`[]`))
				})
			})

		})
	})

	Describe("AppRoutesFromRoutingInfo", func() {
		var (
			routesResult route_helpers.AppRoutes
			routingInfo  receptor.RoutingInfo
		)

		JustBeforeEach(func() {
			routesResult = route_helpers.AppRoutesFromRoutingInfo(routingInfo)
		})

		Context("when lattice app routes are present in the routing info", func() {
			BeforeEach(func() {
				routingInfo = routes.RoutingInfo()
			})

			It("returns the routes", func() {
				Expect(routes).To(Equal(routesResult))
			})

			Context("when the lattice routes are nil", func() {
				BeforeEach(func() {
					routingInfo = receptor.RoutingInfo{route_helpers.AppRouter: nil}
				})

				It("returns nil routes", func() {
					Expect(routesResult).To(BeNil())
				})
			})
		})

		Context("when lattice app routes are not present in the routing info", func() {
			BeforeEach(func() {
				routingInfo = receptor.RoutingInfo{}
			})

			It("returns nil routes", func() {
				Expect(routesResult).To(BeNil())
			})
		})

		Context("when the routing info is nil", func() {
			BeforeEach(func() {
				routingInfo = nil
			})

			It("returns nil routes", func() {
				Expect(routesResult).To(BeNil())
			})

		})
	})

})
