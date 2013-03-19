package hammy

import (
	"fmt"
	"log"
	"strings"
	"database/sql"
	_ "github.com/ziutek/mymysql/godrv"
)

// Driver for saving historical data in MySQL database
// It's assumes the tables structure like this:
//
//  CREATE TABLE `history` (
//    `obj_key` varchar(255) NOT NULL,
//    `item_key` varchar(255) NOT NULL,
//    `timestamp` DATETIME NOT NULL,
//    `value` DOUBLE NOT NULL,
//    PRIMARY KEY (`obj_key`, `item_key`, `timestamp`)
//  ) ENGINE=InnoDB DEFAULT CHARSET=utf8
//
//  CREATE TABLE `history_log` (
//    `obj_key` varchar(255) NOT NULL,
//    `item_key` varchar(255) NOT NULL,
//    `timestamp` DATETIME NOT NULL,
//    `value` TEXT NOT NULL,
//    PRIMARY KEY (`obj_key`, `item_key`, `timestamp`)
//  ) ENGINE=InnoDB DEFAULT CHARSET=utf8
//
type MySQLSaver struct {
	db *sql.DB
	tableName string
	logTableName string
	pool chan int
}

func NewMySQLSaver(cfg Config) (s *MySQLSaver, err error) {
	s = new(MySQLSaver)
	s.db, err = sql.Open("mymysql", cfg.MySQLSaver.Database + "/" + cfg.MySQLSaver.User + "/" + cfg.MySQLSaver.Password)
	if err != nil {
		return
	}

	s.tableName = cfg.MySQLSaver.Table
	s.logTableName = cfg.MySQLSaver.LogTable

	s.pool = make(chan int, cfg.MySQLSaver.MaxConn)
	for i := 0; i < cfg.MySQLSaver.MaxConn; i++ {
		s.pool <- 1
	}

	return
}

func (s *MySQLSaver) Push(data *IncomingData) {
	// Pool limits
	<- s.pool
	defer func() {
		s.pool <- 1
	}()

	for objK, objV := range *data {
		for itemK, itemV := range objV {
			for _, v := range itemV {
				var tName string
				if strings.HasSuffix(itemK, "#log") {
					tName = s.logTableName
				} else {
					tName = s.tableName
				}
				sqlp := fmt.Sprintf("INSERT INTO `%s` SET `obj_key` = ?, `item_key` = ?, `timestamp` = FROM_UNIXTIME(?), `value` = ?", tName)
				_, err := s.db.Exec(sqlp, objK, itemK, v.Timestamp, v.Value)
				if err != nil {
					log.Printf("MySQLSaver error: %v", err)
				}
			}
		}
	}
}