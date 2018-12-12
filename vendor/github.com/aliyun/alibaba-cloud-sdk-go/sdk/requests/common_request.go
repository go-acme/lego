package requests

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"io"
	"strings"
)

type CommonRequest struct {
	*baseRequest

	Version string
	ApiName string
	Product string

	// roa params
	PathPattern string
	PathParams  map[string]string

	Ontology AcsRequest
}

func NewCommonRequest() (request *CommonRequest) {
	request = &CommonRequest{
		baseRequest: defaultBaseRequest(),
	}
	request.Headers["x-sdk-invoke-type"] = "common"
	request.PathParams = make(map[string]string)
	return
}

func (request *CommonRequest) TransToAcsRequest() {
	if len(request.Version) == 0 {
		errors.NewClientError(errors.MissingParamErrorCode, "Common request [version] is required", nil)
	}
	if len(request.ApiName) == 0 && len(request.PathPattern) == 0 {
		errors.NewClientError(errors.MissingParamErrorCode, "At least one of [ApiName] and [PathPattern] should has a value", nil)
	}
	if len(request.Domain) == 0 && len(request.Product) == 0 {
		errors.NewClientError(errors.MissingParamErrorCode, "At least one of [Domain] and [Product] should has a value", nil)
	}

	if len(request.PathPattern) > 0 {
		roaRequest := &RoaRequest{}
		roaRequest.initWithCommonRequest(request)
		request.Ontology = roaRequest
	} else {
		rpcRequest := &RpcRequest{}
		rpcRequest.baseRequest = request.baseRequest
		rpcRequest.product = request.Product
		rpcRequest.version = request.Version
		rpcRequest.actionName = request.ApiName
		request.Ontology = rpcRequest
	}

}

func (request *CommonRequest) BuildUrl() string {
	if len(request.Port) > 0 {
		return strings.ToLower(request.Scheme) + "://" + request.Domain + ":" + request.Port + request.BuildQueries()
	}

	return strings.ToLower(request.Scheme) + "://" + request.Domain + request.BuildQueries()
}

func (request *CommonRequest) BuildQueries() string {
	return request.Ontology.BuildQueries()
}

func (request *CommonRequest) GetUrl() string {
	if len(request.Port) > 0 {
		return strings.ToLower(request.Scheme) + "://" + request.Domain + ":" + request.Port + request.GetQueries()
	}

	return strings.ToLower(request.Scheme) + "://" + request.Domain + request.GetQueries()
}

func (request *CommonRequest) GetQueries() string {
	return request.Ontology.GetQueries()
}

func (request *CommonRequest) GetBodyReader() io.Reader {
	return request.Ontology.GetBodyReader()
}

func (request *CommonRequest) GetStyle() string {
	return request.Ontology.GetStyle()
}

func (request *CommonRequest) addPathParam(key, value string) {
	request.PathParams[key] = value
}
