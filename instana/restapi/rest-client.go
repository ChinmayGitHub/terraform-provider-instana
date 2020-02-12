package restapi

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	resty "gopkg.in/resty.v1"
)

//ErrEntityNotFound error message which is returned when the entity cannot be found at the server
var ErrEntityNotFound = errors.New("Failed to get resource from Instana API. 404 - Resource not found")

//RestClient interface to access REST resources of the Instana API
type RestClient interface {
	GetOne(id string, resourcePath string) ([]byte, error)
	Put(data InstanaDataObject, resourcePath string) ([]byte, error)
	Delete(resourceID string, resourceBasePath string) error
}

//NewClient creates a new instance of the Instana REST API client
func NewClient(apiToken string, host string) RestClient {
	restyClient := resty.New()

	return &restClientImpl{
		apiToken:    apiToken,
		host:        host,
		restyClient: restyClient,
	}
}

type restClientImpl struct {
	apiToken    string
	host        string
	restyClient *resty.Client
}

var emptyResponse = make([]byte, 0)

//GetOne request the resource with the given ID
func (client *restClientImpl) GetOne(id string, resourcePath string) ([]byte, error) {
	url := client.buildResourceURL(resourcePath, id)
	log.Infof("Call GET %s", url)
	resp, err := client.createRequest().Get(url)
	if err != nil {
		return emptyResponse, fmt.Errorf("failed to send HTTP GET request to Instana API; status code = %d; status message = %s, %s", resp.StatusCode(), resp.Status(), err)
	}
	statusCode := resp.StatusCode()
	if statusCode == 404 {
		return emptyResponse, ErrEntityNotFound
	}
	if statusCode < 200 || statusCode >= 300 {
		return emptyResponse, fmt.Errorf("failed to send HTTP GET request to Instana API; status code = %d; status message = %s\nBody: %s", statusCode, resp.Status(), resp.Body())
	}
	return resp.Body(), nil
}

//Put executes a HTTP PUT request to create or update the given resource
func (client *restClientImpl) Put(data InstanaDataObject, resourcePath string) ([]byte, error) {
	url := client.buildResourceURL(resourcePath, data.GetID())
	log.Infof("Call PUT %s", url)
	resp, err := client.createRequest().SetHeader("Content-Type", "application/json; charset=utf-8").SetBody(data).Put(url)
	if err != nil {
		return emptyResponse, fmt.Errorf("failed to send HTTP PUT request to Instana API; status code = %d; status message = %s, %s", resp.StatusCode(), resp.Status(), err)
	}
	statusCode := resp.StatusCode()
	if statusCode < 200 || statusCode >= 300 {
		return emptyResponse, fmt.Errorf("failed to send HTTP PUT request to Instana API; status code = %d; status message = %s\nBody: %s", statusCode, resp.Status(), resp.Body())
	}
	return resp.Body(), nil
}

//Delete executes a HTTP DELETE request to delete the resource with the given ID
func (client *restClientImpl) Delete(resourceID string, resourceBasePath string) error {
	url := client.buildResourceURL(resourceBasePath, resourceID)
	log.Infof("Call DELETE %s", url)
	resp, err := client.createRequest().Delete(url)

	if err != nil {
		return fmt.Errorf("failed to send HTTP DELETE request to Instana API; status code = %d; status message = %s, %s", resp.StatusCode(), resp.Status(), err)
	}
	statusCode := resp.StatusCode()
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("failed to send HTTP DELETE request to Instana API; status code = %d; status message = %s\nBody: %s", statusCode, resp.Status(), resp.Body())
	}
	return nil
}

func (client *restClientImpl) createRequest() *resty.Request {
	return client.restyClient.R().SetHeader("Accept", "application/json").SetHeader("Authorization", fmt.Sprintf("apiToken %s", client.apiToken))
}

func (client *restClientImpl) buildResourceURL(resourceBasePath string, id string) string {
	pattern := "%s/%s"
	if strings.HasSuffix(resourceBasePath, "/") {
		pattern = "%s%s"
	}
	resourcePath := fmt.Sprintf(pattern, resourceBasePath, id)
	return client.buildURL(resourcePath)
}

func (client *restClientImpl) buildURL(resourcePath string) string {
	return fmt.Sprintf("https://%s%s", client.host, resourcePath)
}