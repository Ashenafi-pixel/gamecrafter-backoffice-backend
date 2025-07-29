package logs

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgtype"
)

func convertToPgJSON(data interface{}) (pgtype.JSON, error) {
	var jsonData pgtype.JSON

	if data == nil {
		jsonData.Status = pgtype.Null
		return jsonData, nil
	}

	err := jsonData.Set(data)
	if err != nil {
		return pgtype.JSON{}, fmt.Errorf("failed to convert interface{} to pgtype.JSON: %v", err)
	}

	return jsonData, nil
}

func convertPgJSONToInterface(jsonData pgtype.JSON) (interface{}, error) {
	switch jsonData.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Present:
		var result interface{}
		err := json.Unmarshal(jsonData.Bytes, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal pgtype.JSON to interface{}: %v", err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown pgtype.JSON status: %v", jsonData.Status)
	}
}
