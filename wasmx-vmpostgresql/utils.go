package vmpostgresql

import (
	"encoding/json"

	"github.com/jackc/pgx/v5"
)

func RowsToJSON(rows pgx.Rows) ([]byte, error) {
	results := []map[string]interface{}{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		// Map column names to values
		rowMap := make(map[string]interface{})
		for i, col := range rows.FieldDescriptions() {
			rowMap[col.Name] = values[i]
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return json.Marshal(results)
}
