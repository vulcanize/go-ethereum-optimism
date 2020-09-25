// Copyright 2020 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package postgres

import (
	"fmt"
	"time"
)

// Config holds Postgres connection pool configuration params
type Config struct {
	Database string
	Hostname string
	Port     int
	User     string
	Password string

	// Optimization parameters
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration

	// Node info
	NodeInfo *NodeInfo
}

// NodeInfo struct holds info about the geth node
type NodeInfo struct {
	GenesisBlock string
	NetworkID    string
	ChainID      uint64
	ID           string
	ClientName   string
}

// NewConfig returns a new config struct from provided params
func NewConfig(database, hostname, password, user string, port, maxOpen, maxIdle int, maxLifetime time.Duration, info *NodeInfo) *Config {
	return &Config{
		Database:    database,
		Hostname:    hostname,
		Port:        port,
		User:        user,
		Password:    password,
		MaxOpen:     maxOpen,
		MaxLifetime: maxLifetime,
		MaxIdle:     maxIdle,
		NodeInfo:    info,
	}
}

// DbConnectionString resolves Postgres config params to a connection string
func DbConnectionString(config *Config) string {
	if len(config.User) > 0 && len(config.Password) > 0 {
		return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			config.User, config.Password, config.Hostname, config.Port, config.Database)
	}
	if len(config.User) > 0 && len(config.Password) == 0 {
		return fmt.Sprintf("postgresql://%s@%s:%d/%s?sslmode=disable",
			config.User, config.Hostname, config.Port, config.Database)
	}
	return fmt.Sprintf("postgresql://%s:%d/%s?sslmode=disable", config.Hostname, config.Port, config.Database)
}
