package logging

const (
	// For more details about log levels, see https://github.com/go-logr/zapr?tab=readme-ov-file#increasing-verbosity

	// Level0 Generally useful for this to always be visible to a cluster operator
	//  - Programmer errors
	//  - Logging extra info about a panic
	//  - CLI argument handling
	Level0 = iota

	// Level1 A reasonable default log level if you don't want verbosity.
	//  - Information about config (listening on X, watching Y)
	//  - Errors that repeat frequently that relate to conditions that can be corrected (pod detected as unhealthy)
	Level1

	// Level2 Useful steady state information about the service and important log messages that may correlate to significant changes in the system.
	// This is the recommended default log level for most systems.
	//  - Logging HTTP requests and their exit code
	//  - System state changing (killing pod)
	//  - Controller state change events (starting pods)
	//  - Scheduler log messages
	Level2

	// Level3 Extended information about changes
	//  - More info about system state changes
	Level3

	// Level4 Debug level verbosity
	//  - Logging in particularly thorny parts of code where you may want to come back later and check it
	Level4

	// Level5 Trace level verbosity
	//  - Context to understand the steps leading up to errors and warnings
	//  - More information for troubleshooting reported issues
	Level5

	// Helpers
	LevelInfo  = Level2
	LevelDebug = Level4
	LevelTrace = Level5
)
