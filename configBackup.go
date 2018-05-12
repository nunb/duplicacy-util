// Copyright © 2018 Jeff Coffler <jeff@taltos.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

var Filename string

type ConfigFile struct {
	// Name (without extension) of the configuration file
	configFile string

	// Directory for repository
	repoDir string

	// Storage information for backup, copy, prune, and check commands respectively
	backupInfo []map[string]string
	copyInfo   []map[string]string
	pruneInfo  []map[string]string
	checkInfo  []map[string]string
}

func NewConfigFile() *ConfigFile {
	config := new(ConfigFile)
	return config
}

func (config *ConfigFile) SetConfig(cnfFile string) {
	config.configFile = cnfFile
}

func (config *ConfigFile) LoadConfig(verboseFlag bool, debugFlag bool) error {
	var err error = nil

	// Separate config file should use new viper instance
	v := viper.New()

	// Search config in home directory with name ".duplicacy-util" (without extension).
	v.AddConfigPath("$HOME/.duplicacy-util")
	v.SetConfigName(config.configFile)

	v.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		return err
	}

	// Grab the repository location
	config.repoDir = v.GetString("repository")
	if config.repoDir == "" {
		err = errors.New("missing mandatory repository location")
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
	if _, err = os.Stat(config.repoDir); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
	}

	// Populate backup information from storages
	if v.IsSet("storage") {
		for i := 1; ; i++ {
			var storageMap map[string]string = make(map[string]string)

			key := "storage." + strconv.Itoa(i)
			if v.IsSet(key) {
				if v.IsSet(key + ".name") {
					storageMap["name"] = v.GetString(key + ".name")
				} else {
					err = errors.New("missing mandatory storage field: " + key + ".name")
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
				// Default to -threads:1 if not otherwise specified
				threads := v.GetInt(key + ".threads")
				if threads != 0 {
					storageMap["threads"] = strconv.Itoa(threads)
				} else {
					storageMap["threads"] = "1"
				}
				config.backupInfo = append(config.backupInfo, storageMap)
			} else {
				break
			}
		}

		if len(config.backupInfo) == 0 {
			err = errors.New("no storage locations defined in configuration")
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
	} else {
		err = errors.New("no storage locations defined in configuration")
		fmt.Fprintln(os.Stderr, "Error:", err)
	}

	// Populate copy information
	if v.IsSet("copy") {
		for i := 1; ; i++ {
			var copyMap map[string]string = make(map[string]string)

			key := "copy." + strconv.Itoa(i)
			if v.IsSet(key) {
				if v.IsSet(key+".from") && v.IsSet(key+".to") {
					copyMap["from"] = v.GetString(key + ".from")
					copyMap["to"] = v.GetString(key + ".to")
				} else {
					err = errors.New("missing mandatory storage field: " + key + ".from or " + key + ".to")
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
				// Default to -threads:1 if not otherwise specified
				threads := v.GetInt(key + ".threads")
				if threads != 0 {
					copyMap["threads"] = strconv.Itoa(threads)
				} else {
					copyMap["threads"] = "1"
				}
				config.copyInfo = append(config.copyInfo, copyMap)
			} else {
				break
			}
		}

		if len(config.copyInfo) == 0 {
			err = errors.New("no copy locations defined in configuration")
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
	}

	// Populate prune information
	if v.IsSet("prune") {
		for i := 1; ; i++ {
			var pruneMap map[string]string = make(map[string]string)

			key := "prune." + strconv.Itoa(i)
			if v.IsSet(key) {
				if v.IsSet(key + ".storage") {
					pruneMap["storage"] = v.GetString(key + ".storage")
				} else {
					err = errors.New("Missing mandatory storage field: " + key + ".storage")
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
				if v.IsSet(key + ".keep") {
					// Split/join to get "-keep " before each element
					splitList := strings.Split(v.GetString(key+".keep"), " ")
					for i, element := range splitList {
						splitList[i] = "-keep " + element
					}

					pruneMap["keep"] = strings.Join(splitList, " ")
				} else {
					err = errors.New("Missing mandatory storage field: " + key + ".keep")
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
				config.pruneInfo = append(config.pruneInfo, pruneMap)
			} else {
				break
			}
		}

		if len(config.pruneInfo) == 0 {
			err = errors.New("no prune locations defined in configuration")
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
	} else {
		err = errors.New("no prune locations defined in configuration")
		fmt.Fprintln(os.Stderr, "Error:", err)
	}

	// Populate check information
	if v.IsSet("check") {
		for i := 1; ; i++ {
			var checkMap map[string]string = make(map[string]string)

			key := "check." + strconv.Itoa(i)
			if v.IsSet(key) {
				if v.IsSet(key + ".storage") {
					checkMap["storage"] = v.GetString(key + ".storage")
				} else {
					err = errors.New("missing mandatory storage field: " + key + ".storage")
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
				// See if all is specified
				allFlag := v.GetBool(key + ".all")
				if allFlag {
					checkMap["all"] = "true"
				} else {
					checkMap["all"] = "false"
				}
				config.checkInfo = append(config.checkInfo, checkMap)
			} else {
				break
			}
		}

		if len(config.checkInfo) == 0 {
			err = errors.New("no check locations defined in configuration")
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
	} else {
		err = errors.New("no check locations defined in configuration")
		fmt.Fprintln(os.Stderr, "Error:", err)
	}

	// Generate verbose/debug output if requested (assuming no fatal errors)

	if err == nil {
		fmt.Println("Using config file:  ", v.ConfigFileUsed())

		if verboseFlag {
			fmt.Println()
			fmt.Println("Backup Information:")
			fmt.Printf("  Num\t%-20s%s\n", "Storage", "Threads")
			for i := range config.backupInfo {
				fmt.Printf("  %2d\t%-20s   %-2s\n", i+1, config.backupInfo[i]["name"], config.backupInfo[i]["threads"])
			}
			if len(config.copyInfo) != 0 {
				fmt.Println("Copy Information:")
				fmt.Printf("  Num\t%-20s%-20s%s\n", "From", "To", "Threads")
				for i := range config.copyInfo {
					fmt.Printf("  %2d\t%-20s%-20s   %-2s\n", i+1, config.copyInfo[i]["from"], config.copyInfo[i]["to"], config.copyInfo[i]["threads"])
				}
			}
			fmt.Println()

			fmt.Println("Prune Information:")
			for i := range config.pruneInfo {
				fmt.Printf("  %2d: Storage %s\n      Keep: %s\n", i+1, config.pruneInfo[i]["storage"], config.pruneInfo[i]["keep"])
			}
			fmt.Println()

			fmt.Println("Check Information:")
			fmt.Printf("  Num\t%-20s%s\n", "Storage", "All Snapshots")
			for i := range config.checkInfo {
				fmt.Printf("  %2d\t%-20s    %-2s\n", i+1, config.checkInfo[i]["storage"], config.checkInfo[i]["all"])
			}
			fmt.Println()
		}

		if debugFlag {
			fmt.Println("\nBackup Info:", config.backupInfo)
			fmt.Println("Copy Info:", config.copyInfo)
			fmt.Println("Prune Info:", config.pruneInfo)
			fmt.Println("Check Info", config.checkInfo)
		}
	}

	return err
}
