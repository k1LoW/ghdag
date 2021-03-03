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

	"github.com/Songmu/prompter"
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
		n := args[0]
		path := fmt.Sprintf("%s.yml", n)
		log.Info().Msg(fmt.Sprintf("Creating %s", path))
		if _, err := os.Lstat(path); err == nil {
			return fmt.Errorf("%s already exist", path)
		}
		file, err := os.Create(filepath.Clean(path))
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		ts := `---
# generate by ghdag init
tasks:
  -
    id: set-question-label
    if: 'is_issue && len(labels) == 0 && title endsWith "?"'
    do:
      labels: [question]
    ok:
      run: echo 'Set labels'
    ng:
      run: echo 'failed'
    name: Set 'question' label
`
		tmpl := template.Must(template.New("init").Parse(ts))
		tmplData := map[string]interface{}{}
		if err := tmpl.Execute(file, tmplData); err != nil {
			return err
		}
		yn := prompter.YN("Do you generate a workflow YAML file for GitHub Actions?", true)
		if !yn {
			return nil
		}
		{
			dir := filepath.Join(".github", "workflows")
			if _, err := os.Lstat(dir); err != nil {
				yn := prompter.YN(fmt.Sprintf("%s does not exist. Do you create it?", dir), true)
				if !yn {
					return nil
				}
				if err := os.MkdirAll(dir, 0755); err != nil { // #nosec
					return err
				}
			}
			path := filepath.Join(dir, "ghdag_workflow.yml")
			log.Info().Msg(fmt.Sprintf("Creating %s", path))
			if _, err := os.Lstat(path); err == nil {
				return fmt.Errorf("%s already exist", path)
			}
			file, err := os.Create(filepath.Clean(path))
			if err != nil {
				return err
			}
			defer func() {
				_ = file.Close()
			}()

			ts := `
name: ghdag workflow
on:
  issues:
    types: [opened]
  issue_comment:
    types: [created]
  pull_request:
    types: [opened]

jobs:
  run-workflow:
    name: Run workflow
    runs-on: ubuntu-latest
    container: ghcr.io/k1low/ghdag:latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2
        with:
          token: {{ "${{ secrets.GITHUB_TOKEN }}" }}
      - name: Run ghdag
        run: ghdag run {{ .Name }}.yml
        env:
          GITHUB_TOKEN: {{ "${{ secrets.GITHUB_TOKEN }}" }}
`
			tmpl := template.Must(template.New("workflow").Parse(ts))
			tmplData := map[string]interface{}{
				"Name": n,
			}
			if err := tmpl.Execute(file, tmplData); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
