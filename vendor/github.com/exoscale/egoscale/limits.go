package egoscale

func (*ListResourceLimits) name() string {
	return "listResourceLimits"
}

func (*ListResourceLimits) response() interface{} {
	return new(ListResourceLimitsResponse)
}

func (*UpdateResourceLimit) name() string {
	return "updateResourceLimit"
}

func (*UpdateResourceLimit) response() interface{} {
	return new(UpdateResourceLimitResponse)
}

func (*GetAPILimit) name() string {
	return "getAPILimit"
}

func (*GetAPILimit) response() interface{} {
	return new(GetAPILimitResponse)
}
