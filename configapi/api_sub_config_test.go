// SPDX-License-Identifier: Apache-2.0
// Copyright 2025 Canonical Ltd.

package configapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omec-project/openapi/models"
	"github.com/omec-project/webconsole/configmodels"
	"github.com/omec-project/webconsole/dbadapter"
	"go.mongodb.org/mongo-driver/bson"
)

type MockMongoClientOneSubscriber struct {
	dbadapter.DBInterface
}

type MockMongoClientManySubscribers struct {
	dbadapter.DBInterface
}

type MockMongoClientDeviceGroupsWithSubscriber struct {
	dbadapter.DBInterface
}

func (m *MockMongoClientOneSubscriber) RestfulAPIGetMany(coll string, filter bson.M) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	subscriber := configmodels.ToBsonM(models.AccessAndMobilitySubscriptionData{})
	subscriber["ueId"] = "208930100007487"
	subscriber["servingPlmnId"] = "12345"
	var subscriberBson bson.M
	tmp, _ := json.Marshal(subscriber)
	json.Unmarshal(tmp, &subscriberBson)

	results = append(results, subscriberBson)
	return results, nil
}

func (m *MockMongoClientManySubscribers) RestfulAPIGetMany(coll string, filter bson.M) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	ueIds := []string{"208930100007487", "208930100007488"}
	plmnIDs := []string{"12345", "54321"}
	for i, ueId := range ueIds {
		subscriber := configmodels.ToBsonM(models.AccessAndMobilitySubscriptionData{})
		subscriber["ueId"] = ueId
		subscriber["servingPlmnId"] = plmnIDs[i]
		var subscriberBson bson.M
		tmp, _ := json.Marshal(subscriber)
		json.Unmarshal(tmp, &subscriberBson)

		results = append(results, subscriberBson)
	}
	return results, nil
}

func (m *MockMongoClientDeviceGroupsWithSubscriber) RestfulAPIGetMany(coll string, filter bson.M) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	dg := deviceGroupWithImsis("group1", []string{"208930100007487", "208930100007488"})
	var dgbson bson.M
	tmp, _ := json.Marshal(dg)
	json.Unmarshal(tmp, &dgbson)

	results = append(results, dgbson)
	return results, nil
}

type MockAuthDBClientEmpty struct {
	dbadapter.DBInterface
	PostData *[]map[string]interface{}
}

func (m *MockAuthDBClientEmpty) RestfulAPIGetOne(coll string, filter bson.M) (map[string]interface{}, error) {
	if m.PostData != nil {
		*m.PostData = append(*m.PostData, map[string]interface{}{
			"coll":   coll,
			"filter": filter,
		})
	}
	return nil, nil
}

type MockAuthDBClientWithData struct {
	dbadapter.DBInterface
	PostData *[]map[string]interface{}
}

func (m *MockAuthDBClientWithData) RestfulAPIGetOne(coll string, filter bson.M) (map[string]interface{}, error) {
	if m.PostData != nil {
		*m.PostData = append(*m.PostData, map[string]interface{}{
			"coll":   coll,
			"filter": filter,
		})
	}
	if coll == "subscriptionData.authenticationData.authenticationSubscription" && filter["ueId"] != nil {
		return map[string]interface{}{
			"AuthenticationManagementField": "8000",
			"AuthenticationMethod":          "5G_AKA",
			"Milenage": map[string]interface{}{
				"Op": map[string]interface{}{
					"EncryptionAlgorithm": 0,
					"EncryptionKey":       0,
					"OpValue":             "",
				},
			},
			"Opc": map[string]interface{}{
				"EncryptionAlgorithm": 0,
				"EncryptionKey":       0,
				"OpcValue":            "8e27b6af0e692e750f32667a3b14605d",
			},
			"PermanentKey": map[string]interface{}{
				"EncryptionAlgorithm": 0,
				"EncryptionKey":       0,
				"Value":               "8baf473f2f8fd09487cccbd7097c6862",
			},
			"SequenceNumber": "16f3b3f70fc2",
		}, nil
	}
	return nil, fmt.Errorf("collection %s not found", coll)
}

type MockCommonDBClientEmpty struct {
	dbadapter.DBInterface
	PostData *[]map[string]interface{}
}

func (m *MockCommonDBClientEmpty) RestfulAPIGetOne(coll string, filter bson.M) (map[string]interface{}, error) {
	if m.PostData != nil {
		*m.PostData = append(*m.PostData, map[string]interface{}{
			"coll":   coll,
			"filter": filter,
		})
	}
	return nil, nil

}

func (m *MockCommonDBClientEmpty) RestfulAPIGetMany(coll string, filter bson.M) ([]map[string]interface{}, error) {
	if m.PostData != nil {
		*m.PostData = append(*m.PostData, map[string]interface{}{
			"coll":   coll,
			"filter": filter,
		})
	}
	return nil, nil
}

type MockCommonDBClientWithData struct {
	dbadapter.DBInterface
	PostData *[]map[string]interface{}
}

func (m *MockCommonDBClientWithData) RestfulAPIGetOne(coll string, filter bson.M) (map[string]interface{}, error) {
	if m.PostData != nil {
		*m.PostData = append(*m.PostData, map[string]interface{}{
			"coll":   coll,
			"filter": filter,
		})
	}
	ueId, ok := filter["ueId"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing ueId in filter")
	}

	switch coll {
	case "subscriptionData.provisionedData.amData":
		imsi := ueId
		mcc, mnc := "001", "100"

		return map[string]interface{}{
			"ueId":           "imsi-" + imsi,
			"servingPlmnId":  mcc + mnc,
			"data":           "access management data",
			"filterCriteria": filter,
		}, nil

	case "policyData.ues.amData":
		return map[string]interface{}{
			"ueId":   ueId,
			"amData": "access management data",
		}, nil

	case "policyData.ues.smData":
		return map[string]interface{}{
			"ueId":   ueId,
			"smData": "session policy data",
		}, nil

	case "subscriptionData.provisionedData.smfSelectionSubscriptionData":
		return map[string]interface{}{
			"ueId":          ueId,
			"servingPlmnId": "001100",
			"SubscribedSnssaiInfos": map[string]interface{}{
				"010203": map[string]interface{}{
					"DnnInfos": []map[string]string{
						{"Dnn": "internet"},
					},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("collection %s not found", coll)
	}
}

func (m *MockCommonDBClientWithData) RestfulAPIGetMany(coll string, filter bson.M) ([]map[string]interface{}, error) {
	if m.PostData != nil {
		*m.PostData = append(*m.PostData, map[string]interface{}{
			"coll":   coll,
			"filter": filter,
		})
	}
	if coll == "subscriptionData.provisionedData.smData" && filter["ueId"] != nil {
		return []map[string]interface{}{
			{
				"ueId": filter["ueId"],
				"data": "session management data",
			},
		}, nil
	}
	return nil, fmt.Errorf("collection %s not found", coll)
}

func comparePostData(expected, actual []map[string]interface{}) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("length mismatch: expected %d elements, got %d elements", len(expected), len(actual))
	}

	for i := range expected {
		if !compareMaps(expected[i], actual[i]) {
			return fmt.Errorf("mismatch at index %d: expected %+v, got %+v", i, expected[i], actual[i])
		}
	}

	return nil
}

func compareMaps(map1, map2 map[string]interface{}) bool {
	if len(map1) != len(map2) {
		return false
	}
	for key, val1 := range map1 {
		val2, exists := map2[key]
		if !exists {
			return false
		}
		if fmt.Sprintf("%v", val1) != fmt.Sprintf("%v", val2) {
			return false
		}
	}
	return true
}

func TestGetSubscriberByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	AddApiService(router)
	postData := make([]map[string]interface{}, 0)

	tests := []struct {
		name                    string
		ueId                    string
		route                   string
		commonDbAdapter         dbadapter.DBInterface
		authDbAdapter           dbadapter.DBInterface
		expectedHTTPStatus      int
		expectedCollections     []string
		expectedFullResponse    string
		expectedPostDataDetails []map[string]interface{}
	}{
		{
			name:                 "No subscriber data found",
			ueId:                 "12345",
			route:                "/api/subscriber/:ueId",
			commonDbAdapter:      &MockCommonDBClientEmpty{PostData: &postData},
			authDbAdapter:        &MockAuthDBClientEmpty{PostData: &postData},
			expectedHTTPStatus:   http.StatusNotFound,
			expectedFullResponse: `{"error":"subscriber with ID 12345 not found"}`,
			expectedPostDataDetails: []map[string]interface{}{
				{"coll": "subscriptionData.authenticationData.authenticationSubscription", "filter": map[string]interface{}{"ueId": "12345"}},
				{"coll": "subscriptionData.provisionedData.amData", "filter": map[string]interface{}{"ueId": "12345"}},
				{"coll": "subscriptionData.provisionedData.smData", "filter": map[string]interface{}{"ueId": "12345"}},
				{"coll": "subscriptionData.provisionedData.smfSelectionSubscriptionData", "filter": map[string]interface{}{"ueId": "12345"}},
				{"coll": "policyData.ues.amData", "filter": map[string]interface{}{"ueId": "12345"}},
				{"coll": "policyData.ues.smData", "filter": map[string]interface{}{"ueId": "12345"}},
			},
		},

		{
			name:                 "Valid subscriber data retrieved",
			ueId:                 "12345",
			commonDbAdapter:      &MockCommonDBClientWithData{PostData: &postData},
			authDbAdapter:        &MockAuthDBClientWithData{PostData: &postData},
			route:                "/api/subscriber/:ueId",
			expectedHTTPStatus:   http.StatusOK,
			expectedFullResponse: `{"plmnID":"","ueId":"12345","AuthenticationSubscription":{"authenticationMethod":"5G_AKA","permanentKey":{"permanentKeyValue":"","encryptionKey":0,"encryptionAlgorithm":0},"sequenceNumber":"16f3b3f70fc2","authenticationManagementField":"8000","milenage":{"op":{"opValue":"","encryptionKey":0,"encryptionAlgorithm":0}},"opc":{"opcValue":"8e27b6af0e692e750f32667a3b14605d","encryptionKey":0,"encryptionAlgorithm":0}},"AccessAndMobilitySubscriptionData":{},"SessionManagementSubscriptionData":[{"singleNssai":null}],"SmfSelectionSubscriptionData":{"subscribedSnssaiInfos":{"010203":{"dnnInfos":[{"dnn":"internet"}]}}},"AmPolicyData":{},"SmPolicyData":{"smPolicySnssaiData":null},"FlowRules":null}`,
			expectedPostDataDetails: []map[string]interface{}{
				{"coll": "subscriptionData.authenticationData.authenticationSubscription", "filter": map[string]interface{}{"ueId": "12345"}},
				{"coll": "subscriptionData.provisionedData.amData", "filter": map[string]interface{}{"ueId": "12345"}},
				{"coll": "subscriptionData.provisionedData.smData", "filter": map[string]interface{}{"ueId": "12345"}},
				{"coll": "subscriptionData.provisionedData.smfSelectionSubscriptionData", "filter": map[string]interface{}{"ueId": "12345"}},
				{"coll": "policyData.ues.amData", "filter": map[string]interface{}{"ueId": "12345"}},
				{"coll": "policyData.ues.smData", "filter": map[string]interface{}{"ueId": "12345"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postData = nil
			originalAuthDBClient := dbadapter.AuthDBClient
			originalCommonDBClient := dbadapter.CommonDBClient
			dbadapter.CommonDBClient = tt.commonDbAdapter
			dbadapter.AuthDBClient = tt.authDbAdapter
			defer func() {
				dbadapter.CommonDBClient = originalCommonDBClient
				dbadapter.AuthDBClient = originalAuthDBClient
			}()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/subscriber/%s", tt.ueId), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			responseContent := w.Body.String()
			if !reflect.DeepEqual(responseContent, tt.expectedFullResponse) {
				t.Errorf("Expected full response body `%v`, but got `%v`", tt.expectedFullResponse, responseContent)
			}

			if comparePostData(tt.expectedPostDataDetails, postData) != nil {
				t.Errorf("Expected postData `%v`, but got `%v`", tt.expectedPostDataDetails, postData)
			}
		})
	}
}

func TestSubscriberGetHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	AddApiService(router)

	testCases := []struct {
		name         string
		route        string
		dbAdapter    dbadapter.DBInterface
		expectedCode int
		expectedBody string
	}{
		{
			name:         "SubscriberEmptyDB",
			route:        "/api/subscriber",
			dbAdapter:    &MockMongoClientEmptyDB{},
			expectedCode: http.StatusOK,
			expectedBody: "[]",
		},
		{
			name:         "Get subscribers list with one element",
			route:        "/api/subscriber",
			dbAdapter:    &MockMongoClientOneSubscriber{},
			expectedCode: http.StatusOK,
			expectedBody: `[{"plmnID":"12345","ueId":"208930100007487"}]`,
		},
		{
			name:         "ManySubscribers",
			route:        "/api/subscriber",
			dbAdapter:    &MockMongoClientManySubscribers{},
			expectedCode: http.StatusOK,
			expectedBody: `[{"plmnID":"12345","ueId":"208930100007487"},{"plmnID":"54321","ueId":"208930100007488"}]`,
		},
		{
			name:         "SubscriberDBError",
			route:        "/api/subscriber",
			dbAdapter:    &MockMongoClientDBError{},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"failed to retrieve subscribers list"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			origDBClient := dbadapter.CommonDBClient
			dbadapter.CommonDBClient = tc.dbAdapter
			defer func() { dbadapter.CommonDBClient = origDBClient }()
			req, err := http.NewRequest(http.MethodGet, tc.route, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if tc.expectedCode != w.Code {
				t.Errorf("Expected `%v`, got `%v`", tc.expectedCode, w.Code)
			}
			if w.Body.String() != tc.expectedBody {
				t.Errorf("Expected `%v`, got `%v`", tc.expectedBody, w.Body.String())
			}
		})
	}
}

func TestSubscriberPostHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	AddApiService(router)

	testCases := []struct {
		name            string
		route           string
		inputData       string
		expectedMessage configmodels.ConfigMessage
	}{
		{
			name:      "Create a new subscriber success",
			route:     "/api/subscriber/imsi-208930100007487",
			inputData: `{"plmnID":"12345", "opc":"8e27b6af0e692e750f32667a3b14605d","key":"8baf473f2f8fd09487cccbd7097c6862", "sequenceNumber":"16f3b3f70fc2"}`,
			expectedMessage: configmodels.ConfigMessage{
				MsgType:   configmodels.Sub_data,
				MsgMethod: configmodels.Post_op,
				AuthSubData: &models.AuthenticationSubscription{
					AuthenticationManagementField: "8000",
					AuthenticationMethod:          "5G_AKA",
					Milenage: &models.Milenage{
						Op: &models.Op{
							EncryptionAlgorithm: 0,
							EncryptionKey:       0,
						},
					},
					Opc: &models.Opc{
						EncryptionAlgorithm: 0,
						EncryptionKey:       0,
						OpcValue:            "8e27b6af0e692e750f32667a3b14605d",
					},
					PermanentKey: &models.PermanentKey{
						EncryptionAlgorithm: 0,
						EncryptionKey:       0,
						PermanentKeyValue:   "8baf473f2f8fd09487cccbd7097c6862",
					},
					SequenceNumber: "16f3b3f70fc2",
				},
				Imsi: "imsi-208930100007487",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			origChannel := configChannel
			configChannel = make(chan *configmodels.ConfigMessage, 1)
			defer func() { configChannel = origChannel }()
			req, err := http.NewRequest(http.MethodPost, tc.route, strings.NewReader(tc.inputData))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			expectedCode := http.StatusCreated
			expectedBody := "{}"

			if expectedCode != w.Code {
				t.Errorf("Expected `%v`, got `%v`", expectedCode, w.Code)
			}
			if w.Body.String() != expectedBody {
				t.Errorf("Expected `%v`, got `%v`", expectedBody, w.Body.String())
			}
			select {
			case msg := <-configChannel:

				if msg.MsgType != tc.expectedMessage.MsgType {
					t.Errorf("expected MsgType %+v, but got %+v", tc.expectedMessage.MsgType, msg.MsgType)
				}
				if msg.MsgMethod != tc.expectedMessage.MsgMethod {
					t.Errorf("expected MsgMethod %+v, but got %+v", tc.expectedMessage.MsgMethod, msg.MsgMethod)
				}
				if !reflect.DeepEqual(tc.expectedMessage.AuthSubData, msg.AuthSubData) {
					t.Errorf("expected AuthSubData %+v, but got %+v", tc.expectedMessage.AuthSubData, msg.AuthSubData)
				}
				if tc.expectedMessage.Imsi != msg.Imsi {
					t.Errorf("expected IMSI %+v, but got %+v", tc.expectedMessage.Imsi, msg.Imsi)
				}
			default:
				t.Error("expected message in configChannel, but none received")
			}
		})
	}
}

func TestSubscriberDeleteSuccessNoDeviceGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	AddApiService(router)
	origDBClient := dbadapter.CommonDBClient
	dbAdapter := &MockMongoClientEmptyDB{}
	dbadapter.CommonDBClient = dbAdapter
	route := "/api/subscriber/imsi-208930100007487"
	expectedCode := http.StatusNoContent
	expectedBody := ""
	expectedMessage := configmodels.ConfigMessage{
		MsgType:   configmodels.Sub_data,
		MsgMethod: configmodels.Delete_op,
		Imsi:      "imsi-208930100007487",
	}
	origChannel := configChannel
	configChannel = make(chan *configmodels.ConfigMessage, 3)
	defer func() { configChannel = origChannel; dbadapter.CommonDBClient = origDBClient }()
	req, err := http.NewRequest(http.MethodDelete, route, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if expectedCode != w.Code {
		t.Errorf("Expected `%v`, got `%v`", expectedCode, w.Code)
	}
	if expectedBody != w.Body.String() {
		t.Errorf("Expected `%v`, got `%v`", expectedBody, w.Body.String())
	}
	select {
	case msg := <-configChannel:
		if expectedMessage.MsgType != msg.MsgType {
			t.Errorf("expected MsgType %+v, but got %+v", expectedMessage.MsgType, msg.MsgType)
		}
		if expectedMessage.MsgMethod != msg.MsgMethod {
			t.Errorf("expected MsgMethod %+v, but got %+v", expectedMessage.MsgMethod, msg.MsgMethod)
		}
		if expectedMessage.Imsi != msg.Imsi {
			t.Errorf("expected IMSI %+v, but got %+v", expectedMessage.Imsi, msg.Imsi)
		}
	default:
		t.Error("expected message in configChannel, but none received")
	}
	select {
	case msg := <-configChannel:
		t.Errorf("expected no message in configChannel, but got %+v", msg)
	default:
	}
}

func TestSubscriberDeleteFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	AddApiService(router)
	origDBClient := dbadapter.CommonDBClient
	dbAdapter := &MockMongoClientDBError{}
	dbadapter.CommonDBClient = dbAdapter
	route := "/api/subscriber/imsi-208930100007487"
	expectedCode := http.StatusInternalServerError
	expectedBody := `{"error":"error deleting subscriber"}`

	origChannel := configChannel
	configChannel = make(chan *configmodels.ConfigMessage, 1)
	defer func() { configChannel = origChannel; dbadapter.CommonDBClient = origDBClient }()
	req, err := http.NewRequest(http.MethodDelete, route, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if expectedCode != w.Code {
		t.Errorf("Expected `%v`, got `%v`", expectedCode, w.Code)
	}
	if expectedBody != w.Body.String() {
		t.Errorf("Expected `%v`, got `%v`", expectedBody, w.Body.String())
	}
	select {
	case msg := <-configChannel:
		t.Errorf("expected no message in configChannel, but got %+v", msg)
	default:
	}
}

func TestSubscriberDeleteSuccessWithDeviceGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	AddApiService(router)
	origDBClient := dbadapter.CommonDBClient
	dbAdapter := &MockMongoClientDeviceGroupsWithSubscriber{}
	dbadapter.CommonDBClient = dbAdapter
	route := "/api/subscriber/imsi-208930100007487"
	expectedCode := http.StatusNoContent
	expectedBody := ""
	expectedDeviceGroupMessage := configmodels.ConfigMessage{
		MsgType:      configmodels.Device_group,
		MsgMethod:    configmodels.Post_op,
		DevGroupName: "group1",
		DevGroup:     deviceGroupWithoutImsi(),
	}
	expectedMessage := configmodels.ConfigMessage{
		MsgType:   configmodels.Sub_data,
		MsgMethod: configmodels.Delete_op,
		Imsi:      "imsi-208930100007487",
	}
	origChannel := configChannel
	configChannel = make(chan *configmodels.ConfigMessage, 3)
	defer func() { configChannel = origChannel; dbadapter.CommonDBClient = origDBClient }()
	req, err := http.NewRequest(http.MethodDelete, route, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if expectedCode != w.Code {
		t.Errorf("Expected `%v`, got `%v`", expectedCode, w.Code)
	}
	if expectedBody != w.Body.String() {
		t.Errorf("Expected `%v`, got `%v`", expectedBody, w.Body.String())
	}
	select {
	case msg := <-configChannel:
		if expectedDeviceGroupMessage.MsgType != msg.MsgType {
			t.Errorf("expected MsgType %+v, but got %+v", expectedDeviceGroupMessage.MsgType, msg.MsgType)
		}
		if expectedDeviceGroupMessage.MsgMethod != msg.MsgMethod {
			t.Errorf("expected MsgMethod %+v, but got %+v", expectedDeviceGroupMessage.MsgMethod, msg.MsgMethod)
		}
		if expectedDeviceGroupMessage.DevGroupName != msg.DevGroupName {
			t.Errorf("expected device group name %+v, but got %+v", expectedDeviceGroupMessage.DevGroupName, msg.DevGroupName)
		}
		if !reflect.DeepEqual(expectedDeviceGroupMessage.DevGroup.Imsis, msg.DevGroup.Imsis) {
			t.Errorf("expected IMSIs in device group: %+v, but got %+v", expectedDeviceGroupMessage.DevGroup.Imsis, msg.DevGroup.Imsis)
		}
	default:
		t.Error("expected message in configChannel, but none received")
	}
	select {
	case msg := <-configChannel:
		if expectedMessage.MsgType != msg.MsgType {
			t.Errorf("expected MsgType %+v, but got %+v", expectedMessage.MsgType, msg.MsgType)
		}
		if expectedMessage.MsgMethod != msg.MsgMethod {
			t.Errorf("expected MsgMethod %+v, but got %+v", expectedMessage.MsgMethod, msg.MsgMethod)
		}
		if expectedMessage.Imsi != msg.Imsi {
			t.Errorf("expected IMSI %+v, but got %+v", expectedMessage.Imsi, msg.Imsi)
		}
	default:
		t.Error("expected message in configChannel, but none received")
	}
	select {
	case msg := <-configChannel:
		t.Errorf("expected no message in configChannel, but got %+v", msg)
	default:
	}
}

func deviceGroupWithImsis(name string, imsis []string) configmodels.DeviceGroups {
	trafficClass := configmodels.TrafficClassInfo{
		Name: "platinum",
		Qci:  8,
		Arp:  6,
		Pdb:  300,
		Pelr: 6,
	}
	qos := configmodels.DeviceGroupsIpDomainExpandedUeDnnQos{
		DnnMbrUplink:   10000000,
		DnnMbrDownlink: 10000000,
		BitrateUnit:    "kbps",
		TrafficClass:   &trafficClass,
	}
	ipDomain := configmodels.DeviceGroupsIpDomainExpanded{
		Dnn:          "internet",
		UeIpPool:     "172.250.1.0/16",
		DnsPrimary:   "1.1.1.1",
		DnsSecondary: "8.8.8.8",
		Mtu:          1460,
		UeDnnQos:     &qos,
	}
	deviceGroup := configmodels.DeviceGroups{
		DeviceGroupName:  name,
		Imsis:            imsis,
		SiteInfo:         "demo",
		IpDomainName:     "pool1",
		IpDomainExpanded: ipDomain,
	}
	return deviceGroup
}

func deviceGroupWithoutImsi() *configmodels.DeviceGroups {
	tmp := deviceGroupWithImsis("group1", []string{"208930100007487", "208930100007488"})
	tmp.Imsis = slices.Delete(tmp.Imsis, 0, 1)
	return &tmp
}
