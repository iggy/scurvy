// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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

package cmd

import (
	"fmt"

	"scurvy/config"
	"scurvy/msgs"

	"github.com/spf13/cobra"
)

// testreportfilesCmd represents the testreportfiles command
var testreportfilesCmd = &cobra.Command{
	Use:   "testreportfiles",
	Short: "Send report files test message",
	Long:  `Sends a test message that mimics the message received from syncd.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("testreportfiles called")
		config.ReadConfig()
		msg := msgs.ReportFiles{Full: true, Changed: false}

		msgs.SendNatsMsg(msgs.ReportFilesSubject, msg)
	},
}

func init() {
	rootCmd.AddCommand(testreportfilesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testreportfilesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testreportfilesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
