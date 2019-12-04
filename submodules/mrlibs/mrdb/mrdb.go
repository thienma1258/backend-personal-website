package mrdb

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	mgo "gopkg.in/mgo.v2"
)

const DB_MANGA_ROCK_2 = 101
const DB_MANGA_ROCK_2_HEAVY = 102
const DB_MANGA_ROCK_2_WRITE = 103
const DB_MANGA_ROCK_SOURCE = 104
const DB_ADDONS = 105
const DB_BETA = 106
const DB_MONGO_MANGAROCK = 107
const DB_ADS_CONFIG = 108
const DB_PUSH_READ = 109
const DB_PUSH_WRITE = 110

// MysqlConnectionSetting Connection settings for mysql
type MysqlConnectionSetting struct {
	ConnectionString string
	MaxIdleConns     int
	MaxOpenConns     int
	ConnMaxLifetime  time.Duration
	PingBeforeUsing  bool
	conn             *sql.DB
}

// MongoConnectionSetting Connection settings for mongodb
type MongoConnectionSetting struct {
	ConnectionString string
	conn             *mgo.Session
}

var mysqlConns = make(map[int]*MysqlConnectionSetting)
var mongoConns = make(map[int]*MongoConnectionSetting)

// RegisterMysqlConnection register a mysql connection
func RegisterMysqlConnection(connection int, setting *MysqlConnectionSetting) {
	conn, err := sql.Open("mysql", setting.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	conn.SetMaxIdleConns(setting.MaxIdleConns)
	conn.SetMaxOpenConns(setting.MaxOpenConns)
	conn.SetConnMaxLifetime(setting.ConnMaxLifetime)
	setting.conn = conn
	mysqlConns[connection] = setting
}

// RegisterMongoConnection register a mongo connection
func RegisterMongoConnection(connection int, setting *MongoConnectionSetting) {
	conn, err := mgo.Dial(setting.ConnectionString)
	if err != nil {
		panic(err)
	}
	conn.SetMode(mgo.Monotonic, true)

	setting.conn = conn
	mongoConns[connection] = setting
}

// GetMysqlDB Get a mysql connection
func GetMysqlDB(connection int) *sql.DB {
	if setting, ok := mysqlConns[connection]; ok {
		conn := setting.conn
		if setting.PingBeforeUsing {
			conn.Ping()
		}
		return conn
	}
	return nil
}

// GetMongoDB Get a mysql connection
func GetMongoDB(connection int) *mgo.Session {
	if setting, ok := mongoConns[connection]; ok {
		return setting.conn
	}
	return nil
}
