// Code generated by go-swagger; DO NOT EDIT.

package project

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"k8c.io/kubermatic/v2/pkg/test/e2e/api/utils/apiclient/models"
)

// GetClusterHealthReader is a Reader for the GetClusterHealth structure.
type GetClusterHealthReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetClusterHealthReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetClusterHealthOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetClusterHealthUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewGetClusterHealthForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewGetClusterHealthDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetClusterHealthOK creates a GetClusterHealthOK with default headers values
func NewGetClusterHealthOK() *GetClusterHealthOK {
	return &GetClusterHealthOK{}
}

/*GetClusterHealthOK handles this case with default header values.

ClusterHealth
*/
type GetClusterHealthOK struct {
	Payload *models.ClusterHealth
}

func (o *GetClusterHealthOK) Error() string {
	return fmt.Sprintf("[GET /api/v1/projects/{project_id}/dc/{dc}/clusters/{cluster_id}/health][%d] getClusterHealthOK  %+v", 200, o.Payload)
}

func (o *GetClusterHealthOK) GetPayload() *models.ClusterHealth {
	return o.Payload
}

func (o *GetClusterHealthOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ClusterHealth)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetClusterHealthUnauthorized creates a GetClusterHealthUnauthorized with default headers values
func NewGetClusterHealthUnauthorized() *GetClusterHealthUnauthorized {
	return &GetClusterHealthUnauthorized{}
}

/*GetClusterHealthUnauthorized handles this case with default header values.

EmptyResponse is a empty response
*/
type GetClusterHealthUnauthorized struct {
}

func (o *GetClusterHealthUnauthorized) Error() string {
	return fmt.Sprintf("[GET /api/v1/projects/{project_id}/dc/{dc}/clusters/{cluster_id}/health][%d] getClusterHealthUnauthorized ", 401)
}

func (o *GetClusterHealthUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewGetClusterHealthForbidden creates a GetClusterHealthForbidden with default headers values
func NewGetClusterHealthForbidden() *GetClusterHealthForbidden {
	return &GetClusterHealthForbidden{}
}

/*GetClusterHealthForbidden handles this case with default header values.

EmptyResponse is a empty response
*/
type GetClusterHealthForbidden struct {
}

func (o *GetClusterHealthForbidden) Error() string {
	return fmt.Sprintf("[GET /api/v1/projects/{project_id}/dc/{dc}/clusters/{cluster_id}/health][%d] getClusterHealthForbidden ", 403)
}

func (o *GetClusterHealthForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewGetClusterHealthDefault creates a GetClusterHealthDefault with default headers values
func NewGetClusterHealthDefault(code int) *GetClusterHealthDefault {
	return &GetClusterHealthDefault{
		_statusCode: code,
	}
}

/*GetClusterHealthDefault handles this case with default header values.

errorResponse
*/
type GetClusterHealthDefault struct {
	_statusCode int

	Payload *models.ErrorResponse
}

// Code gets the status code for the get cluster health default response
func (o *GetClusterHealthDefault) Code() int {
	return o._statusCode
}

func (o *GetClusterHealthDefault) Error() string {
	return fmt.Sprintf("[GET /api/v1/projects/{project_id}/dc/{dc}/clusters/{cluster_id}/health][%d] getClusterHealth default  %+v", o._statusCode, o.Payload)
}

func (o *GetClusterHealthDefault) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetClusterHealthDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}