package data

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

func ImportData(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {

		srv := PretendToBeAService(do.Name, do.Operations.ContainerNumber)
		service, exists := perform.ContainerExists(srv.Operations)

		if !exists {
			return fmt.Errorf("There is no data container for that service.")
		}

		containerName := util.DataContainersName(do.Name, do.Operations.ContainerNumber)
		logger.Debugf("Importing FROM =>\t\t%s\n", do.Source)
		os.Chdir(do.Source)

		logger.Debugf("Importing TO =>\t\t\t%s\n", do.Destination)
		reader, err := util.Tar(do.Source, 0)
		if err != nil {
			return err
		}
		defer reader.Close()

		opts := docker.UploadToContainerOptions{
			InputStream:          reader,
			Path:                 do.Destination,
			NoOverwriteDirNonDir: true,
		}

		logger.Infof("Copying into Cont. ID =>\t%s\n", service.ID)
		logger.Debugf("\tPath =>\t\t\t%s\n", do.Source)
		if err := util.DockerClient.UploadToContainer(service.ID, opts); err != nil {
			return err
		}

		doChown := definitions.NowDo()
		doChown.Operations.DataContainerName = containerName
		doChown.Operations.ContainerType = "data"
		doChown.Operations.ContainerNumber = 1
		doChown.Operations.Args = []string{"chown", "--recursive", "eris", do.Destination}
		_, err = perform.DockerRunData(doChown.Operations, nil)
		if err != nil {
			return fmt.Errorf("Error changing owner: %v\n", err)
		}
	} else {
		ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
		if err := perform.DockerCreateData(ops); err != nil {
			return fmt.Errorf("Error creating data container %v.", err)
		}
		return ImportData(do)
	}
	do.Result = "success"
	return nil
}

func ExecData(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		logger.Infoln("Running exec on container with volumes from data container " + do.Operations.DataContainerName)

		ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
		util.Merge(ops, do.Operations)
		if err := perform.DockerExecData(ops, nil); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("The marmots cannot find that data container.\nPlease check the name of the data container with [eris data ls].")
	}
	do.Result = "success"
	return nil
}

func ExportData(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {

		logger.Infoln("Exporting data container", do.Name)

		// we want to export to a temp directory.
		exportPath, err := ioutil.TempDir(os.TempDir(), do.Name) // TODO: do.Operations.ContainerNumber ?
		defer os.Remove(exportPath)
		if err != nil {
			return err
		}

		srv := PretendToBeAService(do.Name, do.Operations.ContainerNumber)
		service, exists := perform.ContainerExists(srv.Operations)

		if !exists {
			return fmt.Errorf("There is no data container for that service.")
		}

		reader, writer := io.Pipe()
		defer reader.Close()

		opts := docker.DownloadFromContainerOptions{
			OutputStream: writer,
			Path:         do.Source,
		}

		go func() {
			logger.Infof("Copying out of Cont. ID =>\t%s\n", service.ID)
			logger.Debugf("\tPath =>\t\t\t%s\n", do.Source)
			IfExit(util.DockerClient.DownloadFromContainer(service.ID, opts)) // TODO: be smarter about catching this error
			writer.Close()
		}()

		logger.Debugf("Untarring Package from Cont =>\t%s\n", exportPath)
		if err = util.Untar(reader, do.Name, exportPath); err != nil {
			return err
		}

		// now if docker dumps to exportPath/.eris we should remove
		//   move everything from .eris to exportPath
		if err := moveOutOfDirAndRmDir(filepath.Join(exportPath, ".eris"), exportPath); err != nil {
			return err
		}

		// // finally remove everything in the data directory and move
		// //   the temp contents there
		if _, err := os.Stat(do.Destination); os.IsNotExist(err) {
			if e2 := os.MkdirAll(do.Destination, 0755); e2 != nil {
				return fmt.Errorf("Error:\tThe marmots could neither find, nor had access to make the directory: (%s)\n", do.Destination)
			}
		}
		// ClearDir(do.Destionation)
		if err := moveOutOfDirAndRmDir(exportPath, do.Destination); err != nil {
			return err
		}

	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}

	do.Result = "success"
	return nil
}

func moveOutOfDirAndRmDir(src, dest string) error {
	logger.Infof("Move all files/dirs out of a dir and rm -rf that dir.\n")
	logger.Debugf("Source of the move =>\t\t%s.\n", src)
	logger.Debugf("Destin of the move =>\t\t%s.\n", dest)
	toMove, err := filepath.Glob(filepath.Join(src, "*"))
	if err != nil {
		return err
	}

	if len(toMove) == 0 {
		logger.Debugln("No files to move.")
	}

	for _, f := range toMove {
		// logger.Debugf("Moving [%s] to [%s].\n", f, filepath.Join(dest, filepath.Base(f)))

		// using a copy (read+write) strategy to get around swap partitions and other
		//   problems that cause a simple rename strategy to fail. it is more io overhead
		//   to do this, but for now that is preferable to alternative solutions.
		Copy(f, filepath.Join(dest, filepath.Base(f)))
	}

	logger.Infof("Removing directory =>\t\t%s.\n", src)
	err = os.RemoveAll(src)
	if err != nil {
		return err
	}

	return nil
}
