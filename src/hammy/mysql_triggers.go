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
//    `host` varchar(255) NOT NULL,
//    `trigger` text,
//    PRIMARY KEY (`host`)
//  ) ENGINE=InnoDB DEFAULT CHARSET=utf8
//
type MySQLTriggersGetter struct {
	db *sql.DB
	tableName string
	pool chan int
}

func NewMySQLTriggersGetter(cfg Config) (tg *MySQLTriggersGetter, err error) {
	tg = new(MySQLTriggersGetter)
	tg.db, err = sql.Open("mymysql", cfg.MySQLTriggers.Database + "/" + cfg.MySQLTriggers.User + "/" + cfg.MySQLTriggers.Password)
	if err != nil {
		return
	}

	tg.tableName = cfg.MySQLTriggers.Table

	tg.pool = make(chan int, cfg.MySQLTriggers.MaxConn)
	for i := 0; i < cfg.MySQLTriggers.MaxConn; i++ {
		tg.pool <- 1
	}

	return
}

func (tg *MySQLTriggersGetter) MGet(keys []string) (triggers map[string]string, err error) {
	// Pool limits
	<- tg.pool
	defer func() {
		tg.pool <- 1
	}()

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

		sql := fmt.Sprintf("SELECT `host`, `trigger` FROM `%s` WHERE `host` IN (?", tg.tableName)
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
