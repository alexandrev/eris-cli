package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------

// Primary Data Sub-Command
var Data = &cobra.Command{
	Use:   "data",
	Short: "Manage data containers for your application.",
	Long: `The data subcommand is used to import, and export
data into containers for use by your application.

The [eris data import] and [eris data export] commands should be 
thought of from the point of view of the container.

The [eris data import] command sends files *as is* from 
~/.eris/data/NAME on the host to ~/.eris/ inside 
of the data container.

The [eris data export] command performs this process in the reverse. 
It sucks out whatever is in the volumes of the data container 
and sticks it back into ~/.eris/data/NAME on the host.

At Eris, we use this functionality to formulate little JSONs
and configs on the host and then "stick them back into the
containers"`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// build the data subcommand
func buildDataCommand() {
	Data.AddCommand(dataImport)
	Data.AddCommand(dataList)
	Data.AddCommand(dataRename)
	Data.AddCommand(dataInspect)
	Data.AddCommand(dataExport)
	Data.AddCommand(dataExec)
	Data.AddCommand(dataRm)
	addDataFlags()
}

var dataImport = &cobra.Command{
	Use:   "import NAME",
	Short: "Import ~/.eris/data/name folder to a named data container",
	Long:  `Import ~/.eris/data/name folder to a named data container`,
	Run:   ImportData,
}

var dataList = &cobra.Command{
	Use:   "ls",
	Short: "List the data containers eris manages for you",
	Long:  `List the data containers eris manages for you`,
	Run:   ListKnownData,
}

var dataExec = &cobra.Command{
	Use:   "exec",
	Short: "Run a command or interactive shell in a data container",
	Long: `Run a command or interactive shell in a container with
volumes-from the data container.

Exec can be used to run a single one off command to interact
with the data. Use it for things like ls.

If you want to pass flags into the command that is run in the
data container, please surround the command you want to pass
in with double quotes. Use it like this: "ls -la".

Exec instances run as the Eris user.

Exec can also be used as an interactive shell. When put in
this mode, you can "get inside of" your containers. You will
have root access to a throwaway container which has the volumes
of the data container mounted to it.`,
	Example: `$ eris data exec name ls /home/eris/.eris -- will list the eris dir
$ eris data exec name "ls -la /home/eris/.eris" -- will pass flags to the ls command
$ eris data exec --interactive name -- will start interactive console`,
	Run: ExecData,
}

var dataRename = &cobra.Command{
	Use:   "rename OLD_NAME NEW_NAME",
	Short: "Rename a data container",
	Long:  `Rename a data container`,
	Run:   RenameData,
}

var dataInspect = &cobra.Command{
	Use:   "inspect NAME [KEY]",
	Short: "Show machine readable details.",
	Long:  `Display machine readable details about running containers.`,
	Run:   InspectData,
}

var dataExport = &cobra.Command{
	Use:   "export NAME",
	Short: "Export a named data container's volumes to ~/.eris/data/name",
	Long:  `Export a named data container's volumes to ~/.eris/data/name`,
	Run:   ExportData,
}

var dataRm = &cobra.Command{
	Use:   "rm NAME",
	Short: "Remove a data container",
	Long:  `Remove a data container`,
	Run:   RmData,
}

//----------------------------------------------------

func addDataFlags() {
	dataRm.Flags().BoolVarP(&do.RmHF, "dir", "", false, "remove data folder from host")

	buildFlag(dataRm, do, "rm-volumes", "data")

	buildFlag(dataExec, do, "interactive", "data")

	dataImport.Flags().StringVarP(&do.Destination, "dest", "", "", "destination for import into data container")
	//XXX not used ... but could be if we wanted to be less opiniated
	//dataImport.Flags().StringVarP(&do.Source, "src", "", "", "source on host to import from")
	//dataExport.Flags().StringVarP(&do.Destination, "dest", "", "", "destination for export on host")
	dataExport.Flags().StringVarP(&do.Source, "src", "", "", "source inside data container to export from")
}

//----------------------------------------------------
func ListKnownData(cmd *cobra.Command, args []string) {
	do.Existing = true
	if err := util.ListAll(do, "data"); err != nil {
		return
	}

	// https://www.reddit.com/r/television/comments/2755ow/hbos_silicon_valley_tells_the_most_elaborate/
	datasToManipulate := do.Result
	for _, s := range strings.Split(datasToManipulate, "||") {
		fmt.Println(s)
	}
}

func RenameData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "ge", cmd, args))
	do.Name = args[0]
	do.NewName = args[1]
	IfExit(data.RenameData(do))
}

func InspectData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	if len(args) == 1 {
		do.Operations.Args = []string{"all"}
	} else {
		do.Operations.Args = []string{args[1]}
	}

	IfExit(data.InspectData(do))
}

func RmData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	IfExit(data.RmData(do))
}

func ImportData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	setDefaultDir("import")
	IfExit(data.ImportData(do))
}

func ExportData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	setDefaultDir("export")
	IfExit(data.ExportData(do))
}

func ExecData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]

	// if interactive, we ignore args. if not, run args as command
	if !do.Operations.Interactive {
		if len(args) < 2 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
		args = args[1:]
		if len(args) == 1 {
			args = strings.Split(args[0], " ")
		}
	}

	do.Operations.Args = args
	IfExit(data.ExecData(do))
}

// we don't set this as the default for the flag because it overwrites
// the unified do.Path script with other packages expect to be able to
// provide their own defaults for.
// [zr] perhaps now we can set default in flag...?
func setDefaultDir(typ string) {
	switch typ {
	case "import":
		//if do.Source == "" {
		do.Source = filepath.Join(DataContainersPath, do.Name)
		//}
		if do.Destination == "" {
			do.Destination = ErisContainerRoot
		}
	case "export":
		if do.Source == "" {
			do.Source = ErisContainerRoot
		}
		//if do.Destination == "" {
		do.Destination = filepath.Join(DataContainersPath, do.Name)
		//}
	}
}
