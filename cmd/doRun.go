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
	"context"
	"errors"
	"os"
	"strings"

	"github.com/k1LoW/ghdag/runner"
	"github.com/spf13/cobra"
)

// doRunCmd represents the doRun command
var doRunCmd = &cobra.Command{
	Use:   "run [COMMAND...]",
	Short: "execute command using `sh -c`",
	Long:  "execute command using `sh -c`.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("GITHUB_EVENT_NAME") == "" && number == 0 {
			return errors.New("env GITHUB_EVENT_NAME is not set. --number required")
		}
		ctx := context.Background()
		r, err := runner.New(nil)
		if err != nil {
			return err
		}
		if err := r.InitClients(); err != nil {
			return err
		}
		if _, err := r.FetchTarget(ctx, number); err != nil {
			return err
		}
		if err := r.PerformRunAction(ctx, nil, strings.Join(args, " ")); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	doRunCmd.Flags().IntVarP(&number, "number", "n", 0, "issue or pull request number")
}
