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
	"os"
	"path/filepath"
	"text/template"

	"github.com/k1LoW/ghdag/version"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a workflow file for ghdag",
	Long:  `Generate a workflow file for ghdag.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().Msg(fmt.Sprintf("%s version %s", version.Name, version.Version))
		path := fmt.Sprintf("%s.yml", args[0])
		log.Info().Msg(fmt.Sprintf("Creating %s", path))
		if _, err := os.Lstat(path); err == nil {
			return fmt.Errorf("%s already exist", path)
		}
		file, err := os.Create(filepath.Clean(path))
		if err != nil {
			return err
		}
		defer file.Close()

		ts := `---
tasks:
  -
    id: sample-task
    if: '"good first issue" in labels'
    do:
      comment: 'Good :+1:'
    ok:
      run: echo 'commented'
    ng:
      run: echo 'comment failed'
    desc: Comment when issue labels contains 'good first issue'
`
		tmpl := template.Must(template.New("init").Parse(ts))
		tmplData := map[string]interface{}{}
		if err := tmpl.Execute(file, tmplData); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
