package scripts

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

func GetFullTable(db *sqlx.DB, tableName string) ([]map[string]interface{}, error) {
	rows, err := db.Queryx(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	datas := []map[string]interface{}{}
	for rows.Next() {
		data := map[string]interface{}{}
		if err := rows.MapScan(data); err != nil {
			return nil, err
		}
		datas = append(datas, data)
	}
	return datas, nil
}
