package audio

import (
	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

func NewCmdAudio(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audio",
		Short: "Generate audio",
	}

	cmd.AddCommand(NewCmdAudioSpeech(f, nil))

	return cmd
}
