package database

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"

	"github.com/golang-base-template/util/database/client"
)

const (
	MasterReplication = "master"
	SlaveReplication  = "slave"
)

type (
	// Database wrapper
	Database interface {
		GetContext(context.Context, interface{}, string, ...interface{}) error
		Ping() error
		Rebind(string) string
		SelectContext(context.Context, interface{}, string, ...interface{}) (interface{}, error)

		// Transaction Stuff
		Beginx() (Tx, error)

		GetConn() error
	}

	database struct {
		DB          client.DatabaseConf
		dbName      string
		replication string
	}

	// Tx Wrapper
	Tx interface {
		Commit() error
		Rollback() error
		QueryRowContext(ctx context.Context, query string, destination []interface{}, args ...interface{}) error
	}
	tx struct {
		clientTx *sql.Tx
	}
)

// GetDB init interface db
func GetDB(name string, replication string) (Database, error) {
	obj := &database{}

	dbConn, err := client.GetConnection(name)
	if err != nil {
		return obj, errors.Wrapf(err, "[DB][GetDB] fail to get connection %s", name)
	}

	obj.DB.Master = dbConn.Master
	obj.DB.Slave = dbConn.Slave
	obj.replication = replication

	obj.dbName = name

	return obj, nil
}

func isMaster(replication string) bool {
	if replication == MasterReplication {
		return true
	}
	return false
}

func (d *database) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if isMaster(d.replication) {
		return d.DB.Master.GetContext(ctx, dest, query, args...)
	}
	return d.DB.Slave.GetContext(ctx, dest, query, args...)
}

func (d *database) Ping() error {
	if isMaster(d.replication) {
		return d.DB.Master.Ping()
	}
	return d.DB.Slave.Ping()
}

func (d *database) Rebind(s string) string {
	if isMaster(d.replication) {
		return d.DB.Master.Rebind(s)
	}
	return d.DB.Slave.Rebind(s)
}

func (d *database) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) (result interface{}, err error) {
	if isMaster(d.replication) {
		err = d.DB.Master.SelectContext(ctx, dest, query, args)
	} else {
		err = d.DB.Slave.SelectContext(ctx, dest, query, args)
	}

	if err != nil {
		return nil, err
	}

	return dest, nil
}

func (d *database) Beginx() (Tx, error) {
	clientTx, err := d.Beginx()
	if err != nil {
		return nil, err
	}
	return clientTx, nil
}

// GetConn is a func for checking DB object is nil or not to prevent nil pointer
func (d *database) GetConn() error {
	if d.DB.Master == nil || d.DB.Slave == nil {
		return errors.New("db not initialized")
	}
	return nil
}

func (t *tx) QueryRowContext(ctx context.Context, query string, destination []interface{}, args ...interface{}) (err error) {
	return t.clientTx.QueryRowContext(ctx, query, args...).Scan(destination...)
}

func (t *tx) Commit() error {
	return t.clientTx.Commit()
}

func (t *tx) Rollback() error {
	return t.clientTx.Rollback()
}
