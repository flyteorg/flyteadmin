package authzserver

import "github.com/ory/fosite"

func interfaceSliceToStringSlice(raw []interface{}) []string {
	res := make([]string, 0, len(raw))
	for _, item := range raw {
		res = append(res, item.(string))
	}

	return res
}


func toClientIface(clients map[string]*fosite.DefaultClient) map[string]fosite.Client {
	res := make(map[string]fosite.Client, len(clients))
	for clientID, client := range clients {
		res[clientID] = client
	}

	return res
}