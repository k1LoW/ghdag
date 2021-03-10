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
	"fmt"

	"github.com/k1LoW/ghdag/gh"
	"github.com/k1LoW/ghdag/task"
	"github.com/spf13/cobra"
)

// doCommentCmd represents the doComment command
var doCommentCmd = &cobra.Command{
	Use:   "comment [COMMENT]",
	Short: "create the comment of the target issue or pull request",
	Long:  "create the comment of the target issue or pull request.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		r, t, err := initRunnerAndTask(ctx, number)
		if err != nil {
			return err
		}
		sig := fmt.Sprintf("%s%s:%s -->", gh.CommentSigPrefix, "cmd-do", task.ActionTypeDo)
		if err := r.PerformCommentAction(ctx, t, args[0], sig); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	doCommentCmd.Flags().IntVarP(&number, "number", "n", 0, "issue or pull request number")
}
