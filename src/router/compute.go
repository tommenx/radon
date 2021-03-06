/*
 * Radon
 *
 * Copyright 2018 The Radon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package router

import (
	"fmt"
	"sort"

	"config"

	"github.com/pkg/errors"
)

// HashUniform used to uniform the hash slots to backends.
func (r *Router) HashUniform(table, shardkey string, backends []string) (*config.TableConfig, error) {
	if table == "" {
		return nil, errors.New("table.cant.be.null")
	}
	if shardkey == "" {
		return nil, errors.New("shard.key.cant.be.null")
	}

	slots := r.conf.Slots
	blocks := r.conf.Blocks
	nums := len(backends)
	if nums == 0 {
		return nil, errors.New("router.compute.backends.is.null")
	}
	if nums >= slots {
		return nil, errors.Errorf("router.compute.backends[%d].too.many:[max:%d]", nums, slots)
	}

	// sort backends.
	sort.Strings(backends)
	tableConf := &config.TableConfig{
		Name:       table,
		Slots:      r.conf.Slots,
		Blocks:     r.conf.Blocks,
		ShardKey:   shardkey,
		ShardType:  methodTypeHash,
		Partitions: make([]*config.PartitionConfig, 0, 16),
	}

	slotsPerShard := slots / nums
	tablesPerShard := slotsPerShard / blocks
	for s := 0; s < nums; s++ {
		for i := 0; i < tablesPerShard; i++ {
			step := s * slotsPerShard
			min := i*blocks + step
			max := (i+1)*blocks + step
			if i == tablesPerShard-1 {
				if s == nums-1 {
					max = slots
				} else {
					max = step + slotsPerShard
				}
			}
			name := s*tablesPerShard + i
			partConf := &config.PartitionConfig{
				Table:   fmt.Sprintf("%s_%04d", table, name),
				Segment: fmt.Sprintf("%d-%d", min, max),
				Backend: backends[s],
			}
			tableConf.Partitions = append(tableConf.Partitions, partConf)
		}
	}
	return tableConf, nil
}

// GlobalUniform used to uniform the global table to backends.
func (r *Router) GlobalUniform(table string, backends []string) (*config.TableConfig, error) {
	if table == "" {
		return nil, errors.New("table.cant.be.null")
	}
	nums := len(backends)
	if nums == 0 {
		return nil, errors.New("router.compute.backends.is.null")
	}

	tableConf := &config.TableConfig{
		Name:       table,
		ShardType:  methodTypeGlobal,
		ShardKey:   "",
		Partitions: make([]*config.PartitionConfig, 0, 16),
	}

	for s := 0; s < nums; s++ {
		partConf := &config.PartitionConfig{
			Table:   table,
			Backend: backends[s],
		}
		tableConf.Partitions = append(tableConf.Partitions, partConf)
	}
	return tableConf, nil
}

// SingleUniform used to uniform the single table to backends.
func (r *Router) SingleUniform(table string, backends []string) (*config.TableConfig, error) {
	if table == "" {
		return nil, errors.New("table.cant.be.null")
	}
	nums := len(backends)
	if nums == 0 {
		return nil, errors.New("router.compute.backends.is.null")
	}

	return &config.TableConfig{
		Name:      table,
		ShardType: methodTypeSingle,
		ShardKey:  "",
		Partitions: []*config.PartitionConfig{&config.PartitionConfig{
			Table:   table,
			Backend: backends[0],
		}},
	}, nil
}
