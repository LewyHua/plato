package ipconf

import "github.com/lewyhua/plato/ipconf/domain"

func top5Endpoints(eds []*domain.Endpoint) []*domain.Endpoint {
	if len(eds) <= 5 {
		return eds
	}
	return eds[:5]
}

func packRes(eds []*domain.Endpoint) Response {
	return Response{
		Message: "OK",
		Code:    0,
		Data:    eds,
	}
}
