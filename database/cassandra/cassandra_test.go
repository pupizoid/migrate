package cassandra

import (
	"context"
	"fmt"
	"strconv"
	"testing"
)

import (
	"github.com/dhui/dktest"
	"github.com/gocql/gocql"
)

import (
	dt "github.com/golang-migrate/migrate/v4/database/testing"
	"github.com/golang-migrate/migrate/v4/dktesting"
)

var (
	opts  = dktest.Options{PortRequired: true, ReadyFunc: isReady}
	specs = []dktesting.ContainerSpec{
		{ImageName: "cassandra:3.0.10", Options: opts},
		{ImageName: "cassandra:3.0", Options: opts},
	}
)

func isReady(ctx context.Context, c dktest.ContainerInfo) bool {
	// Cassandra exposes 5 ports (7000, 7001, 7199, 9042 & 9160)
	// We only need the port bound to 9042
	ip, portStr, err := c.Port(9042)
	if err != nil {
		return false
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false
	}

	cluster := gocql.NewCluster(ip)
	cluster.Port = port
	cluster.Consistency = gocql.All
	p, err := cluster.CreateSession()
	if err != nil {
		return false
	}
	defer p.Close()
	// Create keyspace for tests
	if err = p.Query("CREATE KEYSPACE testks WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor':1}").Exec(); err != nil {
		return false
	}
	return true
}

func Test(t *testing.T) {
	dktesting.ParallelTest(t, specs, func(t *testing.T, c dktest.ContainerInfo) {
		ip, port, err := c.Port(9042)
		if err != nil {
			t.Fatal("Unable to get mapped port:", err)
		}
		addr := fmt.Sprintf("cassandra://%v:%v/testks", ip, port)
		p := &Cassandra{}
		d, err := p.Open(addr)
		if err != nil {
			t.Fatalf("%v", err)
		}
		defer d.Close()
		dt.Test(t, d, []byte("SELECT table_name from system_schema.tables"))
	})
}
