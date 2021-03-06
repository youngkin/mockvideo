// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrationtests

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

const (
	dockerFailed = iota + 1
	buildFailed
	initDBFailed
	svcFailedToStart
)

var (
	update           = flag.Bool("update", false, "update .golden files")
	goldenFileDir    = "testdata"
	goldenFileSuffix = ".golden"
	protocol         = "http"
)

func TestMain(m *testing.M) {
	accountdPID := setup()
	code := m.Run()
	if code != 0 {
		teardown()
		accountdPID.Signal(syscall.SIGTERM) // accountd is run in the background, need to terminate it
	}

	// Rerun tests with protocol set to 'grpc'
	accountdPID.Signal(syscall.SIGTERM)
	protocol = "grpc"
	err := initDB()
	if err != nil {
		os.Exit(initDBFailed)
	}
	accountdPID = startAccountdSvc()
	code = m.Run()

	teardown()
	accountdPID.Signal(syscall.SIGTERM) // accountd is run in the background, need to terminate it
	os.Exit(code)
}

func setup() *os.Process {
	if err := runCmd("cd ..; go build; cd -"); err != nil {
		os.Exit(buildFailed)
	}

	// Launch docker container for MySQL
	if _, found := os.LookupEnv("TRAVIS_BUILD_DIR"); !found { // For Travis CI, Travis will start MySQL
		if err := runCmd("docker run -d --name mysql -p 6603:3306 -e MYSQL_ALLOW_EMPTY_PASSWORD=yes mysql:latest"); err != nil {
			os.Exit(dockerFailed)
		}
	}

	setupDB(15)

	return startAccountdSvc()
}

func startAccountdSvc() *os.Process {
	// Start accountd service
	// Uncomment to run accoutd in docker. If this is uncommented the next 'dCmd := ...' line will have to
	// be commented-out.
	// dCmd := fmt.Sprintf("docker run --name accountd -d -p 5000:5000 -v %s/cmd/accountd/testdata:/opt/mockvideo/accountd local/accountd:latest", getBuildDir())
	dCmd := fmt.Sprintf(`../accountd -configFile ../testdata/config/config -secretsDir ../testdata/secrets -protocol %s`, protocol)
	if _, found := os.LookupEnv("TRAVIS_BUILD_DIR"); found { // For Travis CI need to tweak config path
		dCmd = fmt.Sprintf("%s/accountd -configFile %s/cmd/accountd/testdata/travis/config/config -secretsDir %s/cmd/accountd/testdata/travis/secrets -protocol %s",
			getBuildDir(), getBuildDir(), getBuildDir(), protocol)
	}
	fmt.Printf("\n\n%s\n\n", dCmd)
	// Use 'startCmd()' here so accountd will be started in the background. We need this, the main goroutine,
	// to continue without waiting for accountd to exit (which it won't do until it's sent a SIGTERM)
	p, err := startCmd(dCmd)
	if err != nil {
		os.Exit(svcFailedToStart)
	}
	// Pause while service starts
	time.Sleep(time.Millisecond * 100)

	return p
}

func setupDB(retries int) {
	// Takes a while for the MySQL container to start
	var err error
	for i := 0; i < retries; i++ {
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
}

func teardown() {
	if _, found := os.LookupEnv("TRAVIS_BUILD_DIR"); !found {
		// For Travis CI there is no MySQL container to stop
		runCmd("docker rm -f mysql")
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

// startCmd runs 'cmdStr' in the background
func startCmd(cmdStr string) (*os.Process, error) {
	// The shell in this command 'MUST' be 'bash'. 'sh' apparently creates a different
	// process tree structure where the command run, 'cmdStr' in this case, is in a child
	// process. Signals sent to the 'os.Process' returned from this function get set to
	// the executing shell, and are not propagated to the command run by the shell ('cmdStr').
	// Funny thing is,this behavior only shows up in Travis-CI. Using 'sh' on a Mac works just
	// fine. See https://github.com/travis-ci/travis-ci/issues/8811 for details.
	cmd := exec.Command("/bin/bash", "-c", cmdStr)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	// run command
	if err := cmd.Start(); err != nil {
		fmt.Println("Error:", err)
		return cmd.Process, err
	}
	return cmd.Process, nil
}

func initDB() error {
	var createTbls, popTbls string
	if _, found := os.LookupEnv("TRAVIS_BUILD_DIR"); found { // For Travis CI
		createTbls = fmt.Sprintf("%s/infrastructure/sql/createTablesTravis.sh", getBuildDir())
		popTbls = fmt.Sprintf("%s/infrastructure/sql/createTestDataTravis.sh", getBuildDir())

	} else {
		createTbls = fmt.Sprintf("%s/infrastructure/sql/createTablesDocker.sh", getBuildDir())
		popTbls = fmt.Sprintf("%s/infrastructure/sql/createTestDataDocker.sh", getBuildDir())
	}

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
