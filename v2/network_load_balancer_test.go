package v2

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/jarcoal/httpmock"

	"github.com/exoscale/egoscale/v2/oapi"
)

var (
	testNLBID                                          = new(clientTestSuite).randomID()
	testNLBName                                        = new(clientTestSuite).randomString(10)
	testNLBDescription                                 = new(clientTestSuite).randomString(10)
	testNLBCreatedAt, _                                = time.Parse(iso8601Format, "2020-05-26T12:09:42Z")
	testNLBIPAddress                                   = "101.102.103.104"
	testNLBIPAddressP                                  = net.ParseIP("101.102.103.104")
	testNLBLabels                                      = map[string]string{"k1": "v1", "k2": "v2"}
	testNLBState                                       = oapi.LoadBalancerStateRunning
	testNLBServiceID                                   = new(clientTestSuite).randomID()
	testNLBServiceName                                 = new(clientTestSuite).randomString(10)
	testNLBServiceDescription                          = new(clientTestSuite).randomID()
	testNLBServiceInstancePoolID                       = new(clientTestSuite).randomID()
	testNLBServiceProtocol                             = oapi.LoadBalancerServiceProtocolTcp
	testNLBServicePort                          uint16 = 443
	testNLBServiceTargetPort                    uint16 = 8443
	testNLBServiceStrategy                             = oapi.LoadBalancerServiceStrategyRoundRobin
	testNLBServiceState                                = oapi.DbaasServiceStateRunning
	testNLServiceHealthcheckMode                       = oapi.LoadBalancerServiceHealthcheckModeHttps
	testNLBServiceHealthcheckPort               uint16 = 8080
	testNLBServiceHealthcheckInterval           int64  = 10
	testNLBServiceHealthcheckIntervalD                 = time.Duration(testNLBServiceHealthcheckInterval) * time.Second
	testNLBServiceHealthcheckTimeout            int64  = 3
	testNLBServiceHealthcheckTimeoutD                  = time.Duration(testNLBServiceHealthcheckTimeout) * time.Second
	testNLBServiceHealthcheckRetries            int64  = 1
	testNLBServiceHealthcheckURI                       = new(clientTestSuite).randomString(10)
	testNLBServiceHealthcheckTLSSNI                    = new(clientTestSuite).randomString(10)
	testNLBServiceHealthcheckStatus1InstanceIP         = "1.2.3.4"
	testNLBServiceHealthcheckStatus1InstanceIPP        = net.ParseIP("1.2.3.4")
	testNLBServiceHealthcheckStatus1Status             = oapi.LoadBalancerServerStatusStatusSuccess
	testNLBServiceHealthcheckStatus2InstanceIP         = "5.6.7.8"
	testNLBServiceHealthcheckStatus2InstanceIPP        = net.ParseIP("5.6.7.8")
	testNLBServiceHealthcheckStatus2Status             = oapi.LoadBalancerServerStatusStatusSuccess
)

func (ts *clientTestSuite) TestClient_CreateNetworkLoadBalancer() {
	var (
		testOperationID    = ts.randomID()
		testOperationState = oapi.OperationStateSuccess
	)

	httpmock.RegisterResponder("POST", "/load-balancer",
		func(req *http.Request) (*http.Response, error) {
			var actual oapi.CreateLoadBalancerJSONRequestBody
			ts.unmarshalJSONRequestBody(req, &actual)

			expected := oapi.CreateLoadBalancerJSONRequestBody{
				Description: &testNLBDescription,
				Labels:      &oapi.Labels{AdditionalProperties: testNLBLabels},
				Name:        testNLBName,
			}
			ts.Require().Equal(expected, actual)

			resp, err := httpmock.NewJsonResponse(http.StatusOK, oapi.Operation{
				Id:        &testOperationID,
				State:     &testOperationState,
				Reference: &oapi.Reference{Id: &testNLBID},
			})
			if err != nil {
				ts.T().Fatalf("error initializing mock HTTP responder: %s", err)
			}

			return resp, nil
		})

	ts.mockAPIRequest("GET", fmt.Sprintf("/operation/%s", testOperationID), oapi.Operation{
		Id:        &testOperationID,
		State:     &testOperationState,
		Reference: &oapi.Reference{Id: &testNLBID},
	})

	ts.mockAPIRequest("GET", fmt.Sprintf("/load-balancer/%s", testNLBID), oapi.LoadBalancer{
		CreatedAt:   &testNLBCreatedAt,
		Description: &testNLBDescription,
		Id:          &testNLBID,
		Labels:      &oapi.Labels{AdditionalProperties: testNLBLabels},
		Name:        &testNLBName,
		State:       &testNLBState,
	})

	expected := &NetworkLoadBalancer{
		CreatedAt:   &testNLBCreatedAt,
		Description: &testNLBDescription,
		ID:          &testNLBID,
		Labels:      &testNLBLabels,
		Name:        &testNLBName,
		Services:    []*NetworkLoadBalancerService{},
		State:       (*string)(&testNLBState),
	}

	actual, err := ts.client.CreateNetworkLoadBalancer(context.Background(), testZone, &NetworkLoadBalancer{
		Description: &testNLBDescription,
		Labels:      &testNLBLabels,
		Name:        &testNLBName,
	})
	ts.Require().NoError(err)
	ts.Require().Equal(expected, actual)
}

func (ts *clientTestSuite) TestClient_CreateNetworkLoadBalancerService() {
	var (
		testOperationID    = ts.randomID()
		testOperationState = oapi.OperationStateSuccess
	)

	httpmock.RegisterResponder("POST", fmt.Sprintf("/load-balancer/%s/service", testNLBID),
		func(req *http.Request) (*http.Response, error) {
			var actual oapi.AddServiceToLoadBalancerJSONRequestBody
			ts.unmarshalJSONRequestBody(req, &actual)

			expected := oapi.AddServiceToLoadBalancerJSONRequestBody{
				Description: &testNLBServiceDescription,
				Healthcheck: oapi.LoadBalancerServiceHealthcheck{
					Interval: &testNLBServiceHealthcheckInterval,
					Mode:     &testNLServiceHealthcheckMode,
					Port:     func() *int64 { v := int64(testNLBServiceHealthcheckPort); return &v }(),
					Retries:  &testNLBServiceHealthcheckRetries,
					Timeout:  &testNLBServiceHealthcheckTimeout,
					TlsSni:   &testNLBServiceHealthcheckTLSSNI,
					Uri:      &testNLBServiceHealthcheckURI,
				},
				InstancePool: oapi.InstancePool{Id: &testNLBServiceInstancePoolID},
				Name:         testNLBServiceName,
				Port:         int64(testNLBServicePort),
				Protocol:     oapi.AddServiceToLoadBalancerJSONBodyProtocol(testNLBServiceProtocol),
				Strategy:     oapi.AddServiceToLoadBalancerJSONBodyStrategy(testNLBServiceStrategy),
				TargetPort:   int64(testNLBServiceTargetPort),
			}
			ts.Require().Equal(expected, actual)

			resp, err := httpmock.NewJsonResponse(http.StatusOK, oapi.Operation{
				Id:        &testOperationID,
				State:     &testOperationState,
				Reference: &oapi.Reference{Id: &testNLBID},
			})
			if err != nil {
				ts.T().Fatalf("error initializing mock HTTP responder: %s", err)
			}

			return resp, nil
		})

	ts.mockAPIRequest("GET", fmt.Sprintf("/operation/%s", testOperationID), oapi.Operation{
		Id:        &testOperationID,
		State:     &testOperationState,
		Reference: &oapi.Reference{Id: &testNLBID},
	})

	ts.mockAPIRequest("GET", fmt.Sprintf("/load-balancer/%s", testNLBID), oapi.LoadBalancer{
		CreatedAt:   &testNLBCreatedAt,
		Description: &testNLBDescription,
		Id:          &testNLBID,
		Ip:          &testNLBIPAddress,
		Name:        &testNLBName,
		Services: &[]oapi.LoadBalancerService{{
			Description: &testNLBServiceDescription,
			Healthcheck: &oapi.LoadBalancerServiceHealthcheck{
				Interval: &testNLBServiceHealthcheckInterval,
				Mode:     &testNLServiceHealthcheckMode,
				Port:     func() *int64 { v := int64(testNLBServiceHealthcheckPort); return &v }(),
				Retries:  &testNLBServiceHealthcheckRetries,
				Timeout:  &testNLBServiceHealthcheckTimeout,
				TlsSni:   &testNLBServiceHealthcheckTLSSNI,
				Uri:      &testNLBServiceHealthcheckURI,
			},
			HealthcheckStatus: &[]oapi.LoadBalancerServerStatus{
				{
					PublicIp: &testNLBServiceHealthcheckStatus1InstanceIP,
					Status:   &testNLBServiceHealthcheckStatus1Status,
				},
				{
					PublicIp: &testNLBServiceHealthcheckStatus2InstanceIP,
					Status:   &testNLBServiceHealthcheckStatus2Status,
				},
			},
			Id:           &testNLBServiceID,
			InstancePool: &oapi.InstancePool{Id: &testNLBServiceInstancePoolID},
			Name:         &testNLBServiceName,
			Port:         func() *int64 { v := int64(testNLBServicePort); return &v }(),
			Protocol:     &testNLBServiceProtocol,
			Strategy:     &testNLBServiceStrategy,
			TargetPort:   func() *int64 { v := int64(testNLBServiceTargetPort); return &v }(),
			State:        (*oapi.LoadBalancerServiceState)(&testNLBServiceState),
		}},
		State: &testNLBState,
	})

	nlb := &NetworkLoadBalancer{
		CreatedAt:   &testNLBCreatedAt,
		Description: &testNLBDescription,
		ID:          &testNLBID,
		IPAddress:   &testNLBIPAddressP,
		Name:        &testNLBName,
		State:       (*string)(&testNLBState),
	}

	expected := &NetworkLoadBalancerService{
		Description: &testNLBServiceDescription,
		Healthcheck: &NetworkLoadBalancerServiceHealthcheck{
			Interval: &testNLBServiceHealthcheckIntervalD,
			Mode:     (*string)(&testNLServiceHealthcheckMode),
			Port:     &testNLBServiceHealthcheckPort,
			Retries:  &testNLBServiceHealthcheckRetries,
			TLSSNI:   &testNLBServiceHealthcheckTLSSNI,
			Timeout:  &testNLBServiceHealthcheckTimeoutD,
			URI:      &testNLBServiceHealthcheckURI,
		},
		HealthcheckStatus: []*NetworkLoadBalancerServerStatus{
			{
				InstanceIP: &testNLBServiceHealthcheckStatus1InstanceIPP,
				Status:     (*string)(&testNLBServiceHealthcheckStatus1Status),
			},
			{
				InstanceIP: &testNLBServiceHealthcheckStatus2InstanceIPP,
				Status:     (*string)(&testNLBServiceHealthcheckStatus2Status),
			},
		},
		ID:             &testNLBServiceID,
		InstancePoolID: &testNLBServiceInstancePoolID,
		Name:           &testNLBServiceName,
		Port:           &testNLBServicePort,
		Protocol:       (*string)(&testNLBServiceProtocol),
		Strategy:       (*string)(&testNLBServiceStrategy),
		TargetPort:     &testNLBServiceTargetPort,
		State:          (*string)(&testNLBServiceState),
	}

	actual, err := ts.client.CreateNetworkLoadBalancerService(context.Background(), testZone, nlb, expected)
	ts.Require().NoError(err)
	ts.Require().Equal(expected, actual)
}

func (ts *clientTestSuite) TestClient_DeleteNetworkLoadBalancer() {
	var (
		testOperationID    = ts.randomID()
		testOperationState = oapi.OperationStateSuccess
		deleted            = false
	)

	httpmock.RegisterResponder("DELETE", fmt.Sprintf("/load-balancer/%s", testNLBID),
		func(req *http.Request) (*http.Response, error) {
			deleted = true

			resp, err := httpmock.NewJsonResponse(http.StatusOK, oapi.Operation{
				Id:        &testOperationID,
				State:     &testOperationState,
				Reference: &oapi.Reference{Id: &testNLBID},
			})
			if err != nil {
				ts.T().Fatalf("error initializing mock HTTP responder: %s", err)
			}

			return resp, nil
		})

	ts.mockAPIRequest("GET", fmt.Sprintf("/operation/%s", testOperationID), oapi.Operation{
		Id:        &testOperationID,
		State:     &testOperationState,
		Reference: &oapi.Reference{Id: &testNLBID},
	})

	ts.Require().NoError(ts.client.DeleteNetworkLoadBalancer(
		context.Background(),
		testZone,
		&NetworkLoadBalancer{ID: &testNLBID},
	))
	ts.Require().True(deleted)
}

func (ts *clientTestSuite) TestClient_DeleteNetworkLoadBalancerService() {
	var (
		testOperationID    = ts.randomID()
		testOperationState = oapi.OperationStateSuccess
		deleted            = false
	)

	httpmock.RegisterResponder("DELETE",
		fmt.Sprintf("/load-balancer/%s/service/%s", testNLBID, testNLBServiceID),
		func(req *http.Request) (*http.Response, error) {
			deleted = true

			resp, err := httpmock.NewJsonResponse(http.StatusOK, oapi.Operation{
				Id:        &testOperationID,
				State:     &testOperationState,
				Reference: &oapi.Reference{Id: &testNLBServiceID},
			})
			if err != nil {
				ts.T().Fatalf("error initializing mock HTTP responder: %s", err)
			}

			return resp, nil
		})

	ts.mockAPIRequest("GET", fmt.Sprintf("/operation/%s", testOperationID), oapi.Operation{
		Id:        &testOperationID,
		State:     &testOperationState,
		Reference: &oapi.Reference{Id: &testNLBServiceID},
	})

	nlb := &NetworkLoadBalancer{
		ID:       &testNLBID,
		Services: []*NetworkLoadBalancerService{{ID: &testNLBServiceID}},
	}

	ts.Require().NoError(ts.client.DeleteNetworkLoadBalancerService(
		context.Background(),
		testZone,
		nlb,
		nlb.Services[0],
	))
	ts.Require().True(deleted)
}

func (ts *clientTestSuite) TestClient_FindNetworkLoadBalancer() {
	ts.mockAPIRequest("GET", "/load-balancer", struct {
		LoadBalancers *[]oapi.LoadBalancer `json:"load-balancers,omitempty"`
	}{
		LoadBalancers: &[]oapi.LoadBalancer{{
			CreatedAt: &testNLBCreatedAt,
			Id:        &testNLBID,
			Name:      &testNLBName,
			State:     &testNLBState,
		}},
	})
	ts.mockAPIRequest("GET", fmt.Sprintf("/load-balancer/%s", testNLBID), oapi.LoadBalancer{
		CreatedAt: &testNLBCreatedAt,
		Id:        &testNLBID,
		Ip:        &testNLBIPAddress,
		Name:      &testNLBName,
		State:     &testNLBState,
	})

	expected := &NetworkLoadBalancer{
		CreatedAt: &testNLBCreatedAt,
		ID:        &testNLBID,
		IPAddress: &testNLBIPAddressP,
		Name:      &testNLBName,
		Services:  []*NetworkLoadBalancerService{},
		State:     (*string)(&testNLBState),
	}

	actual, err := ts.client.FindNetworkLoadBalancer(context.Background(), testZone, *expected.ID)
	ts.Require().NoError(err)
	ts.Require().Equal(expected, actual)

	actual, err = ts.client.FindNetworkLoadBalancer(context.Background(), testZone, *expected.Name)
	ts.Require().NoError(err)
	ts.Require().Equal(expected, actual)
}

func (ts *clientTestSuite) TestClient_GetNetworkLoadBalancer() {
	ts.mockAPIRequest("GET", fmt.Sprintf("/load-balancer/%s", testNLBID), oapi.LoadBalancer{
		CreatedAt:   &testNLBCreatedAt,
		Description: &testNLBDescription,
		Id:          &testNLBID,
		Ip:          &testNLBIPAddress,
		Name:        &testNLBName,
		Services: &[]oapi.LoadBalancerService{{
			Description: &testNLBServiceDescription,
			Healthcheck: &oapi.LoadBalancerServiceHealthcheck{
				Interval: &testNLBServiceHealthcheckInterval,
				Mode:     &testNLServiceHealthcheckMode,
				Port:     func() *int64 { v := int64(testNLBServiceHealthcheckPort); return &v }(),
				Retries:  &testNLBServiceHealthcheckRetries,
				Timeout:  &testNLBServiceHealthcheckTimeout,
				Uri:      &testNLBServiceHealthcheckURI,
			},
			HealthcheckStatus: &[]oapi.LoadBalancerServerStatus{
				{
					PublicIp: &testNLBServiceHealthcheckStatus1InstanceIP,
					Status:   &testNLBServiceHealthcheckStatus1Status,
				},
				{
					PublicIp: &testNLBServiceHealthcheckStatus2InstanceIP,
					Status:   &testNLBServiceHealthcheckStatus2Status,
				},
			},
			Id:           &testNLBServiceID,
			InstancePool: &oapi.InstancePool{Id: &testNLBServiceInstancePoolID},
			Name:         &testNLBServiceName,
			Port:         func() *int64 { v := int64(testNLBServicePort); return &v }(),
			Protocol:     &testNLBServiceProtocol,
			State:        (*oapi.LoadBalancerServiceState)(&testNLBServiceState),
			Strategy:     &testNLBServiceStrategy,
			TargetPort:   func() *int64 { v := int64(testNLBServiceTargetPort); return &v }(),
		}},
		State: &testNLBState,
	})

	expected := &NetworkLoadBalancer{
		CreatedAt:   &testNLBCreatedAt,
		Description: &testNLBDescription,
		ID:          &testNLBID,
		IPAddress:   &testNLBIPAddressP,
		Name:        &testNLBName,
		State:       (*string)(&testNLBState),
		Services: []*NetworkLoadBalancerService{{
			Description: &testNLBServiceDescription,
			Healthcheck: &NetworkLoadBalancerServiceHealthcheck{
				Interval: &testNLBServiceHealthcheckIntervalD,
				Mode:     (*string)(&testNLServiceHealthcheckMode),
				Port:     &testNLBServiceHealthcheckPort,
				Retries:  &testNLBServiceHealthcheckRetries,
				Timeout:  &testNLBServiceHealthcheckTimeoutD,
				URI:      &testNLBServiceHealthcheckURI,
			},
			HealthcheckStatus: []*NetworkLoadBalancerServerStatus{
				{
					InstanceIP: &testNLBServiceHealthcheckStatus1InstanceIPP,
					Status:     (*string)(&testNLBServiceHealthcheckStatus1Status),
				},
				{
					InstanceIP: &testNLBServiceHealthcheckStatus2InstanceIPP,
					Status:     (*string)(&testNLBServiceHealthcheckStatus2Status),
				},
			},
			ID:             &testNLBServiceID,
			InstancePoolID: &testNLBServiceInstancePoolID,
			Name:           &testNLBServiceName,
			Port:           &testNLBServicePort,
			Protocol:       (*string)(&testNLBServiceProtocol),
			State:          (*string)(&testNLBServiceState),
			Strategy:       (*string)(&testNLBServiceStrategy),
			TargetPort:     &testNLBServiceTargetPort,
		}},
	}

	actual, err := ts.client.GetNetworkLoadBalancer(context.Background(), testZone, *expected.ID)
	ts.Require().NoError(err)
	ts.Require().Equal(expected, actual)
}

func (ts *clientTestSuite) TestClient_ListNetworkLoadBalancers() {
	ts.mockAPIRequest("GET", "/load-balancer", struct {
		LoadBalancers *[]oapi.LoadBalancer `json:"load-balancers,omitempty"`
	}{
		LoadBalancers: &[]oapi.LoadBalancer{{
			CreatedAt:   &testNLBCreatedAt,
			Description: &testNLBDescription,
			Id:          &testNLBID,
			Name:        &testNLBName,
			Services: &[]oapi.LoadBalancerService{{
				Description: &testNLBServiceDescription,
				Healthcheck: &oapi.LoadBalancerServiceHealthcheck{
					Interval: &testNLBServiceHealthcheckInterval,
					Mode:     &testNLServiceHealthcheckMode,
					Port:     func() *int64 { v := int64(testNLBServiceHealthcheckPort); return &v }(),
					Retries:  &testNLBServiceHealthcheckRetries,
					Timeout:  &testNLBServiceHealthcheckTimeout,
					Uri:      &testNLBServiceHealthcheckURI,
				},
				HealthcheckStatus: &[]oapi.LoadBalancerServerStatus{
					{
						PublicIp: &testNLBServiceHealthcheckStatus1InstanceIP,
						Status:   &testNLBServiceHealthcheckStatus1Status,
					},
					{
						PublicIp: &testNLBServiceHealthcheckStatus2InstanceIP,
						Status:   &testNLBServiceHealthcheckStatus2Status,
					},
				},
				Id:           &testNLBServiceID,
				InstancePool: &oapi.InstancePool{Id: &testNLBServiceInstancePoolID},
				Name:         &testNLBServiceName,
				Port:         func() *int64 { v := int64(testNLBServicePort); return &v }(),
				Protocol:     &testNLBServiceProtocol,
				State:        (*oapi.LoadBalancerServiceState)(&testNLBState),
				Strategy:     &testNLBServiceStrategy,
				TargetPort:   func() *int64 { v := int64(testNLBServiceTargetPort); return &v }(),
			}},
			State: &testNLBState,
		}},
	})

	expected := []*NetworkLoadBalancer{{
		CreatedAt:   &testNLBCreatedAt,
		Description: &testNLBDescription,
		ID:          &testNLBID,
		Name:        &testNLBName,
		Services: []*NetworkLoadBalancerService{{
			Description: &testNLBServiceDescription,
			Healthcheck: &NetworkLoadBalancerServiceHealthcheck{
				Interval: &testNLBServiceHealthcheckIntervalD,
				Mode:     (*string)(&testNLServiceHealthcheckMode),
				Port:     &testNLBServiceHealthcheckPort,
				Retries:  &testNLBServiceHealthcheckRetries,
				Timeout:  &testNLBServiceHealthcheckTimeoutD,
				URI:      &testNLBServiceHealthcheckURI,
			},
			HealthcheckStatus: []*NetworkLoadBalancerServerStatus{
				{
					InstanceIP: &testNLBServiceHealthcheckStatus1InstanceIPP,
					Status:     (*string)(&testNLBServiceHealthcheckStatus1Status),
				},
				{
					InstanceIP: &testNLBServiceHealthcheckStatus2InstanceIPP,
					Status:     (*string)(&testNLBServiceHealthcheckStatus2Status),
				},
			},
			ID:             &testNLBServiceID,
			InstancePoolID: &testNLBServiceInstancePoolID,
			Name:           &testNLBServiceName,
			Port:           &testNLBServicePort,
			Protocol:       (*string)(&testNLBServiceProtocol),
			State:          (*string)(&testNLBState),
			Strategy:       (*string)(&testNLBServiceStrategy),
			TargetPort:     &testNLBServiceTargetPort,
		}},
		State: (*string)(&testNLBState),
	}}

	actual, err := ts.client.ListNetworkLoadBalancers(context.Background(), testZone)
	ts.Require().NoError(err)
	ts.Require().Equal(expected, actual)
}

func (ts *clientTestSuite) TestClient_UpdateNetworkLoadBalancer() {
	var (
		testNLBDescriptionUpdated = testNLBDescription + "-updated"
		testNLBLabelsUpdated      = map[string]string{"k3": "v3"}
		testNLBNameUpdated        = testNLBName + "-updated"
		testOperationID           = ts.randomID()
		testOperationState        = oapi.OperationStateSuccess
		updated                   = false
	)

	httpmock.RegisterResponder("PUT", fmt.Sprintf("/load-balancer/%s", testNLBID),
		func(req *http.Request) (*http.Response, error) {
			updated = true

			var actual oapi.UpdateLoadBalancerJSONRequestBody
			ts.unmarshalJSONRequestBody(req, &actual)

			expected := oapi.UpdateLoadBalancerJSONRequestBody{
				Description: &testNLBDescriptionUpdated,
				Labels:      &oapi.Labels{AdditionalProperties: testNLBLabelsUpdated},
				Name:        &testNLBNameUpdated,
			}
			ts.Require().Equal(expected, actual)

			resp, err := httpmock.NewJsonResponse(http.StatusOK, oapi.Operation{
				Id:        &testOperationID,
				State:     &testOperationState,
				Reference: &oapi.Reference{Id: &testNLBID},
			})
			if err != nil {
				ts.T().Fatalf("error initializing mock HTTP responder: %s", err)
			}

			return resp, nil
		})

	ts.mockAPIRequest("GET", fmt.Sprintf("/operation/%s", testOperationID), oapi.Operation{
		Id:        &testOperationID,
		State:     &testOperationState,
		Reference: &oapi.Reference{Id: &testNLBID},
	})

	ts.mockAPIRequest("GET", fmt.Sprintf("/load-balancer/%s", testNLBID), oapi.LoadBalancer{
		CreatedAt:   &testNLBCreatedAt,
		Description: &testNLBDescriptionUpdated,
		Id:          &testNLBID,
		Ip:          &testNLBIPAddress,
		Labels:      &oapi.Labels{AdditionalProperties: testNLBLabelsUpdated},
		Name:        &testNLBNameUpdated,
	})

	nlbUpdated := NetworkLoadBalancer{
		Description: &testNLBDescriptionUpdated,
		ID:          &testNLBID,
		Labels:      &testNLBLabelsUpdated,
		Name:        &testNLBNameUpdated,
	}

	ts.Require().NoError(ts.client.UpdateNetworkLoadBalancer(context.Background(), testZone, &nlbUpdated))
	ts.Require().True(updated)
}

func (ts *clientTestSuite) TestClient_UpdateNetworkLoadBalancerService() {
	var (
		testNLBServiceNameUpdated                 = testNLBServiceName + "-updated"
		testNLBServiceDescriptionUpdated          = testNLBServiceDescription + "-updated"
		testNLBServiceHealthcheckModeUpdated      = oapi.LoadBalancerServiceHealthcheckModeHttp
		testNLBServiceHealthcheckPortUpdated      = testNLBServiceHealthcheckPort + 1
		testNLBServiceHealthcheckRetriesUpdated   = testNLBServiceHealthcheckRetries + 1
		testNLBServiceHealthcheckTLSSNIUpdated    = ""
		testNLBServiceHealthcheckIntervalUpdated  = testNLBServiceHealthcheckInterval + 1
		testNLBServiceHealthcheckIntervalDUpdated = time.Duration(testNLBServiceHealthcheckIntervalUpdated) * time.Second
		testNLBServiceHealthcheckTimeoutUpdated   = testElasticIPHealthcheckTimeout + 1
		testNLBServiceHealthcheckTimeoutDUpdated  = time.Duration(testNLBServiceHealthcheckTimeoutUpdated) * time.Second
		testNLBServiceHealthcheckURIUpdated       = ""
		testOperationID                           = ts.randomID()
		testOperationState                        = oapi.OperationStateSuccess
		updated                                   = false
	)

	nlb := &NetworkLoadBalancer{
		ID: &testNLBID,
		Services: []*NetworkLoadBalancerService{{
			ID:          &testNLBServiceID,
			Name:        &testNLBServiceName,
			Description: &testNLBServiceDescription,
		}},
	}

	httpmock.RegisterResponder("PUT", fmt.Sprintf("/load-balancer/%s/service/%s",
		*nlb.ID,
		*nlb.Services[0].ID),
		func(req *http.Request) (*http.Response, error) {
			updated = true

			var actual oapi.UpdateLoadBalancerServiceJSONRequestBody
			ts.unmarshalJSONRequestBody(req, &actual)

			expected := oapi.UpdateLoadBalancerServiceJSONRequestBody{
				Name:        &testNLBServiceNameUpdated,
				Description: &testNLBServiceDescriptionUpdated,
				Healthcheck: &oapi.LoadBalancerServiceHealthcheck{
					Interval: &testNLBServiceHealthcheckIntervalUpdated,
					Mode:     &testNLBServiceHealthcheckModeUpdated,
					Port:     func() *int64 { v := int64(testNLBServiceHealthcheckPortUpdated); return &v }(),
					Retries:  &testNLBServiceHealthcheckRetriesUpdated,
					Timeout:  &testNLBServiceHealthcheckTimeoutUpdated,
					TlsSni:   &testNLBServiceHealthcheckTLSSNIUpdated,
					Uri:      &testNLBServiceHealthcheckURIUpdated,
				},
			}
			ts.Require().Equal(expected, actual)

			resp, err := httpmock.NewJsonResponse(http.StatusOK, oapi.Operation{
				Id:        &testOperationID,
				State:     &testOperationState,
				Reference: &oapi.Reference{Id: &testNLBServiceID},
			})
			if err != nil {
				ts.T().Fatalf("error initializing mock HTTP responder: %s", err)
			}

			return resp, nil
		})

	ts.mockAPIRequest("GET", fmt.Sprintf("/operation/%s", testOperationID), oapi.Operation{
		Id:        &testOperationID,
		State:     &testOperationState,
		Reference: &oapi.Reference{Id: &testNLBServiceID},
	})

	ts.Require().NoError(ts.client.UpdateNetworkLoadBalancerService(context.Background(), testZone, nlb,
		&NetworkLoadBalancerService{
			ID:          nlb.Services[0].ID,
			Name:        &testNLBServiceNameUpdated,
			Description: &testNLBServiceDescriptionUpdated,
			Healthcheck: &NetworkLoadBalancerServiceHealthcheck{
				Interval: &testNLBServiceHealthcheckIntervalDUpdated,
				Mode:     (*string)(&testNLBServiceHealthcheckModeUpdated),
				Port:     &testNLBServiceHealthcheckPortUpdated,
				Retries:  &testNLBServiceHealthcheckRetriesUpdated,
				TLSSNI:   &testNLBServiceHealthcheckTLSSNIUpdated,
				Timeout:  &testNLBServiceHealthcheckTimeoutDUpdated,
				URI:      &testNLBServiceHealthcheckURIUpdated,
			},
		}))
	ts.Require().True(updated)
}
