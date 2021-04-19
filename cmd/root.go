package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zlei1/engines"
)

var (
	sl      string
	tl      string
	engine  string
	rootCmd = &cobra.Command{
		Use:   "go-translator {word}",
		Short: "translate words",
		Long:  `translate words to other language by cmdline`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			words := strings.Join(args, " ")
			switch {
			case engine == "youdao":
				yd := engines.NewYoudaoTranslator(sl, tl, words)
				err := yd.Perform()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			case engine == "baidu":
				if tl == "auto" {
					tl = "zh"
				}

				bd := engines.NewBaiduTranslator(sl, tl, words)
				err := bd.Perform()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			default:
			}
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&sl, "source-language", "s", "auto", "来源语言")
	rootCmd.Flags().StringVarP(&tl, "target-language", "t", "auto", "目标语言")
	rootCmd.Flags().StringVarP(&engine, "engine", "e", "youdao", "翻译引擎")
}
