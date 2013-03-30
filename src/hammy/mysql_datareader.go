package hammy

import (
	"fmt"
	"strings"
	"database/sql"
	_ "github.com/ziutek/mymysql/godrv"
)

// Driver for retriving historical data from MySQL database
// See mysql_saver.go for details about schema
type MySQLDataReader struct {
	db *sql.DB
	tableName string
	logTableName string
	pool chan int
}

func NewMySQLDataReader(cfg Config) (dr *MySQLDataReader, err error) {
	dr = new(MySQLDataReader)
	dr.db, err = sql.Open("mymysql", cfg.MySQLDataReader.Database + "/" + cfg.MySQLDataReader.User + "/" + cfg.MySQLDataReader.Password)
	if err != nil {
		return
	}

	dr.tableName = cfg.MySQLDataReader.Table
	dr.logTableName = cfg.MySQLDataReader.LogTable

	dr.pool = make(chan int, cfg.MySQLDataReader.MaxConn)
	for i := 0; i < cfg.MySQLDataReader.MaxConn; i++ {
		dr.pool <- 1
	}

	return
}

func (dr *MySQLDataReader) Read(hostKey string, itemKey string, from uint64, to uint64) (data []IncomingValueData, err error) {
	data = make([]IncomingValueData, 0)

	var tName string
	logValue := false
	if strings.HasSuffix(itemKey, "#log") {
		tName = dr.logTableName
		logValue = true
	} else {
		tName = dr.tableName
	}

	sqlq := fmt.Sprintf("SELECT UNIX_TIMESTAMP(`timestamp`), `value` FROM `%s` WHERE `host` = ? AND `item` = ? AND `timestamp` >= FROM_UNIXTIME(?) AND `timestamp` <= FROM_UNIXTIME(?) ORDER BY `timestamp`", tName)
	rows, err := dr.db.Query(sqlq, hostKey, itemKey, from, to)
	if err != nil {
		return
	}

	for rows.Next() {
		var ts uint64
		var value interface{}

		if logValue {
			var val string
			err = rows.Scan(&ts, &val)
			if err != nil {
				return
			}
			value = val
		} else {
			var val float64
			err = rows.Scan(&ts, &val)
			if err != nil {
				return
			}
			value = val
		}

		elem := IncomingValueData{
			Timestamp: ts,
			Value: value,
		}
		data = append(data, elem)
	}

	return
}
