package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

var (
	seedConfig     string
	enableServices string
	force          bool
	hwArch         string
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&enableServices, "enable-services", "y", "", "Comma separated list of services to enable in the config file")
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "If specified, this command will overwrite any existing config file")
	initCmd.Flags().StringVarP(&hwArch, "hw-arch", "x", "x86-64", "Hardware architecture for the platform")
	initCmd.Flags().StringVarP(&seedConfig, "seed-config", "e", "", "The name of a predefined stack to base this new platform on")
	initCmd.Flags().StringP("config-file", "c", "config.yml", "The name of the local config file (defaults to config.yml)")

}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the current directory to be the root for the Modern (Data) Platform by creating an initial config file, if one does not already exists",
	Long: `Initializes the current directory to be the root for a platys platform by creating an initial
config file, if one does not already exists The stack to use as well as its version need to be passed by the --stack and --stack-version options.
By default 'config.yml' is used for the name of the config file, which is created by the init`,
	Run: func(cmd *cobra.Command, args []string) {

		var configFile, err = cmd.Flags().GetString("config-file")
		if err != nil {
			panic(err)
		}
		fmt.Println(fmt.Sprintf("Running using config file [%v]", configFile))

		_, err = os.Stat("./" + configFile)

		if err == nil && !force {
			log.Fatal("config.yml already exists if you want to override it use the [-f] option")
		}

		var ymlConfig yaml.Node
		err = yaml.Unmarshal([]byte(pullConfig()), &ymlConfig)

		if err != nil {
			log.Fatal(err)
		}

		if enableServices != "" { // services where passed as an argument

			services := strings.Split(enableServices, ",") // separate service by coma
			servicesYml := ymlConfig.Content[0].Content
			updatedYml := make([]*yaml.Node, 0)
			copied := 0
			copyNext := false

			for _, n := range servicesYml {

				if copyNext {
					n.Value = "true"
					updatedYml = append(updatedYml, n)
					copied++
					copyNext = false
					continue
				}

				if strings.Contains(n.Value, "platys") { // copy the platys keys
					updatedYml = append(updatedYml, n)
					copied++
					copyNext = true // mark it so as the mapping values are copied during next iteration
					continue
				}

				if strings.Contains(n.Value, "use_timezone") || strings.Contains(n.Value, "private_docker_repository_name") {
					updatedYml = append(updatedYml, n)
					copied++
					copyNext = true
					continue
				}

				if !strings.Contains(n.Value, "_enable") {
					continue
				}

				service := strings.Replace(n.Value, "_enable", "", 1)

				if in_array(service, services) {

					fmt.Println(fmt.Sprintf("Enabling service %v", service))
					updatedYml = append(updatedYml, n)
					copied++
					copyNext = true

				} else {
					fmt.Println(fmt.Sprintf("removing service %v", service))

				}
			}

			updatedYml = updatedYml[:copied]
			ymlConfig.Content[0].Content = updatedYml
		}

		b, _ := yaml.Marshal(&ymlConfig)
		b = addRootIndent(b, 6)

		file, err := os.OpenFile("config.yml", os.O_WRONLY, os.ModeAppend)

		if err != nil {
			log.Fatal(fmt.Sprintf("Unable to open file %v", err))
		}

		defer file.Close()

		for _, s := range strings.SplitN(string(b), "\n", -1) {
			_, err = file.Write([]byte(s + "\n"))
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func addRootIndent(b []byte, n int) []byte {
	prefix := append([]byte("\n"), bytes.Repeat([]byte(" "), n)...)
	b = append(prefix[1:], b...) // Indent first line
	return bytes.ReplaceAll(b, []byte("\n"), prefix)
}

func in_array(val string, list []string) bool {
	for _, b := range list {
		if b == val {
			return true
		}
	}
	return false
}
