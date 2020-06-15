package handler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	pb "github.com/linkingthing/pg-agent/pkg/proto"
	"github.com/zdnscloud/cement/log"

	"github.com/linkingthing/pg-ha/config"
)

const (
	RunPGContainer = "docker run -d --rm --name %s -e POSTGRES_PASSWORD=%s -e POSTGRES_USER=%s -e POSTGRES_DB=%s -p %d:%d -v %s:/var/lib/postgresql/data -v /etc/localtime:/etc/localtime postgres"
	PGConnStr      = "user=%s password=%s host=localhost port=%d database=%s sslmode=disable pool_max_conns=10"
	SyncCommand    = "rsync -ae \"ssh -o StrictHostKeyChecking=no\" --delete %s %s:%s"
)

type PGProxy struct {
	grpcClient      pb.PGManagerClient
	dbDataDir       string
	dbName          string
	dbUser          string
	dbPass          string
	dbPort          uint32
	anotherIP       string
	dbContainerName string
}

func newPGProxy(conf *config.PGHAConfig, conn *grpc.ClientConn) *PGProxy {
	anotherIP := conf.Server.SlaveIP
	if conf.Server.Role == config.DBRoleSlave {
		anotherIP = conf.Server.MasterIP
	}

	return &PGProxy{
		grpcClient:      pb.NewPGManagerClient(conn),
		dbDataDir:       conf.DB.DataDir,
		dbName:          conf.DB.Name,
		dbUser:          conf.DB.User,
		dbPass:          conf.DB.Password,
		dbPort:          conf.DB.Port,
		anotherIP:       anotherIP,
		dbContainerName: conf.DB.ContainerName,
	}
}

func (p *PGProxy) getAnotherIP() string {
	return p.anotherIP
}

func (p *PGProxy) genPGConfigFile(isMaster bool, isSlave bool) error {
	_, err := p.grpcClient.UpdatePostgresqlConf(context.TODO(), &pb.UpdatePostgresqlConfRequest{
		Host:     p.anotherIP,
		User:     p.dbUser,
		Password: p.dbPass,
		Port:     p.dbPort,
		IsMaster: isMaster,
		IsSlave:  isSlave,
	})
	if err != nil {
		log.Errorf("update postgresql.conf failed: %s", err.Error())
		return err
	}

	if _, err := p.grpcClient.UpdatePGHBAConf(context.TODO(), &pb.UpdatePGHBAConfRequest{User: p.dbUser, AnotherIp: p.anotherIP}); err != nil {
		log.Errorf("update pg_ha.conf failed: %s", err.Error())
		return err
	}

	if isSlave {
		_, err = p.grpcClient.CreateStandbySignal(context.TODO(), &pb.CreateStandbySignalRequest{})
	} else {
		_, err = p.grpcClient.DeleteStandbySignal(context.TODO(), &pb.DeleteStandbySignalRequest{})
	}
	if err != nil {
		log.Errorf("handle standby.signal failed: %s", err.Error())
	}
	return err
}

func (p *PGProxy) runDB() error {
	if err := execCommand("docker", "rm", "-f", p.dbContainerName); err != nil {
		log.Warnf("docker rm failed: %s", err.Error())
	}

	args := []string{"run", "-d", "--rm", "--name", p.dbContainerName, "-e", "POSTGRES_PASSWORD=lx", "-e", "POSTGRES_USER=lx", "-e", "POSTGRES_DB=lx",
		"-p", "5432:5432", "-v", "/home/lx/data:/var/lib/postgresql/data", "-v", "/etc/localtime:/etc/localtime", "postgres"}
	if err := execCommand("docker", args...); err != nil {
		log.Warnf("docker run failed: %s", err.Error())
	}

	time.Sleep(1 * time.Second)
	return nil
}

func execCommand(c string, args ...string) error {
	cmd := exec.Command(c, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *PGProxy) stopDB() error {
	if err := execCommand("docker", "rm", "-f", p.dbContainerName); err != nil {
		log.Warnf("docker rm failed: %s", err.Error())
	}

	return nil
}

func (p *PGProxy) syncData() error {
	db, err := pgxpool.Connect(context.Background(), fmt.Sprintf(PGConnStr, p.dbUser, p.dbPass, p.dbPort, p.dbName))
	if err != nil {
		log.Errorf("connect postges failed: %s", err.Error())
		return err
	}

	if _, err := db.Exec(context.Background(), "select pg_start_backup('replication', true)"); err != nil {
		log.Errorf("set backup failed: %s", err.Error())
		return err
	}

	if err := execCommand("bash", "-c",
		fmt.Sprintf(SyncCommand, p.dbDataDir, p.anotherIP, p.dbDataDir[:strings.LastIndex(p.dbDataDir, "/")+1])); err != nil {
		if matched, _ := regexp.MatchString("exit status 24$", err.Error()); matched != true {
			log.Errorf("exec sync command failed: %s", err.Error())
			return err
		}
	}

	_, err = db.Exec(context.Background(), "select pg_stop_backup()")
	if err != nil {
		log.Errorf("stop backup failed: %s", err.Error())
	}

	return err
}
