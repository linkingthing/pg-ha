package config

import (
	"github.com/zdnscloud/cement/configure"
)

type DBRole string

const (
	DBRoleMaster    DBRole = "master"
	DBRoleSlave     DBRole = "slave"
	DBRoleTmpMaster DBRole = "tmp_master"
	DBRoleSingle    DBRole = "single"
)

type PGHAConfig struct {
	Path    string      `yaml:"-"`
	Server  ServerConf  `yaml:"server"`
	DB      DBConf      `yaml:"db"`
	PGAgent PGAgentConf `yaml:"pg_agent"`
	DDICtrl DDICtrlConf `yaml:"ddi_ctrl"`
}

type ServerConf struct {
	Role     DBRole `yaml:"role"`
	MasterIP string `yaml:"master_ip"`
	SlaveIP  string `yaml:"slave_ip"`
}

type DBConf struct {
	ContainerName string `yaml:"container_name"`
	Name          string `yaml:"name"`
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	Port          uint32 `yaml:"port"`
	VolumeName    string `yaml:"volume_name"`
}

type PGAgentConf struct {
	Addr string `yaml:"addr"`
}

type DDICtrlConf struct {
	MasterAddr string `yaml:"master_addr"`
	SlaveAddr  string `yaml:"slave_addr"`
}

var gConf *PGHAConfig

func Load(path string) (*PGHAConfig, error) {
	var conf PGHAConfig
	conf.Path = path
	if err := conf.Reload(); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (c *PGHAConfig) Reload() error {
	var newConf PGHAConfig
	if err := configure.Load(&newConf, c.Path); err != nil {
		return err
	}

	newConf.Path = c.Path
	*c = newConf
	gConf = &newConf
	return nil
}

func Get() *PGHAConfig {
	return gConf
}
