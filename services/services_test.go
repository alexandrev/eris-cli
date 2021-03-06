package services

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/loaders"
	tests "github.com/eris-ltd/eris-cli/testutils"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var srv *def.ServiceDefinition

var servName string = "ipfs"
var hash string

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	tests.IfExit(testsInit())

	exitCode := m.Run()

	logger.Infoln("Commensing with Tests Tear Down.")
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
	}

	os.Exit(exitCode)
}

func TestKnownService(t *testing.T) {
	do := def.NowDo()
	do.Known = true
	do.Existing = false
	do.Running = false
	do.Operations.Args = []string{"testing"}
	tests.IfExit(util.ListAll(do, "services"))
	k := strings.Split(do.Result, "\n") // tests output formatting.

	if len(k) != 3 {
		tests.IfExit(fmt.Errorf("Did not find exactly 3 service definitions files. Something is wrong.\n"))
	}

	if k[1] != "ipfs" {
		tests.IfExit(fmt.Errorf("Could not find ipfs service definition. Services found =>\t%v\n", k))
	}
}

func TestLoadServiceDefinition(t *testing.T) {
	var e error
	srv, e = loaders.LoadServiceDefinition(servName, true, 1)
	if e != nil {
		logger.Errorln(e)
		tests.IfExit(e)
	}

	if srv.Name != servName {
		logger.Errorf("FAILURE: improper name on LOAD. expected: %s\tgot: %s\n", servName, srv.Name)
	}

	if srv.Service.Name != servName {
		logger.Errorf("FAILURE: improper service name on LOAD. expected: %s\tgot: %s\n", servName, srv.Service.Name)
		tests.IfExit(e)
	}

	if !srv.Service.AutoData {
		logger.Errorf("FAILURE: data_container not properly read on LOAD.\n")
		tests.IfExit(e)
	}

	if srv.Operations.DataContainerName == "" {
		logger.Errorf("FAILURE: data_container_name not set.\n")
		tests.IfExit(e)
	}
}

func TestStartKillService(t *testing.T) {
	testStartService(t, servName, false)
	testKillService(t, servName, true)
}

func TestInspectService(t *testing.T) {
	testStartService(t, servName, false)
	defer testKillService(t, servName, true)

	do := def.NowDo()
	do.Name = servName
	do.Operations.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Inspect service (via tests) =>\t%s:%v:%d\n", servName, do.Operations.Args, do.Operations.ContainerNumber)
	e := InspectService(do)
	if e != nil {
		logger.Infof("Error inspecting service =>\t%v\n", e)
		tests.IfExit(e)
	}

	do = def.NowDo()
	do.Name = servName
	do.Operations.Args = []string{"config.user"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Inspect service (via tests) =>\t%s:%v\n", servName, do.Operations.Args)
	e = InspectService(do)
	if e != nil {
		logger.Infof("Error inspecting service =>\t%v\n", e)
		tests.IfExit(e)
	}
}

func TestLogsService(t *testing.T) {
	testStartService(t, servName, false)
	defer testKillService(t, servName, true)
	do := def.NowDo()
	do.Name = servName
	do.Follow = false
	do.Tail = "5"
	logger.Debugf("Inspect logs (via tests) =>\t%s:%v\n", servName, do.Tail)
	e := LogsService(do)
	if e != nil {
		logger.Errorln(e)
		tests.IfExit(e)
	}
}

func TestExecService(t *testing.T) {
	/*if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have exec privileges (due to their driver). Skipping test.")
		return
	}*/

	testStartService(t, servName, true)
	defer testKillService(t, servName, true)

	do := def.NowDo()
	do.Name = servName
	do.Operations.Interactive = false
	do.Operations.Args = strings.Fields("ls -la /root/")
	logger.Debugf("Exec-ing serv (via tests) =>\t%s:%v\n", servName, strings.Join(do.Operations.Args, " "))
	e := ExecService(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
}

func TestUpdateService(t *testing.T) {
	testStartService(t, servName, false)
	defer testKillService(t, servName, true)
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do := def.NowDo()
	do.Name = servName
	do.Pull = false
	do.Timeout = 1
	logger.Debugf("Update serv (via tests) =>\t%s\n", servName)
	e := UpdateService(do)
	if e != nil {
		logger.Errorln(e)
		tests.IfExit(e)
	}

	testExistAndRun(t, servName, 1, true, true)
	testNumbersExistAndRun(t, servName, 1, 1)
}

func TestKillRmService(t *testing.T) {
	testStartService(t, servName, false)
	do := def.NowDo()
	do.Name = servName
	do.Rm = false
	do.RmD = false
	do.Operations.Args = []string{servName}
	logger.Debugf("Stopping serv (via tests) =>\t%s\n", servName)
	if e := KillService(do); e != nil {
		logger.Errorln(e)
		tests.IfExit(e)
	}

	testExistAndRun(t, servName, 1, true, false)
	testNumbersExistAndRun(t, servName, 1, 0)

	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do = def.NowDo()
	do.Name = servName
	do.Operations.Args = []string{servName}
	do.File = false
	do.RmD = true
	logger.Debugf("Removing serv (via tests) =>\t%s\n", servName)
	if e := RmService(do); e != nil {
		logger.Errorln(e)
		tests.IfExit(e)
	}

	testExistAndRun(t, servName, 1, false, false)
	testNumbersExistAndRun(t, servName, 0, 0)
}

func TestExportService(t *testing.T) {
	testStartService(t, "ipfs", false)
	time.Sleep(5 * time.Second)

	do := def.NowDo()
	do.Name = "ipfs"

	// because IPFS is testy, we retry the put up to
	// 10 times.
	passed := false
	for i := 0; i < 9; i++ {
		if err := ExportService(do); err != nil {
			time.Sleep(2 * time.Second)
			continue
		} else {
			passed = true
			break
		}
	}
	if !passed {
		// final time will throw
		if err := ExportService(do); err != nil {
			tests.IfExit(err)
		}
	}

	hash = do.Result
	testExistAndRun(t, "ipfs", 1, true, true)
}

func TestImportService(t *testing.T) {
	do := def.NowDo()
	do.Name = "eth"
	if hash == "" {
		do.Hash = "QmQ1LZYPNG4wSb9dojRicWCmM4gFLTPKFUhFnMTR3GKuA2"
	} else {
		do.Hash = hash
	}
	logger.Debugf("Import-ing serv (via tests) =>\t%s:%v\n", do.Name, do.Hash)

	// because IPFS is testy, we retry the put up to
	// 10 times.
	passed := false
	for i := 0; i < 9; i++ {
		if err := ImportService(do); err != nil {
			time.Sleep(2 * time.Second)
			continue
		} else {
			passed = true
			break
		}
	}
	if !passed {
		// final time will throw
		if err := ImportService(do); err != nil {
			tests.IfExit(err)
		}
	}

	testKillService(t, "ipfs", true)
	testExistAndRun(t, "ipfs", 1, false, false)
}

func TestNewService(t *testing.T) {
	do := def.NowDo()
	servName := "keys"
	do.Name = servName
	do.Operations.Args = []string{"quay.io/eris/keys"}
	logger.Debugf("New-ing serv (via tests) =>\t%s:%v\n", do.Name, do.Operations.Args)
	e := NewService(do)
	if e != nil {
		logger.Errorln(e)
		tests.IfExit(e)
	}

	do = def.NowDo()
	do.Operations.Args = []string{servName}
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	//do.Operations.ContainerNumber = 1
	logger.Debugf("Starting serv (via tests) =>\t%v:%d\n", do.Operations.Args, do.Operations.ContainerNumber)
	e = StartService(do)
	if e != nil {
		logger.Errorln(e)
		tests.IfExit(e)
	}

	testExistAndRun(t, servName, 1, true, true)
	testNumbersExistAndRun(t, servName, 1, 1)
	testKillService(t, servName, true)
	testExistAndRun(t, servName, 1, false, false)
}

func TestRenameService(t *testing.T) {
	testStartService(t, "keys", false)
	testExistAndRun(t, "keys", 1, true, true)
	testNumbersExistAndRun(t, "keys", 1, 1)

	do := def.NowDo()
	do.Name = "keys"
	do.NewName = "syek"
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	//do.Operations.ContainerNumber = 1
	logger.Debugf("Renaming serv (via tests) =>\t%s:%v\n", do.Name, do.NewName)
	if e := RenameService(do); e != nil {
		tests.IfExit(fmt.Errorf("Error (tests fail) =>\t\t%v\n", e))
	}

	testExistAndRun(t, "syek", 1, true, true)
	testExistAndRun(t, "keys", 1, false, false)
	testNumbersExistAndRun(t, "syek", 1, 1)
	testNumbersExistAndRun(t, "keys", 0, 0)

	do = def.NowDo()
	do.Name = "syek"
	do.NewName = "keys"
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	//do.Operations.ContainerNumber = 1
	logger.Debugf("Renaming serv (via tests) =>\t%s:%v\n", do.Name, do.NewName)
	if e := RenameService(do); e != nil {
		logger.Errorf("Error (tests fail) =>\t\t%v\n", e)
		tests.IfExit(e)
	}

	testExistAndRun(t, "keys", 1, true, true)
	testExistAndRun(t, "syek", 1, false, false)
	testNumbersExistAndRun(t, "keys", 1, 1)
	testNumbersExistAndRun(t, "syek", 0, 0)

	testKillService(t, "keys", true)
	testExistAndRun(t, "keys", 1, false, false)
}

func TestCatService(t *testing.T) {
	do := def.NowDo()
	do.Name = "do_not_use"
	if err := CatService(do); err != nil {
		tests.IfExit(err)
	}

	if do.Result != ini.DefaultIpfs2() {
		tests.IfExit(fmt.Errorf("Cat Service on keys does not match DefaultKeys. Got %s \n Expected %s", do.Result, ini.DefaultIpfs2()))
	}
}

func TestStartKillServiceWithDependencies(t *testing.T) {
	do := def.NowDo()
	do.Operations.Args = []string{"do_not_use"}
	//do.Operations.ContainerNumber = 1
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	logger.Debugf("Starting service with deps =>\t%s:%s\n", servName, "keys")
	if e := StartService(do); e != nil {
		logger.Infof("Error starting service =>\t%v\n", e)
		tests.IfExit(e)
	}

	defer func() {
		testKillService(t, "do_not_use", true)

		testExistAndRun(t, servName, 1, false, false)
		testNumbersExistAndRun(t, servName, 0, 0)

		// XXX: option for kill to kill dependencies too
		testKillService(t, "keys", true)
		//testExistAndRun(t, "keys", 1, false, false)
		//testNumbersExistAndRun(t, "keys", 1, 0)
	}()

	testExistAndRun(t, servName, 1, true, true)
	testExistAndRun(t, "keys", 1, true, true)

	testNumbersExistAndRun(t, "keys", 1, 1)
	testNumbersExistAndRun(t, servName, 1, 1)
}

//----------------------------------------------------------------------
// test utils!

func testStartService(t *testing.T, serviceName string, publishAll bool) {
	do := def.NowDo()
	do.Operations.Args = []string{serviceName}
	do.Operations.ContainerNumber = 1 //util.AutoMagic(0, "service", true)
	do.Operations.PublishAllPorts = publishAll
	logger.Debugf("Starting service (via tests) =>\t%s:%d\n", serviceName, do.Operations.ContainerNumber)
	e := StartService(do)
	if e != nil {
		logger.Infof("Error starting service =>\t%v\n", e)
		tests.IfExit(e)
	}

	testExistAndRun(t, serviceName, 1, true, true)
	testNumbersExistAndRun(t, serviceName, 1, 1)
}

func testKillService(t *testing.T, serviceName string, wipe bool) {
	logger.Debugf("Stopping serv (via tests) =>\t%s\n", servName)

	do := def.NowDo()
	do.Name = serviceName
	do.Operations.Args = []string{serviceName}
	if wipe {
		do.Rm = true
		do.RmD = true
	}
	e := KillService(do)
	if e != nil {
		logger.Errorln(e)
		tests.IfExit(e)
	}
	testExistAndRun(t, serviceName, 1, !wipe, false)
	testNumbersExistAndRun(t, serviceName, 0, 0)
}

func testExistAndRun(t *testing.T, servName string, containerNumber int, toExist, toRun bool) {
	if tests.TestExistAndRun(servName, "services", containerNumber, toExist, toRun) {
		tests.IfExit(nil)
	}
}

func testNumbersExistAndRun(t *testing.T, servName string, containerExist, containerRun int) {
	logger.Infof("\nTesting number of (%s) containers. Existing? (%d) and Running? (%d)\n", servName, containerExist, containerRun)

	logger.Debugf("Checking Existing Containers =>\t%s\n", servName)
	exist := util.HowManyContainersExisting(servName, "service")
	logger.Debugf("Checking Running Containers =>\t%s\n", servName)
	run := util.HowManyContainersRunning(servName, "service")

	if exist != containerExist {
		tests.IfExit(fmt.Errorf("Wrong number of containers existing for service (%s). Expected (%d). Got (%d).\n", servName, containerExist, exist))
	}

	if run != containerRun {
		tests.IfExit(fmt.Errorf("Wrong number of containers running for service (%s). Expected (%d). Got (%d).\n", servName, containerRun, run))
	}

	logger.Infoln("All good.\n")
}

func testsInit() error {
	if err := tests.TestsInit("services"); err != nil {
		return err
	}
	return nil
}
