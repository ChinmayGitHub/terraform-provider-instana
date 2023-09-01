package restapi_test

import (
	"encoding/json"
	"errors"
	"testing"

	. "github.com/gessnerfl/terraform-provider-instana/instana/restapi"
	"github.com/gessnerfl/terraform-provider-instana/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSuccessfulCreateOfTestObjectThroughCreatePUTUpdateNotSupportedRestResource(t *testing.T) {
	executeCreateOrUpdateOperationThroughCreatePUTUpdateNotSupportedRestResourceTest(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		testObject := makeTestObject()
		serializedJSON, _ := json.Marshal(testObject)

		client.EXPECT().Put(gomock.Eq(testObject), gomock.Eq(testObjectResourcePath)).Return(serializedJSON, nil)
		client.EXPECT().Post(gomock.Any(), gomock.Eq(testObjectResourcePath)).Times(0)
		unmarshaller.EXPECT().Unmarshal(serializedJSON).Times(1).Return(testObject, nil)

		result, err := sut.Create(testObject)

		assert.NoError(t, err)
		assert.Equal(t, testObject, result)
	})
}

func TestShouldFailToCreateTestObjectThroughCreatePUTUpdateNotSupportedRestResourceWhenErrorIsReturnedFromRestClient(t *testing.T) {
	executeCreateOrUpdateOperationThroughCreatePUTUpdateNotSupportedRestResourceTest(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		testObject := makeTestObject()

		client.EXPECT().Put(gomock.Eq(testObject), gomock.Eq(testObjectResourcePath)).Return(nil, errors.New("error during test"))
		client.EXPECT().Post(gomock.Any(), gomock.Eq(testObjectResourcePath)).Times(0)
		unmarshaller.EXPECT().Unmarshal(gomock.Any()).Times(0)

		_, err := sut.Create(testObject)

		assert.Error(t, err)
	})
}

func TestShouldFailToCreateTestObjectThroughCreatePUTUpdateNotSupportedRestResourceWhenResponseCannotBeUnmarshalled(t *testing.T) {
	executeCreateOrUpdateOperationThroughCreatePUTUpdateNotSupportedRestResourceTest(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		testObject := makeTestObject()
		expectedError := errors.New("test")

		client.EXPECT().Put(gomock.Eq(testObject), gomock.Eq(testObjectResourcePath)).Return(invalidResponse, nil)
		client.EXPECT().Post(gomock.Any(), gomock.Eq(testObjectResourcePath)).Times(0)
		unmarshaller.EXPECT().Unmarshal(invalidResponse).Times(1).Return(nil, expectedError)

		_, err := sut.Create(testObject)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestShouldFailedToCreateTestObjectThroughCreatePUTUpdateNotSupportedRestResourceWhenAnInvalidTestObjectIsProvided(t *testing.T) {
	executeCreateOrUpdateOperationThroughCreatePUTUpdateNotSupportedRestResourceTest(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		testObject := &testObject{
			ID:   "some id",
			Name: "invalid name",
		}

		client.EXPECT().Put(gomock.Eq(testObject), gomock.Eq(testObjectResourcePath)).Times(0)
		client.EXPECT().Post(gomock.Any(), gomock.Eq(testObjectResourcePath)).Times(0)
		unmarshaller.EXPECT().Unmarshal(gomock.Any()).Times(0)

		_, err := sut.Create(testObject)

		assert.Error(t, err)
	})
}

func TestShouldFailedToCreateTestObjectThroughCreatePUTUpdateNotSupportedRestResourceWhenAnInvalidTestObjectIsReceived(t *testing.T) {
	executeCreateOrUpdateOperationThroughCreatePUTUpdateNotSupportedRestResourceTest(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		object := makeTestObject()

		client.EXPECT().Put(gomock.Eq(object), gomock.Eq(testObjectResourcePath)).Times(1).Return(invalidResponse, nil)
		client.EXPECT().Post(gomock.Any(), gomock.Eq(testObjectResourcePath)).Times(0)
		unmarshaller.EXPECT().Unmarshal(invalidResponse).Times(1).Return(&testObject{ID: object.ID, Name: "invalid"}, nil)

		_, err := sut.Create(object)

		assert.Error(t, err)
	})
}

func TestShouldFailToUpdateTestObjectThroughCreatePUTUpdateNotSupportedRestResourceWhenEmptyObjectCanBeCreated(t *testing.T) {
	executeCreateOrUpdateOperationThroughCreatePUTUpdateNotSupportedRestResourceTest(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		testData := makeTestObject()
		emptyObject := &testObject{}

		unmarshaller.EXPECT().Unmarshal(gomock.Any()).Times(1).Return(emptyObject, nil)

		_, err := sut.Update(testData)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "update is not supported for /test")
	})
}

func TestShouldFailToUpdateTestObjectThroughCreatePUTUpdateNotSupportedRestResourceWhenEmptyObjectCannotBeCreated(t *testing.T) {
	executeCreateOrUpdateOperationThroughCreatePUTUpdateNotSupportedRestResourceTest(t, func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject]) {
		testData := makeTestObject()
		emptyObject := &testObject{}
		unmarshallingError := errors.New("unmarshalling-error")

		unmarshaller.EXPECT().Unmarshal(gomock.Any()).Times(1).Return(emptyObject, unmarshallingError)

		_, err := sut.Update(testData)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "update is not supported for /test; unmarshalling-error")
	})
}

func executeCreateOrUpdateOperationThroughCreatePUTUpdateNotSupportedRestResourceTest(t *testing.T, testFunction func(t *testing.T, sut RestResource[*testObject], client *mocks.MockRestClient, unmarshaller *mocks.MockJSONUnmarshaller[*testObject])) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mocks.NewMockRestClient(ctrl)
	unmarshaller := mocks.NewMockJSONUnmarshaller[*testObject](ctrl)

	sut := NewCreatePUTUpdateNotSupportedRestResource[*testObject](testObjectResourcePath, unmarshaller, client)

	testFunction(t, sut, client, unmarshaller)
}
