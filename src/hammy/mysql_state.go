package hammy

import (
	"fmt"
	"encoding/json"
	"database/sql"
	_ "github.com/ziutek/mymysql/godrv"
)

// Driver for retriving and saving state in MySQL database
// It's assumes the table structure like this:
//
//  CREATE TABLE `states` (
//    `obj_key` varchar(255) NOT NULL,
//    `obj_state` text,
//    `cas` BIGINT NOT NULL DEFAULT 0,
//    PRIMARY KEY (`obj_key`)
//  ) ENGINE=InnoDB DEFAULT CHARSET=utf8
//
type MySQLStateKeeper struct {
	db *sql.DB
	tableName string
}

func NewMySQLStateKeeper(cfg Config) (sk *MySQLStateKeeper, err error) {
	sk = new(MySQLStateKeeper)
	sk.db, err = sql.Open("mymysql", cfg.MySQLStates.Database + "/" + cfg.MySQLStates.User + "/" + cfg.MySQLStates.Password)
	sk.tableName = cfg.MySQLStates.Table

	return
}

func (sk *MySQLStateKeeper) Get(key string) (ans StateKeeperAnswer) {
	var stateRaw []byte
	var cas uint64

	sqlq := fmt.Sprintf("SELECT `obj_state`, `cas` FROM `%s` WHERE `obj_key` = ?", sk.tableName)
	row := sk.db.QueryRow(sqlq, key)
	err := row.Scan(&stateRaw, &cas)

	var s State
	switch err {
		case nil:
			e := json.Unmarshal(stateRaw, &s)
			if e != nil {
				ans.Err = e
				return
			}
		case sql.ErrNoRows:
			// Do nothing
		default:
			ans.Err = err
			return
	}

	ans.State = s
	ans.Cas = &cas
	return
}

func (sk *MySQLStateKeeper) MGet(keys []string) (states map[string]StateKeeperAnswer) {
	states = make(map[string]StateKeeperAnswer)

	n := len(keys)
	// Selecting states by 10 rows
SUBKEYS:	for i := 0; i < n; i += 10 {
		var subkeys []string
		if (i + 10) < n {
			subkeys = keys[i:i+10]
		} else {
			subkeys = keys[i:]
		}

		m := len(subkeys)

		sql := fmt.Sprintf("SELECT `obj_key`, `obj_state`, `cas` FROM `%s` WHERE `obj_key` IN (?", sk.tableName)
		for j := 1; j < m; j++ {
			sql += ", ?"
		}
		sql += ")"

		args := make([]interface{}, m)
		for k, s := range subkeys {
			args[k] = s
		}

		rows, e := sk.db.Query(sql, args...)
		if e != nil {
			for _, k := range subkeys {
				states[k] = StateKeeperAnswer{
					State: nil,
					Cas: nil,
					Err: fmt.Errorf("Query error: %v", e),
				}
			}
			continue
		}

		for rows.Next() {
			var objK string
			var stateRaw []byte
			var cas uint64

			err := rows.Scan(&objK, &stateRaw, &cas)
			if err != nil {
				for _, k := range subkeys {
					states[k] = StateKeeperAnswer{
						State: nil,
						Cas: nil,
						Err: fmt.Errorf("Query error: %v", err),
					}
				}
				continue SUBKEYS
			}

			var s State
			err = json.Unmarshal(stateRaw, &s)
			if err != nil {
				states[objK] = StateKeeperAnswer{
					State: nil,
					Cas: nil,
					Err: fmt.Errorf("Unmarshal error: %v", err),
				}
			} else {
				states[objK] = StateKeeperAnswer{
					State: s,
					Cas: &cas,
					Err: nil,
				}
			}
		}
	}

	for _, k := range keys {
		if _, found := states[k]; !found {
			states[k] = StateKeeperAnswer{
				State: *NewState(),
				Cas: nil,
				Err: nil,
			}
		}
	}

	return
}

func (sk *MySQLStateKeeper) Set(key string, data State, cas *uint64) (retry bool, err error) {
	stateRaw, err := json.Marshal(data)
	if err != nil {
		return
	}

	if cas == nil {
		sqlq := fmt.Sprintf("INSERT INTO `%s` SET `obj_key` = ?, `obj_state` = ?, `cas` = ?", sk.tableName)
		_, e := sk.db.Exec(sqlq, key, stateRaw, 0)
		if e != nil {
			//TODO
			err = e
			return
		}
	} else {
		newCas := *cas + 1
		sqlq := fmt.Sprintf("UPDATE `%s` SET `obj_state` = ?, `cas` = ? WHERE `obj_key` = ? AND `cas` = ?", sk.tableName)
		res, e := sk.db.Exec(sqlq, stateRaw, newCas, key, *cas)
		if e != nil {
			err = e
			return
		}
		rowsAffected, e := res.RowsAffected()
		if e != nil {
			err = e
			return
		}
		if rowsAffected != 1 {
			if rowsAffected > 1 {
				panic("More than one row has been affected")
			}
			retry = true
		}
	}

	return
}