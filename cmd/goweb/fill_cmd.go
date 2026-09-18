package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xyzzyapps/goweb/pkg/fill"
)

func registerFillCmd() {
	fillCmd := &cobra.Command{
		Use:   "fill <source.md>",
		Short: "Fill chunks marked ai:open (model or --mock)",
		Long: `Fill rewrites chunk bodies in the Markdown source.
Only <<name>>= … ai:open is a candidate (ai:filled with --refill).
Unmarked chunks are never filled. Tangle never calls a model.

  goweb fill --mock program.md
  goweb fill --model grok-4.5 program.md
  goweb fill --dry-run program.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vars := parseVars(varFlags)
			mock, _ := cmd.Flags().GetBool("mock")
			mockDir, _ := cmd.Flags().GetString("mock-dir")
			model, _ := cmd.Flags().GetString("model")
			name, _ := cmd.Flags().GetString("name")
			refill, _ := cmd.Flags().GetBool("refill")
			dry, _ := cmd.Flags().GetBool("dry-run")
			strict, _ := cmd.Flags().GetBool("strict")
			ctx, _ := cmd.Flags().GetBool("context")
			n, err := fill.Fill(args[0], vars, fill.Options{
				Mock: mock, MockDir: mockDir, Model: model, Name: name,
				Refill: refill, DryRun: dry, Strict: strict, Context: ctx,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "filled %d hole(s)\n", n)
			return nil
		},
	}
	fillCmd.Flags().StringArrayVarP(&varFlags, "var", "v", nil, "Set a variable (key=value)")
	fillCmd.Flags().Bool("mock", false, "Read mock/<chunk>.txt instead of calling a model")
	fillCmd.Flags().String("mock-dir", "", "Mock directory (default: <dir of source>/mock)")
	fillCmd.Flags().String("model", "", "Chat Completions model")
	fillCmd.Flags().String("name", "", "Only this chunk name")
	fillCmd.Flags().Bool("refill", false, "Also refill ai:filled")
	fillCmd.Flags().Bool("dry-run", false, "List holes; do not write")
	fillCmd.Flags().Bool("strict", false, "Fail if the model returns an empty body")
	fillCmd.Flags().Bool("context", false, "Send the whole .md as extra prompt context")

	sealCmd := &cobra.Command{
		Use:     "seal <source.md>",
		Aliases: []string{"accept"},
		Short:   "Mark ai:filled chunks as ai:sealed",
		Long:    `Seal freezes filled holes so later fill passes skip them. Does not auto-run after fill.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vars := parseVars(varFlags)
			name, _ := cmd.Flags().GetString("name")
			emptyOK, _ := cmd.Flags().GetBool("empty-ok")
			dry, _ := cmd.Flags().GetBool("dry-run")
			n, err := fill.Seal(args[0], vars, fill.Options{
				Name: name, EmptyOK: emptyOK, DryRun: dry,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "sealed %d hole(s)\n", n)
			return nil
		},
	}
	sealCmd.Flags().StringArrayVarP(&varFlags, "var", "v", nil, "Set a variable (key=value)")
	sealCmd.Flags().String("name", "", "Only this chunk name")
	sealCmd.Flags().Bool("empty-ok", false, "Allow sealing ai:open stubs")
	sealCmd.Flags().Bool("dry-run", false, "Count without writing")

	rootCmd.AddCommand(fillCmd, sealCmd)
}
