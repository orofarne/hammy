package hammy

import (
	"fmt"
	"database/sql"
	_ "github.com/ziutek/mymysql/godrv"
)

// Driver for retriving triggers from MySQL database
// It's assumes the table structure like this:
//
//  CREATE TABLE `triggers` (
//    `obj_key` varchar(255) NOT NULL,
//    `obj_trigger` text,
//    PRIMARY KEY (`obj_key`)
//  ) ENGINE=InnoDB DEFAULT CHARSET=utf8
//
type MySQLTriggersGetter struct {
	db *sql.DB
	tableName string
}

func NewMySQLTriggersGetter(cfg Config) (tg *MySQLTriggersGetter, err error) {
	tg = new(MySQLTriggersGetter)
	tg.db, err = sql.Open("mymysql", cfg.MySQLTriggers.Database + "/" + cfg.MySQLTriggers.User + "/" + cfg.MySQLTriggers.Password)
	tg.tableName = cfg.MySQLTriggers.Table

	return
}

func (tg *MySQLTriggersGetter) MGet(keys []string) (triggers map[string]string, err error) {
	triggers = make(map[string]string)

	n := len(keys)
	// Selecting triggers by 10 rows
	for i := 0; i < n; i += 10 {
		var subkeys []string
		if (i + 10) < n {
			subkeys = keys[i:i+10]
		} else {
			subkeys = keys[i:]
		}

		m := len(subkeys)

		sql := fmt.Sprintf("SELECT `obj_key`, `obj_trigger` FROM `%s` WHERE `obj_key` IN (?", tg.tableName)
		for j := 1; j < m; j++ {
			sql += ", ?"
		}
		sql += ")"

		args := make([]interface{}, m)
		for k, s := range subkeys {
			args[k] = s
		}

		rows, e := tg.db.Query(sql, args...)
		if e != nil {
			err = e
			return
		}

		for rows.Next() {
			var k, tr string
			err = rows.Scan(&k, &tr)
			if err != nil {
				return
			}

			triggers[k] = tr
		}
	}

	return
}