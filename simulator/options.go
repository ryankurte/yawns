package sim

// Options defines the command line options available to ons instances
type Options struct {
	ConfigFile string `short:"c" long:"config" description:"Simulation configuration file"`
	BindAddr   string `short:"a" long:"address" description:"Simulator Bind Address (passed to connector)"`
	Connector  string `short:"n" long:"connector" description:"Network connector to use"`
	PCAPFile   string `short:"f" long:"pcap-file" description:"PCap Output File"`
	PCAPStream string `short:"s" long:"pcap-stream" description:"PCap Output Stream"`
	ReportFile string `short:"r" long:"report" description:"Report file to write"`
	LogDir     string `short:"l" long:"log-dir" description:"Log file directory"`
	RunClients string `short:"x" long:"run-clients" description:"Enable simulator client management and specify the command to run (ie. ./app {{server}} {{address}})"`
}

// DefaultOptions creates an options instance populated with the default options
func DefaultOptions() Options {
	return Options{
		ConfigFile: "./example.yml",
		BindAddr:   "ipc:///tmp/ons",
		Connector:  "zmq",
		PCAPFile:   "",
		PCAPStream: "",
	}
}
