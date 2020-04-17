package integrationtests

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const (
	dockerFailed = iota + 1
	initDBFailed
)

var (
	update           = flag.Bool("update", false, "update .golden files")
	goldenFileDir    = "testdata"
	goldenFileSuffix = ".golden"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	// Do setup here, like launch docker containers for MySQL and the accountd service
	if err := runCmd("docker run -d --name mysql -p 6603:3306 -e MYSQL_ALLOW_EMPTY_PASSWORD=yes mysql:latest"); err != nil {
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
	dCmd := fmt.Sprintf("docker run --name accountd -d -p 5000:5000 -v %s/src/cmd/accountd/testdata:/opt/mockvideo/accountd jenkins/accountd:latest", getBuildDir())
	fmt.Println(dCmd)
	if err := runCmd(dCmd); err != nil {
		os.Exit(dockerFailed)
	}
}

func teardown() {
	if err := runCmd("docker rm -f mysql accountd"); err != nil {
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
	createTbls := fmt.Sprintf("%s/infrastructure/sql/createTablesDocker.sh", getBuildDir())
	popTbls := fmt.Sprintf("%s/infrastructure/sql/createTestDataDocker.sh", getBuildDir())

	if err := runCmd(createTbls); err != nil {
		return err
	}

	if err := runCmd(popTbls); err != nil {
		return err
	}

	return nil
}

func getBuildDir() string {
	bd, found := os.LookupEnv("TRAVIS_BUILD_DIR") // For Travis CI
	if !found {
		bd, found = os.LookupEnv("WORKSPACE") // For Jenkins
	}
	if !found {
		bd = "/Users/rich_youngkin/Software/repos/mockvideo" // Locally
	}

	return bd
}

func updateGoldenFile(t *testing.T, testName string, contents string) {
	gf := filepath.Join(goldenFileDir, testName+goldenFileSuffix)
	t.Log("update golden file")
	if err := ioutil.WriteFile(gf, []byte(contents), 0644); err != nil {
		t.Fatalf("failed to update golden file: %s", err)
	}
}

func readGoldenFile(t *testing.T, testName string) string {
	gf := filepath.Join(goldenFileDir, testName+goldenFileSuffix)
	gfc, err := ioutil.ReadFile(gf)
	if err != nil {
		t.Fatalf("failed reading golden file: %s", err)
	}
	return string(gfc)
}
