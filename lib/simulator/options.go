package sim

// Options defines the command line options available to ons instances
type Options struct {
	ConfigFile string `short:"c" long:"config" description:"Simulation configuration file"`
	BindAddr   string `short:"a" long:"address" description:"Simulator Bind Address"`

	PCAPFile   string `short:"f" long:"pcap-file" description:"PCap Output File"`
	PCAPStream string `short:"s" long:"pcap-stream" description:"PCap Output Stream"`

	ReportFile string `short:"r" long:"report" description:"Report file to write"`
	LogDir     string `short:"l" long:"log-dir" description:"Log file directory"`

	ClientAddr string `short:"b" long:"client-address" description:"Client bind address for autorun clients"`
}

// DefaultOptions creates an options instance populated with the default options
func DefaultOptions() Options {
	return Options{
		ConfigFile: "./example.yml",
		BindAddr:   "tcp://*:10109",
		PCAPFile:   "",
		PCAPStream: "",
		ClientAddr: "tcp://localhost:10109",
	}
}
