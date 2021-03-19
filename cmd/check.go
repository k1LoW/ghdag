/*
Copyright © 2021 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/ghdag/config"
	"github.com/k1LoW/ghdag/version"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check syntax of workflow file",
	Long:  `Check syntax of workflow file.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().Msg(fmt.Sprintf("%s version %s", version.Name, version.Version))
		b, err := ioutil.ReadFile(filepath.Clean(args[0]))
		if err != nil {
			return errors.WithStack(err)
		}
		c := &config.Config{}

		if err := yaml.Unmarshal(b, c); err != nil {
			return err
		}

		if err := c.CheckSyntax(); err != nil {
			return err
		}

		log.Info().Msg(fmt.Sprintf("the workflow file %s syntax is ok", args[0]))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
