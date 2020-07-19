package pg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	pb "github.com/linkingthing/pg-agent/pkg/proto"
	"github.com/zdnscloud/cement/log"

	"github.com/linkingthing/pg-ha/config"
)

const (
	PGConnStr  = "user=%s password=%s host=localhost port=%d database=%s sslmode=disable pool_max_conns=10"
	RetryCount = 60
)

type PGProxy struct {
	grpcClient      pb.PGManagerClient
	dbVolumeName    string
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
		dbVolumeName:    conf.DB.VolumeName,
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

	args := []string{"run", "-d", "--rm", "--name", p.dbContainerName,
		"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", p.dbPass),
		"-e", fmt.Sprintf("POSTGRES_USER=%s", p.dbUser),
		"-e", fmt.Sprintf("POSTGRES_DB=%s", p.dbName),
		"-p", fmt.Sprintf("%d:%d", p.dbPort, p.dbPort),
		"-v", fmt.Sprintf("%s:/var/lib/postgresql/data", p.dbVolumeName),
		"-v", "/etc/localtime:/etc/localtime",
		"postgres:12.2"}
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
	var db *pgxpool.Pool
	var err error
	for i := 1; i <= RetryCount; i++ {
		db, err = pgxpool.Connect(context.Background(), fmt.Sprintf(PGConnStr, p.dbUser, p.dbPass, p.dbPort, p.dbName))
		if err != nil {
			log.Infof("try to connect postges failed: %s and will retry %d", err.Error(), i)
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("connect postges failed: %s", err.Error())
	}

	if _, err := db.Exec(context.Background(), "select pg_start_backup('replication', true)"); err != nil {
		log.Errorf("set backup failed: %s", err.Error())
		return err
	}

	if _, err = p.grpcClient.RsyncPostgresqlData(context.TODO(), &pb.RsyncPostgresqlDataRequest{Address: p.anotherIP}); err != nil {
		log.Errorf("exec sync command failed: %s", err.Error())
		return err
	}

	_, err = db.Exec(context.Background(), "select pg_stop_backup()")
	if err != nil {
		log.Errorf("stop backup failed: %s", err.Error())
	}

	return err
}
