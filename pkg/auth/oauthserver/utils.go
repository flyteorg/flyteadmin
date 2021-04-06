package oauthserver

func interfaceSliceToStringSlice(raw []interface{}) []string {
	res := make([]string, 0, len(raw))
	for _, item := range raw {
		res = append(res, item.(string))
	}

	return res
}
