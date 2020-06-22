# system-conf

A flexible configuration system for Go inspired by the unit file and drop-in system from systemd.

system-conf is a flexible configuration system that puts the user in control of how their configuration
is structured. It's based on the unit system used in systemd. It uses an INI style configuration and supports
drop-in files and templating. 

## Quick Start

```go
package main

import "github.com/ppacher/system-conf/conf"

var triggerSpec = []conf.OptionSpec {
    {
        Name: "Trigger",
        Type: conf.StringSliceType,
        Required: true,
        Description: "Event to trigger",
    },
    {
        Name: "Match",
        Type: conf.StringType,
        Description: "Only trigger if value matches, optional",
    },
}

var actionSpec = []conf.OptionSpec {
    {
        Name: "Exec",
        Required: true,
        Type: conf.StringType,
        Description: "Command to execute",
    },
    {
        Name: "WorkDir",
        Default: ".",
        Type: conf.StringType,
        Description: "Optional working directory for the command",
    },
}

var sectionSpecs = map[string][]conf.OptionSpec {
    "Trigger": triggerSpec,
    "Action": actionSpec,
}

func checkErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    // Read all .hook files from ./config/
    files, err := conf.ReadDir("./config", "hook", sectionSpecs) 
    checkErr(err)

    for _, file := range files {
        fmt.Println(file.Path)
        
        trigger := file.Get("Trigger")
        if trigger == nil {
            log.Fatalf("Missing [Trigger] section")
        }
        
        actions := file.GetAll("Actions")
        if len(actions) == 0 {
            log.Fatalf("At least one [Action] section must be defined")
        }
        
        events := trigger.GetStringSlice("Trigger")
        
        // ...
    }
}

```