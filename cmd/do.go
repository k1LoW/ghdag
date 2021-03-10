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

	"github.com/k1LoW/ghdag/runner"
	"github.com/k1LoW/ghdag/target"
	"github.com/spf13/cobra"
)

var number int

// doCmd represents the do command
var doCmd = &cobra.Command{
	Use:   "do",
	Short: "do ghdag action as oneshot command",
	Long:  `do ghdag action as oneshot command.`,
}

func init() {
	rootCmd.AddCommand(doCmd)
	doCmd.AddCommand(doRunCmd)
	doCmd.AddCommand(doLabelsCmd)
	doCmd.AddCommand(doAssigneesCmd)
	doCmd.AddCommand(doReviewersCmd)
	doCmd.AddCommand(doCommentCmd)
	doCmd.AddCommand(doStateCmd)
	doCmd.AddCommand(doNotifyCmd)
}

func initRunnerAndTask(ctx context.Context, number int) (*runner.Runner, *target.Target, error) {
	if os.Getenv("GITHUB_EVENT_NAME") == "" && number == 0 {
		return nil, nil, errors.New("env GITHUB_EVENT_NAME is not set. --number required")
	}
	r, err := runner.New(nil)
	if err != nil {
		return nil, nil, err
	}
	if err := r.InitClients(); err != nil {
		return nil, nil, err
	}
	t, err := r.FetchTarget(ctx, number)
	if err != nil {
		return nil, nil, err
	}
	return r, t, nil
}
