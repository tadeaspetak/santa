package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/cmd/cmdData"
	"github.com/tadeaspetak/santa/cmd/participants"
	"github.com/tadeaspetak/santa/internal/data"
)

func ask(question string) bool {
	prompt := promptui.Prompt{
		Label:     fmt.Sprint(question),
		IsConfirm: true,
		Default:   "y",
	}
	_, err := prompt.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrAbort) {
			return false
		}

		log.Fatalf("Prompt failed %v\n", err)
	}

	return true

}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize your settings",
	Run: func(cmd *cobra.Command, args []string) {
		dat := (&cmdData.CmdData{}).Load(cmd)

		hasMailgun := dat.Mailgun != data.Mailgun{}
		hasTemplate := dat.Template != data.Template{}
		hasParticipants := len(dat.Participants) != 0

		if hasMailgun && hasTemplate && hasParticipants {
			fmt.Print("\nYou seem to be fully set up, use individual commands to edit the settings.")
			return
		}

		fmt.Print(`\n\n`)

		shouldContinue := ask("Let's get you going in no time ⏳ Are you ready?")
		if !shouldContinue {
			fmt.Print("Doing nothing then.\n")
			return
		}

		if !hasMailgun {
			mailgunCmd.Run(cmd, args)
		} else {
			fmt.Print("\nYour Mailgun config is already present.")
		}
		fmt.Print("\n(You can always edit your mailgun settings using the `mailgun` command.)\n\n")

		if !hasTemplate {
			templateCmd.Run(cmd, args)
		} else {
			fmt.Print("\nYour email template is already configured, let's move on to participants.")
		}
		fmt.Print("\n(You can always edit your template settings using the `template` command.)\n\n")

		if !hasParticipants {
			fmt.Print(
				"\nAdd your participants. Once ready, cmd/ctrl+c out of this and further edit them using the `participants edit` command.\n",
			)

			for {
				participants.AddCmd.Run(cmd, args)
				fmt.Println()
			}
		}
	},
}
