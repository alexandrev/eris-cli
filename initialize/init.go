package initialize

import (
	"fmt"
	"os"
	//	"path"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func Initialize(do *definitions.Do) error {

	log.Warn("Checking for Eris Root Directory")
	newDir, err := checkThenInitErisRoot()
	if err != nil {
		return err
	}

	if !newDir { //new ErisRoot won't have either...can skip
		log.Warn("Checking if migration is required")
		if err := checkIfMigrationRequired(do.Yes); err != nil {
			return err
		}

		if err := checkIfCanOverwrite(); err != nil {
			return err
		}
	} else {
		do.Yes = true //no need to prompt if fresh install
	}

	if do.Pull { //true by default; if imgs already exist, will check for latest anyways
		if err := GetTheImages(); err != nil {
			return err
		}
	}

	//drops: services, actions, & chain defaults from toadserver
	log.Warn("Initializing defaults")
	if err := InitDefaults(do, newDir); err != nil {
		return fmt.Errorf("Error:\tcould not Instantiate default services.\n%s\n", err)
	}

	//TODO: when called from cli provide option to go on tour, like `ipfs tour`
	//[zr] this'll be cleaner with `make`
	log.Warn("\nThe marmots have everything set up for you.\nIf you are just getting started please type [eris] to get an overview of the tool.\n")

	return nil
}

func InitDefaults(do *definitions.Do, newDir bool) error {
	//do.Yes skips the ask & was set to true if newDir = true or by flag
	//TODO fix the ask to pull (or override with pull approve & test properly

	var srvPath string
	var actPath string
	var chnPath string

	/*if do.Quiet {
		srvPath = "/tmp/eris/services"
		actPath = "/tmp/eris/actions"
		chnPath = "/tmp/eris/chains"
	} else {*/
	srvPath = common.ServicesPath
	actPath = common.ActionsPath
	chnPath = common.ChainsPath
	//	}

	if askToPull(do.Yes, "services") {
		if err := dropServiceDefaults(srvPath, do.Source); err != nil {
			return err
		}
	}

	if askToPull(do.Yes, "actions") {
		if err := dropActionDefaults(actPath, do.Source); err != nil {
			return err
		}
	}

	if askToPull(do.Yes, "chains") {
		if err := dropChainDefaults(chnPath, do.Source); err != nil {
			return err
		}
	}

	log.WithField("root", common.ErisRoot).Warn("Initialized eris root directory with default service, action, and chain files.")

	return nil
}

func checkThenInitErisRoot() (bool, error) {
	var newDir bool
	if !util.DoesDirExist(common.ErisRoot) {
		log.Warn("Eris Root Directory does not exist. The marmots will initialize this directory for you.")
		if err := common.InitErisDir(); err != nil {
			return true, fmt.Errorf("Error:\tcould not Initialize the Eris Root Directory.\n%s\n", err)
		}
		newDir = true
	} else { // ErisRoot exists
		log.Warn("Eris Root Directory already exists. Backup up important files in (...) or decline the overwrite.")
		newDir = false
	}
	return newDir, nil
}

func checkIfMigrationRequired(doYes bool) error {
	var prompt bool
	if doYes || os.Getenv("ERIS_MIGRATE_APPROVE") == "true" {
		prompt = false
	} else {
		prompt = true
	}
	if err := util.MigrateDeprecatedDirs(common.DirsToMigrate, prompt); err != nil {
		return fmt.Errorf("Error:\tcould not migrate directories.\n%s\n", err)
	}
	return nil
}

func checkIfCanOverwrite() error {
	var input string
	log.WithField("root", common.ErisRoot).Warn("Eris Root Directory already exists.\nContinuing may overwrite files in:\n%s\n%s\nDo you wish to continue? (y/n): ") // common.ChainsPath ??
	if _, err := fmt.Scanln(&input); err != nil {
		return fmt.Errorf("Error reading from stdin: %v\n", err)
	}
	if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
		log.Debug("Confirmation verified. Proceeding.")
	} else {
		log.Warn("The marmots will not proceed without your permission to overwrite.\nPlease backup your files and try again.")
		return fmt.Errorf("Error:\tno permission given to overwrite services and actions.\n")
	}
	return nil
}

func GetTheImages() error {
	if os.Getenv("ERIS_PULL_APPROVE") == "true" {
		if err := pullDefaultImages(); err != nil {
			return err
		}
	} else {
		var input string
		//there's gotta be a better way (logrus?)
		log.Warn("WARNING: Approximately 5 gigabytes of docker images are about to be pulled onto your host machine.")
		log.Warn("Please ensure that you have sufficient bandwidth to handle the download.")
		log.Warn("On a remote host in the cloud, this should only take a few minutes but can sometimes take 10 or more...")
		log.Warn("These times can double or triple on local host machines.")
		log.Warn("If you already have these images, they will be updated") //[zr] test that
		log.Warn("To avoid this warning on all future pulls, set ERIS_PULL_APPROVE=true as an environment variable")
		log.Warn("Confirm pull: (y/n)")

		fmt.Scanln(&input)
		if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
			log.Warn("Pulling default docker images from quay.io")
			if err := pullDefaultImages(); err != nil {
				return err
			}
		}
	}
	log.Warn("Pulling of default images successful")
	return nil
}

func askToPull(skip bool, location string) bool {
	if skip || os.Getenv("ERIS_PULL_APPROVE") == "true" {
		return true
	}
	var input string
	//TODO be more specific about what's in the dir
	log.WithField("eris", location).Warn("Looks like the eris directory exists.\nWould you like the marmots to pull in any recent changes? (y/n): ")
	fmt.Scanln(&input)

	if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
		return true
	}
	return false
}
