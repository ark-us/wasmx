package vmsql

import (
	"database/sql"
	"encoding/json"
)

func RowsToJSON(rows *sql.Rows) ([]byte, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := []map[string]interface{}{}

	for rows.Next() {
		// Create a slice of empty interfaces to hold each column value
		columnPointers := make([]interface{}, len(cols))
		columnValues := make([]interface{}, len(cols))
		for i := range columnPointers {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		// Map column names to values
		rowMap := make(map[string]interface{})
		for i, col := range cols {
			val := columnValues[i]
			rowMap[col] = val
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return json.Marshal(results)
}
