package restapi_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"testing"

	. "github.com/gessnerfl/terraform-provider-instana/instana/restapi"
	mocks "github.com/gessnerfl/terraform-provider-instana/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	testObjectResourcePath = "/test"
	testObjectID           = "test-object-id"
	testObjectName         = "test-object-name"
)

type testObject struct {
	ID   string
	Name string
}

func (t *testObject) GetIDForResourcePath() string {
	return t.ID
}

func (t *testObject) Validate() error {
	if t.Name != testObjectName {
		return errors.New("Name differs")
	}
	return nil
}

func makeTestObject() *testObject {
	return &testObject{ID: testObjectID, Name: testObjectName}
}

func TestSuccessfulGetOneTestObjectThroughDefaultRestResource(t *testing.T) {
	executeForAllImplementationsOfDefaultRestResource(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		testObject := makeTestObject()
		serializedJSON, _ := json.Marshal(testObject)

		client.EXPECT().GetOne(gomock.Eq(testObject.ID), gomock.Eq(testObjectResourcePath)).Return(serializedJSON, nil)
		unmarshaller.EXPECT().Unmarshal(serializedJSON).Times(1).Return(testObject, nil)

		data, err := sut.GetOne(testObject.ID)

		assert.NoError(t, err)
		assert.Equal(t, testObject, data)
	})
}

func TestShouldFailToGetOneTestObjectThroughDefaultRestResourceWhenErrorIsRetrievedFromRestClient(t *testing.T) {
	executeForAllImplementationsOfDefaultRestResource(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		client.EXPECT().GetOne(gomock.Eq(testObjectID), gomock.Eq(testObjectResourcePath)).Return(nil, errors.New("error during test"))
		unmarshaller.EXPECT().Unmarshal(gomock.Any()).Times(0)

		_, err := sut.GetOne(testObjectID)

		assert.Error(t, err)
	})
}

func TestShouldFailToGetOneTestObjectThroughDefaultRestResourceWhenResponseCannotBeUnmarshalled(t *testing.T) {
	executeForAllImplementationsOfDefaultRestResource(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		expectedError := errors.New("test")
		response := []byte("[{ \"invalid\" : \"data\" }]")

		client.EXPECT().GetOne(gomock.Eq(testObjectID), gomock.Eq(testObjectResourcePath)).Return(response, nil)
		unmarshaller.EXPECT().Unmarshal(response).Times(1).Return(nil, expectedError)

		_, err := sut.GetOne(testObjectID)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestShouldFailToGetOneTestObjectThroughDefaultRestResourceWhenUnmarshalledObjectIsNotValid(t *testing.T) {
	executeForAllImplementationsOfDefaultRestResource(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		response := []byte("[{ \"some\" : \"data\" }]")
		object := makeTestObject()
		object.Name = "invalid"

		client.EXPECT().GetOne(gomock.Eq(testObjectID), gomock.Eq(testObjectResourcePath)).Return(response, nil)
		unmarshaller.EXPECT().Unmarshal(response).Times(1).Return(object, nil)

		_, err := sut.GetOne(testObjectID)

		assert.Error(t, err)
	})
}

func TestSuccessfulDeleteOfTestObjectByObjectThroughDefaultRestResource(t *testing.T) {
	executeForAllImplementationsOfDefaultRestResource(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		testObject := makeTestObject()

		client.EXPECT().Delete(gomock.Eq(testObjectID), gomock.Eq(testObjectResourcePath)).Return(nil)

		err := sut.Delete(testObject)

		assert.NoError(t, err)
	})
}

func TestShouldFailToDeleteTestObjectThroughDefaultRestResourceWhenErrorIsRetrunedFromRestClient(t *testing.T) {
	executeForAllImplementationsOfDefaultRestResource(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		testObject := makeTestObject()

		client.EXPECT().Delete(gomock.Eq(testObjectID), gomock.Eq(testObjectResourcePath)).Return(errors.New("Error during test"))

		err := sut.Delete(testObject)

		assert.Error(t, err)
	})
}

func executeForAllImplementationsOfDefaultRestResource(t *testing.T, testFunc func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject])) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mocks.NewMockRestClient(ctrl)
	unmarshaller := mocks.NewMockJSONUnmarshaller[*testObject](ctrl)

	implementations := map[DefaultRestResourceMode]RestResource[*testObject]{
		DefaultRestResourceModeCreateAndUpdatePUT:  NewCreatePUTUpdatePUTRestResource[*testObject](testObjectResourcePath, unmarshaller, client),
		DefaultRestResourceModeCreatePOSTUpdatePUT: NewCreatePOSTUpdatePUTRestResource[*testObject](testObjectResourcePath, unmarshaller, client),
	}

	caller := getCallerName()
	for k, v := range implementations {
		t.Run(fmt.Sprintf("%s[%s]", caller, k), func(t *testing.T) {
			testFunc(t, v, client, unmarshaller)
		})
	}
}

func getCallerName() string {
	pc, _, _, ok := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		return details.Name()
	}
	return "undefined"
}
