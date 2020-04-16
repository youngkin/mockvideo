package integrationtests

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

const (
	dockerFailed = iota + 1
	initDBFailed
)

func TestTest(t *testing.T) {
	// Takes a while for the accountd container to start
	time.Sleep(500 * time.Millisecond)

	if err := runCmd("curl -i http://localhost:5000/users"); err != nil {
		t.Errorf("Error curl-ing endpoint: %s", err)
	}
	if err := runCmd("/usr/local/bin/docker logs accountd"); err != nil {
		t.Errorf("Error printing accountd logs: %s", err)
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()

	os.Exit(code)
}

func setup() {
	// Do setup here, like launch docker containers for MySQL and the accountd service
	if err := runCmd("/usr/local/bin/docker run -d --name mysql -p 6603:3306 -e MYSQL_ALLOW_EMPTY_PASSWORD=yes mysql:latest"); err != nil {
		os.Exit(dockerFailed)
	}

	// Takes a while for the MySQL container to start
	var err error
	for i := 0; i < 15; i++ {
		time.Sleep(1 * time.Second)

		if err = initDB(); err != nil {
			fmt.Println("attempt", i, "initializing DB")
			continue
		}
		err = nil
		fmt.Println("Success initializing DB on attempt", i)
		break
	}
	if err != nil {
		os.Exit(initDBFailed)
	}

	// Start accountd service
	if err := runCmd("/usr/local/bin/docker run --name accountd -d -p 5000:5000 -v /Users/rich_youngkin/Software/repos/mockvideo/src/cmd/accountd/testdata:/opt/mockvideo/accountd jenkins/accountd:latest"); err != nil {
		os.Exit(dockerFailed)
	}
}

func teardown() {
	if err := runCmd("/usr/local/bin/docker rm -f mysql accountd"); err != nil {
		os.Exit(dockerFailed)
	}
}

func runCmd(cmdStr string) error {
	cmd := exec.Command("/bin/sh", "-c", cmdStr)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	// run command
	if err := cmd.Run(); err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return nil
}

func initDB() error {
	createTbls := "/Users/rich_youngkin/Software/repos/mockvideo/infrastructure/sql/createTablesDocker.sh"
	popTbls := "/Users/rich_youngkin/Software/repos/mockvideo/infrastructure/sql/createTestDataDocker.sh"

	if err := runCmd(createTbls); err != nil {
		return err
	}

	if err := runCmd(popTbls); err != nil {
		return err
	}

	return nil
}
