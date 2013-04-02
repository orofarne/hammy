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
//  CREATE TABLE `history_host` (
//    `id` int(11) NOT NULL AUTO_INCREMENT,
//    `name` varchar(255) NOT NULL,
//    PRIMARY KEY (`id`),
//    UNIQUE KEY `by_name` (`name`)
//  ) ENGINE=InnoDB DEFAULT CHARSET=utf8
//
//  CREATE TABLE `history_item` (
//    `id` int(11) NOT NULL AUTO_INCREMENT,
//    `host_id` int(11) NOT NULL,
//    `name` varchar(255) NOT NULL,
//    PRIMARY KEY (`id`),
//    UNIQUE KEY `by_name` (`host_id`, `name`)
//  ) ENGINE=InnoDB DEFAULT CHARSET=utf8
//
//  CREATE TABLE `history` (
//    `item_id` int(11) NOT NULL,
//    `timestamp` DATETIME NOT NULL,
//    `value` DOUBLE NOT NULL,
//    PRIMARY KEY (`item_id`, `timestamp`)
//  ) ENGINE=InnoDB DEFAULT CHARSET=utf8
//
//  CREATE TABLE `history_log` (
//    `item_id` int(11) NOT NULL,
//    `timestamp` DATETIME NOT NULL,
//    `value` TEXT NOT NULL,
//    PRIMARY KEY (`item_id`, `timestamp`)
//  ) ENGINE=InnoDB DEFAULT CHARSET=utf8
//
type MySQLSaver struct {
	db *sql.DB
	tableName string
	logTableName string
	hostTableName string
	itemTableName string
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
	s.hostTableName = cfg.MySQLSaver.HostTable
	s.itemTableName = cfg.MySQLSaver.ItemTable

	s.pool = make(chan int, cfg.MySQLSaver.MaxConn)
	for i := 0; i < cfg.MySQLSaver.MaxConn; i++ {
		s.pool <- 1
	}

	return
}

func (s *MySQLSaver) getOrCreateHost(host string) (int, error) {
	sqlp := fmt.Sprintf("SELECT `id` FROM `%s` WHERE `name` = ?", s.hostTableName)
	row := s.db.QueryRow(sqlp, host)

	var hostId int
	err := row.Scan(&hostId)

	switch err {
		case nil:
			// Do nothing
		case sql.ErrNoRows:
			sqlp = fmt.Sprintf("INSERT INTO `%s` SET `name` = ?", s.hostTableName)
			_, err := s.db.Exec(sqlp, host)
			if err != nil {
				return 0, err
			}

			row = s.db.QueryRow("SELECT LAST_INSERT_ID()")
			err = row.Scan(&hostId)
			if err != nil {
				return 0, err
			}
		default:
			return 0, err
	}

	return hostId, nil
}

func (s *MySQLSaver) getOrCreateItem(host_id int, item string) (int, error) {
	sqlp := fmt.Sprintf("SELECT `id` FROM `%s` WHERE `host_id` = ? AND `name` = ?", s.itemTableName)
	row := s.db.QueryRow(sqlp, host_id, item)

	var itemId int
	err := row.Scan(&itemId)

	switch err {
		case nil:
			// Do nothing
		case sql.ErrNoRows:
			sqlp = fmt.Sprintf("INSERT INTO `%s` SET `host_id` = ?, `name` = ?", s.itemTableName)
			_, err := s.db.Exec(sqlp, host_id, item)
			if err != nil {
				return 0, err
			}

			row = s.db.QueryRow("SELECT LAST_INSERT_ID()")
			err = row.Scan(&itemId)
			if err != nil {
				return 0, err
			}
		default:
			return 0, err
	}

	return itemId, nil
}

func (s *MySQLSaver) Push(data *IncomingData) {
	// Pool limits
	<- s.pool
	defer func() {
		s.pool <- 1
	}()

	for hostK, hostV := range *data {
		for itemK, itemV := range hostV {
			for _, v := range itemV {
				hostId, err := s.getOrCreateHost(hostK)
				if err != nil {
					log.Printf("MySQLSaver can't get host_id: %v", err)
					continue
				}

				itemId, err := s.getOrCreateItem(hostId, itemK)
				if err != nil {
					log.Printf("MySQLSaver can't get item_id: %v", err)
					continue
				}

				var tName string
				if strings.HasSuffix(itemK, "#log") {
					tName = s.logTableName
				} else {
					tName = s.tableName
				}
				sqlp := fmt.Sprintf("INSERT INTO `%s` SET `item_id` = ?, `timestamp` = FROM_UNIXTIME(?), `value` = ?", tName)
				_, err = s.db.Exec(sqlp, itemId, v.Timestamp, v.Value)
				if err != nil {
					log.Printf("MySQLSaver error: %v", err)
				}
			}
		}
	}
}
