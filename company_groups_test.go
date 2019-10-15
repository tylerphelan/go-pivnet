package pivnet_test

import (
	"fmt"
	"github.com/pivotal-cf/go-pivnet/v2/go-pivnetfakes"
	"net/http"

	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/v2"
	"github.com/pivotal-cf/go-pivnet/v2/logger"
	"github.com/pivotal-cf/go-pivnet/v2/logger/loggerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PivnetClient - company groups", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		apiAddress string
		userAgent  string

		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService
		response               interface{}
		responseStatusCode     int
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
		fakeAccessTokenService = &gopivnetfakes.FakeAccessTokenService{}
		newClientConfig = pivnet.ClientConfig{
			Host:      apiAddress,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(fakeAccessTokenService, newClientConfig, fakeLogger)

		responseStatusCode = http.StatusOK
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("List", func() {
		It("returns all company groups", func() {
			response := `{"company_groups": [{"id":2,"name":"company group 1"},{"id": 3, "name": "company group 2"}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/company_groups", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			companyGroups, err := client.CompanyGroups.List()
			Expect(err).NotTo(HaveOccurred())

			Expect(companyGroups).To(HaveLen(2))
			Expect(companyGroups[0].ID).To(Equal(2))
			Expect(companyGroups[1].ID).To(Equal(3))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/company_groups", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.CompanyGroups.List()
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/company_groups", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.CompanyGroups.List()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Get Company Group", func() {
		var (
			companyGroupID int
		)

		BeforeEach(func() {
			companyGroupID = 1234

			response = pivnet.CompanyGroup{
				ID:   companyGroupID,
				Name: "some company group",
				Members: []pivnet.CompanyGroupMember{
					{
						ID:      4321,
						Name:    "company group member 1",
						Email:   "dude@dude.dude",
						IsAdmin: false,
					},
					{
						ID:      9876,
						Name:    "company group member 2",
						Email:   "buddy@buddy.buddy",
						IsAdmin: true,
					},
				},
				PendingInvitations: []string{},
				Entitlements:       []pivnet.CompanyGroupEntitlement{},
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/company_groups/%d",
							apiPrefix,
							companyGroupID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns company group without errors", func() {
			_, err := client.CompanyGroups.Get(companyGroupID)

			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
				responseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf(
								"%s/company_groups/%d",
								apiPrefix,
								companyGroupID,
							),
						),
						ghttp.RespondWith(responseStatusCode, body),
					),
				)

				_, err := client.CompanyGroups.Get(
					companyGroupID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("AddMember", func() {
		var (
			companyGroupID      int
			memberEmailAddress  string
			expectedRequestBody string
		)

		BeforeEach(func() {
			companyGroupID = 1234
			memberEmailAddress = "dude@dude.dude"

			response = pivnet.CompanyGroup{
				ID:   companyGroupID,
				Name: "some company group",
				Members: []pivnet.CompanyGroupMember{
					{
						ID:      4321,
						Name:    "company group member 1",
						Email:   "dude@dude.dude",
						IsAdmin: false,
					},
					{
						ID:      9876,
						Name:    "company group member 2",
						Email:   "buddy@buddy.buddy",
						IsAdmin: true,
					},
				},
				PendingInvitations: []string{},
				Entitlements:       []pivnet.CompanyGroupEntitlement{},
			}

			expectedRequestBody = fmt.Sprintf(
				`{"member":{"email":"%s","admin":false}}`,
				memberEmailAddress,
			)
		})

		It("should return the changed company group when successful", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/company_groups/%d/add_member",
							apiPrefix,
							companyGroupID,
						),
					),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)

			_, err := client.CompanyGroups.AddMember(companyGroupID, memberEmailAddress, "false")

			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf("%s/company_groups/%d/add_member", apiPrefix, 1234)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.CompanyGroups.AddMember(1234, "dude@dude.dude", "false")
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/company_groups/%d/add_member",
							apiPrefix,
							4321,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.CompanyGroups.AddMember(4321, memberEmailAddress, "false")
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("RemoveMember", func() {
		var (
			companyGroupID      int
			memberEmailAddress  string
			expectedRequestBody string
		)

		BeforeEach(func() {
			companyGroupID = 1234
			memberEmailAddress = "dude@dude.dude"

			response = pivnet.CompanyGroup{
				ID:   companyGroupID,
				Name: "some company group",
				Members: []pivnet.CompanyGroupMember{
					{
						ID:      4321,
						Name:    "company group member 1",
						Email:   "dude@dude.dude",
						IsAdmin: false,
					},
					{
						ID:      9876,
						Name:    "company group member 2",
						Email:   "buddy@buddy.buddy",
						IsAdmin: true,
					},
				},
				PendingInvitations: []string{},
				Entitlements:       []pivnet.CompanyGroupEntitlement{},
			}

			expectedRequestBody = fmt.Sprintf(
				`{"member":{"email":"%s"}}`,
				memberEmailAddress,
			)
		})

		It("should return the changed company group when successful", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/company_groups/%d/remove_member",
							apiPrefix,
							companyGroupID,
						),
					),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)

			_, err := client.CompanyGroups.RemoveMember(companyGroupID, memberEmailAddress)

			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf("%s/company_groups/%d/remove_member", apiPrefix, 1234)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.CompanyGroups.RemoveMember(1234, "dude@dude.dude")
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/company_groups/%d/remove_member",
							apiPrefix,
							4321,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.CompanyGroups.RemoveMember(4321, memberEmailAddress)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

})
